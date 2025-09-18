from langchain.prompts import HumanMessagePromptTemplate, ChatPromptTemplate
from langchain.output_parsers import PydanticOutputParser
from langchain.agents import create_tool_calling_agent, AgentExecutor
from prompts.nextjs_prompt import NEXTJS_PROMPT
from prompts.query_prompt import QUERY_PROMPT
from services.models.ai_models import QueryResult, CodeAgentResult, CodeAgentData
from services.sandbox_service import SandboxService
from clients.openai_client import OpenAIClient
from services.agent_callback_service import CodeAgentCallBack
from loguru import logger
'''
OpenAI service (sends requests to OpenAI llm). 
Can handle coding with tools and will execute them in the sandbox.
Can handle general queries as well.
'''

class OpenAICodeAgentService:
    def __init__(self, openai_client: OpenAIClient, sandbox_service:SandboxService):
        try:
            self.llm = openai_client.get_client()
            code_agent_tools = sandbox_service.get_tools()
            self.parser = PydanticOutputParser(pydantic_object=CodeAgentResult)
            prompt = ChatPromptTemplate.from_template(
            template=f"""
            {NEXTJS_PROMPT}
            {{format_instructions}}
            {{input}}
            {{agent_scratchpad}}
            """
            )
            prompt_formatted = prompt.partial(format_instructions=self.parser.get_format_instructions())
            code_agent = create_tool_calling_agent(
                    llm=self.llm,
                    prompt=prompt_formatted,
                    tools=code_agent_tools 
                )
            self.agent = AgentExecutor(agent=code_agent, tools=code_agent_tools, verbose=True)
        except Exception as e:
            logger.error(f"Failed to initialize OpenAICodeAgentService: {str(e)}")
            raise

    async def process_code_request(self, sandbox_id:str, user_message:str):
        try: 
            contextual_input:str = f"Sandbox ID: {sandbox_id}\nTask: {user_message}"
            callback = CodeAgentCallBack()
            result = await self.agent.ainvoke({"input": contextual_input}, config={"callbacks":[callback]})
            logger.info(result)
            parsed_result = self.parser.parse(result["output"])
            agent_actions = callback.get_result()

            return CodeAgentData(
                summary=parsed_result.summary,
                commands=agent_actions.commands_executed,
                files=agent_actions.updated_files
            ) 
        except Exception as e:
            raise Exception(f"openai code agent failed to generate response. error: {str(e)}")


class OpenAIService:
    def __init__(self, openai_client: OpenAIClient):
        self.llm = openai_client.get_client()

    async def process_query_request(self, user_message:str) -> str:
        try:
            parser = PydanticOutputParser(pydantic_object=QueryResult)
            message = HumanMessagePromptTemplate.from_template(template=QUERY_PROMPT)
            chat_prompt = ChatPromptTemplate.from_messages(messages=[message])

            chat_prompt_with_values = chat_prompt.format_prompt(user_query=user_message, format_instructions=parser.get_format_instructions())
            
            response = self.llm(chat_prompt_with_values.to_messages())
            
            data = parser.parse(response.content)
            return data.response
        except Exception as e:
            raise Exception(f"openai llm failed to generate response. error: {str(e)}")

    
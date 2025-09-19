from langchain.prompts import HumanMessagePromptTemplate, ChatPromptTemplate
from langchain.output_parsers import PydanticOutputParser
from langchain.agents import create_tool_calling_agent, AgentExecutor
from prompts.nextjs_prompt import NEXTJS_PROMPT
from prompts.query_prompt import QUERY_PROMPT
from services.models.ai_models import AIClient, QueryResult, CodeAgentResult, CodeAgentData
from services.sandbox_service import SandboxService
from services.agent_callback_service import CodeAgentCallBack
from loguru import logger
'''
Ai Service service (sends requests to llm). 
Can handle coding with tools and will execute them in the sandbox.
Can handle general queries as well.
'''

class CodeAgentService:
    def __init__(self, llm: AIClient, sandbox_service:SandboxService):
        try:
            self.llm = llm.get_client()
            code_agent_tools = sandbox_service.get_tools()
            self.parser = PydanticOutputParser(pydantic_object=CodeAgentResult)
            prompt = ChatPromptTemplate.from_template(
            template=f"""
            {NEXTJS_PROMPT}
            CRITICAL: Your response MUST be valid JSON matching this exact format:
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
            self.agent = AgentExecutor(agent=code_agent, tools=code_agent_tools, verbose=False)
        except Exception as e:
            logger.error(f"failed to initialize AICodeAgentService. error: {str(e)}")
            raise

    async def process_code_request(self, sandbox_id:str, user_message:str):
        try: 
            contextual_input:str = f"Sandbox ID: {sandbox_id}\nTask: {user_message}"
            callback = CodeAgentCallBack()
            result = await self.agent.ainvoke({"input": contextual_input}, config={"callbacks":[callback]})

            # Validate output before parsing
            output = result.get("output", "")
            if not output:
                logger.error("LLM returned empty output")
                # Return with callback data but generic summary
                agent_actions = callback.get_result()
                code_result = CodeAgentData(
                    summary="Task completed successfully",
                    commands=agent_actions.commands_executed,
                    files=agent_actions.updated_files
                )
            else: 
                parsed_result = self.parser.parse(result["output"])
                agent_actions = callback.get_result()
                code_result = CodeAgentData(
                    summary=parsed_result.summary,
                    commands=agent_actions.commands_executed,
                    files=agent_actions.updated_files
                ) 

            return code_result

        except Exception as e:
            raise Exception(f"code agent failed to generate response. error: {str(e)}")

class GeneralAIService:
    def __init__(self, llm: AIClient):
        self.llm = llm.get_client()

    async def process_query_request(self, user_message:str) -> str:
        try:
            parser = PydanticOutputParser(pydantic_object=QueryResult)
            message = HumanMessagePromptTemplate.from_template(template=QUERY_PROMPT)
            chat_prompt = ChatPromptTemplate.from_messages(messages=[message])

            chat_prompt_with_values = chat_prompt.format_prompt(user_message=user_message, format_instructions=parser.get_format_instructions())
            
            response = self.llm(chat_prompt_with_values.to_messages())
            
            content:str = str(response.content)

            data = parser.parse(content)
            return data.response
        except Exception as e:
            raise Exception(f"llm failed to generate response. error: {str(e)}")

    
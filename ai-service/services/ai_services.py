from langchain.prompts import HumanMessagePromptTemplate, ChatPromptTemplate
from langchain.output_parsers import PydanticOutputParser
from langchain.agents import create_tool_calling_agent, AgentExecutor
from prompts.nextjs_prompt import NEXTJS_PROMPT
from prompts.query_prompt import QUERY_PROMPT
from services.models.ai_models import AIClient, QueryResult, CodeAgentResult, CodeAgentData
from services.sandbox_service import SandboxService
from services.agent_callback_service import CodeAgentCallBack
from utils.logging import logger

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
            logger.error("code_agent_service_initialization_failed", 
                error_type=type(e).__name__,
                error=str(e),
                exc_info=True)
            raise

    async def process_code_request(self, sandbox_id:str, user_message:str):
        try: 
            contextual_input:str = f"Sandbox ID: {sandbox_id}\nTask: {user_message}"
            callback = CodeAgentCallBack()
            
            logger.debug("calling_llm_agent")

            result = await self.agent.ainvoke({"input": contextual_input}, config={"callbacks":[callback]})
            
            # Validate output before parsing
            # TODO Google gemeni fails to parse correctly need to come up with a consistent output method
            output = result.get("output", "")
            if not output:
                logger.warning("llm_returned_empty_output", 
                              using_fallback_summary=True)

                agent_actions = callback.get_result()
                code_result = CodeAgentData(
                    summary="Task completed successfully",
                    commands=agent_actions.commands_executed,
                    files=agent_actions.updated_files
                )
            else: 
                logger.debug("parsing_llm_output")
                parsed_result = self.parser.parse(result["output"])
                agent_actions = callback.get_result()
                code_result = CodeAgentData(
                    summary=parsed_result.summary,
                    commands=agent_actions.commands_executed,
                    files=agent_actions.updated_files
                ) 

            return code_result

        except Exception as e:
            logger.error("code_agent_processing_failed", 
                    message_length=len(user_message),
                    error_type=type(e).__name__,
                    error=str(e),
                    exc_info=True)
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
            
            response = await self.llm.ainvoke(chat_prompt_with_values.to_messages())
            
            content:str = str(response.content)

            data = parser.parse(content)
            return data.response
        except Exception as e:
            logger.error("general_query_processing_failed", 
                        message_length=len(user_message),
                        error_type=type(e).__name__,
                        error=str(e),
                        exc_info=True)
            raise Exception(f"llm failed to generate response. error: {str(e)}")

    
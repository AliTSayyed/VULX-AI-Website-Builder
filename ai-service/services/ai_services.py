from langchain.prompts import HumanMessagePromptTemplate, ChatPromptTemplate
from langchain.output_parsers import PydanticOutputParser
from langchain.agents import create_tool_calling_agent, AgentExecutor
from prompts.nextjs_prompt import NEXTJS_PROMPT
from prompts.query_prompt import QUERY_PROMPT
from services.models.ai_models import (
    AIClient,
    LLMQueryResult,
    CodeAgentResult,
    CodeAgentData,
)
from services.sandbox_service import SandboxService
from services.agent_callback_service import CodeAgentCallBack
from utils.logging import logger

"""
Ai Service service (sends requests to llm). 
Can handle coding with tools and will execute them in the sandbox.
Can handle general queries as well.
"""


class CodeAgentService:
    def __init__(self, llm: AIClient, sandbox_service: SandboxService) -> None:
        try:
            self.llm = llm.get_client()
            code_agent_tools = sandbox_service.get_tools()
            self.parser = PydanticOutputParser(pydantic_object=CodeAgentResult)
            prompt = ChatPromptTemplate.from_template(template=NEXTJS_PROMPT)
            prompt_formatted = prompt.partial(
                format_instructions=self.parser.get_format_instructions()
            )
            code_agent = create_tool_calling_agent(
                llm=self.llm, prompt=prompt_formatted, tools=code_agent_tools
            )
            self.agent = AgentExecutor(
                agent=code_agent, tools=code_agent_tools, verbose=False
            )

        except Exception as e:
            logger.error(
                "code_agent_service_initialization_failed",
                error_type=type(e).__name__,
                error=str(e),
                exc_info=True,
            )
            raise

    async def process_code_request(
        self, sandbox_id: str, user_message: str
    ) -> CodeAgentData:
        try:
            contextual_input = f"Sandbox ID: {sandbox_id}\nTask: {user_message}"
            callback = CodeAgentCallBack()

            logger.debug("calling_llm_agent")

            result = await self.agent.ainvoke(
                {"input": contextual_input}, config={"callbacks": [callback]}
            )

            # Validate output before parsing
            output = result.get("output", "")
            if not output:
                logger.warning("llm_returned_empty_output", using_fallback_summary=True)

                agent_actions = callback.get_result()
                code_result = CodeAgentData(
                    summary="Task completed successfully",
                    commands=agent_actions.commands_executed,
                    files=agent_actions.updated_files,
                )
            else:
                logger.debug("parsing_llm_output")
                parsed_result: CodeAgentResult = self.parser.parse(result["output"])
                agent_actions = callback.get_result()
                code_result = CodeAgentData(
                    summary=parsed_result.summary,
                    commands=agent_actions.commands_executed,
                    files=agent_actions.updated_files,
                )

            return code_result

        except Exception as e:
            logger.error(
                "code_agent_processing_failed",
                message_length=len(user_message),
                error_type=type(e).__name__,
                error=str(e),
                exc_info=True,
            )
            raise


class GeneralAIService:
    def __init__(self, llm: AIClient) -> None:
        self.llm = llm.get_client()

    async def process_query_request(self, user_message: str) -> str:
        try:
            parser = PydanticOutputParser(pydantic_object=LLMQueryResult)
            message = HumanMessagePromptTemplate.from_template(template=QUERY_PROMPT)
            chat_prompt = ChatPromptTemplate.from_messages(messages=[message])

            chat_prompt_with_values = chat_prompt.format_prompt(
                user_message=user_message,
                format_instructions=parser.get_format_instructions(),
            )

            response = await self.llm.ainvoke(chat_prompt_with_values.to_messages())

            content = str(response.content)

            data: LLMQueryResult = parser.parse(content)
            return data.response
        except Exception as e:
            logger.error(
                "general_query_processing_failed",
                message_length=len(user_message),
                error_type=type(e).__name__,
                error=str(e),
                exc_info=True,
            )
            raise

from langchain.prompts import HumanMessagePromptTemplate, ChatPromptTemplate
from langchain.output_parsers import PydanticOutputParser
from langchain.agents import create_tool_calling_agent, AgentExecutor
from prompts.nextjs_prompt import NEXTJS_PROMPT
from prompts.query_prompt import QUERY_PROMPT
from services.models.ai_models import QueryResponse
from services.sandbox_service import SandboxService
from clients.openai_client import OpenAIClient
from pydantic import BaseModel, Field

'''
OpenAI service (sends requests to OpenAI llm). 
Can handle coding with tools and will execute them in the sandbox.
Can handle general queries as well.
'''

# TODO create methods and also instantiate OpenAIService with the sandbox service so it can run code directly?
# TODO add tools from langchain community / make my own. Shell tool, File System tool (to create/update files and then to read files)
# TODO add a prompt
class OpenAICodeAgentService:
    def __init__(self, openai_client: OpenAIClient, sandbox_service:SandboxService):
        self.llm = openai_client.get_client()
        code_agent_tools = sandbox_service.get_tools()
        code_agent = create_tool_calling_agent(
                llm=self.llm,
                prompt=NEXTJS_PROMPT,
                tools=code_agent_tools 
            )
        self.agent = AgentExecutor(agent=code_agent, tools=code_agent_tools, verbose=True)

    async def process_code_request(self, sandbox_id:str, user_message:str):
        try: 
            contextual_input = f"Sandbox ID: {sandbox_id}\nTask: {user_message}"
        
            result = await self.agent.ainvoke({"input": contextual_input})
            return result
        except Exception as e:
            raise e


class OpenAIService:
    def __init__(self, openai_client: OpenAIClient):
        self.llm = openai_client.get_client()

    async def process_query_request(self, user_message:str) -> str:
        try:
            parser = PydanticOutputParser(pydantic_object=QueryResponse)
            message = HumanMessagePromptTemplate.from_template(template=QUERY_PROMPT)
            chat_prompt = ChatPromptTemplate.from_messages(messages=[message])

            chat_prompt_with_values = chat_prompt.format_prompt(user_query=user_message, format_instructions=parser.get_format_instructions())
            
            response = self.llm(chat_prompt_with_values.to_messages())
            
            data = parser.parse(response.content)
            return data.response
        except Exception as e:
            pass

    
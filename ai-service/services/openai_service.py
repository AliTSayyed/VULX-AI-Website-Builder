from langchain_openai import ChatOpenAI
from langchain.prompts import HumanMessagePromptTemplate, ChatPromptTemplate
from langchain.output_parsers import PydanticOutputParser
from langchain.agents import AgentExecutor
from prompts.query_prompt import QUERY_PROMPT
from service_models import QueryResponse
from sandbox_service import SandboxService

'''
OpenAI service (sends requests to OpenAI llm). 
Can handle coding with tools and will execute them in the sandbox.
Can handle general queries as well.
'''

# TODO create methods and also instantiate OpenAIService with the sandbox service so it can run code directly?
# TODO add tools from langchain community / make my own. Shell tool, File System tool (to create/update files and then to read files)
# TODO add a prompt
class OpenAICodeAgentService:
    def __init__(self, llm: ChatOpenAI, agent:AgentExecutor, sandbox: SandboxService):
        self.llm = llm
        self.agent = agent
        self.sandbox = sandbox
    
    def process_code_agent_request(self, sandbox_id:str, user_message:str):
        try:
            self.agent.invoke({"user_message":user_message})
            pass
        except Exception as e:
            pass

class OpenAIService:
    def __init__(self, llm: ChatOpenAI):
        self.llm = llm

    def process_query_request(self, user_message:str) -> str:
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
    
    
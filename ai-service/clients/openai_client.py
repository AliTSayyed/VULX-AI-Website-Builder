from fastapi import Depends
from langchain_openai import ChatOpenAI
from langchain.agents import create_tool_calling_agent, AgentExecutor
from langchain.schema.runnable import Runnable 
from api.config import settings
from typing import Annotated
from prompts.nextjs_prompt import NEXTJS_PROMPT
from langchain_community.tools import ShellTool
from langchain_community.agent_toolkits import FileManagementToolkit

'''
Class that will configure and hold the OpenAI connection.
It will also old a connection to the open ai specific coding agent.
Contains a method to return the configured ChatOpenAI instance.
Create a function that will make an OpenAIClient instance.
Use fastapi's dependency annotation for dependency injection into routes.
'''

class OpenAIClient:
    def __init__(self):    
        self.client = ChatOpenAI(
                api_key=settings.openai_api_key,
                model=settings.openai_model,
                temperature=0.1,
                max_retries=15
        )

        shell_tool = ShellTool()
        file_management_toolkit = FileManagementToolkit(root_dir="/home/user/") # root dir determined by docker file in sandbox template
        file_management_tools = file_management_toolkit.get_tools() # giving access to all file management tools since operating in a sandbox
        all_tools = [shell_tool] + file_management_tools 

        code_agent = create_tool_calling_agent(
          llm=self.client,
          prompt= NEXTJS_PROMPT,
          tools= all_tools
        )

        self.agent_executor = AgentExecutor(agent=code_agent, tools=all_tools, verbose=True)

    def get_client(self) -> ChatOpenAI:
        return self.client
    
    def get_code_agent(self) -> AgentExecutor:
        return self.agent_executor

def get_openai_client() -> OpenAIClient:
    return OpenAIClient()

# Annotated[Type, Metadata]
openai_dependency = Annotated[OpenAIClient, Depends(get_openai_client)]

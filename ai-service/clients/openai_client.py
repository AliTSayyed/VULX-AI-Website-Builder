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

    def get_client(self) -> ChatOpenAI:
        return self.client
 
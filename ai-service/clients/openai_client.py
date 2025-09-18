from langchain_openai import ChatOpenAI
from api.config import settings
from prompts.nextjs_prompt import NEXTJS_PROMPT
'''
Class that will configure and hold the OpenAI connection.
Contains a method to return the configured ChatOpenAI instance.
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
 
from langchain_openai import ChatOpenAI
from api.config import settings
from utils.logging import logger

'''
Class that will configure and hold the OpenAI connection.
Contains a method to return the configured ChatOpenAI instance.
'''

class OpenAIClient:
    def __init__(self):    
        try:
            self.client = ChatOpenAI(
                api_key=settings.openai_api_key,
                model=settings.openai_model,
                temperature=0.1,
                max_retries=15
            )
        
        except Exception as e:
            logger.error("openai_client_initialization_failed",
                        model=settings.openai_model,
                        error_type=type(e).__name__,
                        error=str(e),
                        exc_info=True)
            raise

    def get_client(self) -> ChatOpenAI:
        return self.client
 
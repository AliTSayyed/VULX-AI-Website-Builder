from langchain_google_genai import ChatGoogleGenerativeAI
from api.config import settings

'''
Class that will configure and hold the Google connection.
Contains a method to return the configured Google instance.
'''

class GoogleClient():
    def __init__(self):    
        self.client = ChatGoogleGenerativeAI(
                google_api_key=settings.google_api_key,
                model=settings.google_model,
                temperature=0.1,
                max_retries=15
        )

    def get_client(self) -> ChatGoogleGenerativeAI:
        return self.client

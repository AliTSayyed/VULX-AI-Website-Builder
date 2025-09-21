from langchain_google_genai import ChatGoogleGenerativeAI
from api.config import settings
from utils.logging import logger

"""
Class that will configure and hold the Google connection.
Contains a method to return the configured Google instance.
"""


class GoogleClient:
    def __init__(self):
        try:
            self.client = ChatGoogleGenerativeAI(
                google_api_key=settings.google_api_key, model=settings.google_model
            )

        except Exception as e:
            logger.error(
                "google_client_initialization_failed",
                model=settings.google_model,
                error_type=type(e).__name__,
                error=str(e),
                exc_info=True,
            )
            raise

    def get_client(self) -> ChatGoogleGenerativeAI:
        return self.client

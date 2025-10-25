from langchain_anthropic import ChatAnthropic
from api.config import settings
from utils.logging import logger

"""
Class that will configure and hold the Anthropic/Claude connection.
Contains a method to return the configured ChatAnthropic instance.
"""


class AnthropicClient:
    def __init__(self):
        try:
            self.client = ChatAnthropic(
                api_key=settings.anthropic_api_key,
                model_name=settings.anthropic_model,
                temperature=0.1,
                max_retries=5,
                timeout=60,
                stop=[],
            )
        except Exception as e:
            logger.error(
                "anthropic_client_initialization_failed",
                model=settings.openai_model,
                error_type=type(e).__name__,
                error=str(e),
                exc_info=True,
            )
            raise

    def get_client(self) -> ChatAnthropic:
        return self.client

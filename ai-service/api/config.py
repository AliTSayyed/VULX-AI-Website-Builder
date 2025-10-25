from typing import Optional
from pydantic_settings import BaseSettings
from pydantic import SecretStr
import structlog


class Settings(BaseSettings):
    model_config = {"env_file": ".env"}

    environment: str = "local"

    # API Keys
    e2b_api_key: str = ""
    openai_api_key: Optional[SecretStr] = None
    google_api_key: Optional[SecretStr] = None
    anthropic_api_key: SecretStr = SecretStr("")

    # e2b template id
    e2b_sandbox_nextjs_template_id: str = ""

    # llm models
    openai_model: str = ""
    google_model: str = ""
    anthropic_model: str = ""

    # configure logger
    def configure_logging(self):
        structlog.configure(
            processors=[
                structlog.contextvars.merge_contextvars,
                structlog.processors.TimeStamper(fmt="iso"),
                structlog.processors.add_log_level,
                structlog.processors.JSONRenderer(),
            ]
        )


# global instance where env loading happens
settings = Settings()
settings.configure_logging()

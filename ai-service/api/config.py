from typing import Optional
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    # API Keys
    e2b_api_key: Optional[str] = None

    class Config:
        env_file = ".env"   

# global instance where env loading happens
settings = Settings()

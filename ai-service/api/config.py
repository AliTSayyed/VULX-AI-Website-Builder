from typing import Optional
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    # API Keys
    e2b_api_key: Optional[str] = None
    e2b_sandbox_nextjs_template_id: str
    openai_api_key: str
    openai_model: str 
    
    class Config:
        env_file = ".env"   

# global instance where env loading happens
settings = Settings()

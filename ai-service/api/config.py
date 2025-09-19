from typing import Optional
from pydantic_settings import BaseSettings
from pydantic import SecretStr

class Settings(BaseSettings):
    # API Keys
    e2b_api_key:str = "" 
    e2b_sandbox_nextjs_template_id: str = "" 
    openai_api_key: Optional[SecretStr] = None 
    openai_model: str = "" 
    google_api_key: Optional[SecretStr] = None 
    google_model:str = "" 
    
    class Config:
        env_file = ".env"   

# global instance where env loading happens
settings = Settings()

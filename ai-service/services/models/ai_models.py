from pydantic import BaseModel, Field, model_validator
from typing import List, Dict, Optional
from clients.openai_client import OpenAIClient
from clients.google_client import GoogleClient

# MAKE THESE OPTIONAL
class AIClient(BaseModel):
    openai_client: Optional[OpenAIClient] = None
    google_client: Optional[GoogleClient] = None
    
    @model_validator(mode='after')
    def validate_at_least_one_client(self):
        if not self.openai_client and not self.google_client:
            raise ValueError("At least one AI client must be provided")
        return self
    
    def get_client(self):
        if self.openai_client:
            return self.openai_client.get_client()
        elif self.google_client:
            return self.google_client.get_client()
        else:
            raise ValueError("No client available")

# response of ai agent after executing code in a sandbox
class CodeAgentResult(BaseModel):
    summary:str = Field(..., description=("Summary of the steps the agent took to complete the task."))

class CodeAgentData(BaseModel):
    summary:str = Field(..., description="summary of ai agent task completion") 
    commands:List[str] = Field(..., description="List of commands run by agent")
    files: Dict[str, str] = Field(..., description="all modified files with respective paths")

# response of llm after sending a regualr query
class QueryResult(BaseModel):
    response: str = Field(description="LLM Query Response")

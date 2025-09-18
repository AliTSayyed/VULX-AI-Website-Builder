from pydantic import BaseModel, Field
from typing import List, Dict

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

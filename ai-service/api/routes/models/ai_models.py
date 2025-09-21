from pydantic import BaseModel, Field
from typing import Dict, List


class AICodeAgentRequest(BaseModel):
    message: str = Field(..., description="human message to code agent")


class AICodeAgentResponse(BaseModel):
    human_message: str = Field(..., description="human message to code agent")
    summary: str = Field(..., description="code agent summary of task completion")
    commands: List[str] = Field(..., description="list of commands run by agent")
    files: Dict[str, str] = Field(
        ..., description="list of files updated and their respective paths"
    )


class AIRequest(BaseModel):
    message: str = Field(..., description="human message to llm")


class AIResponse(BaseModel):
    content: str = Field(..., description="open ai llm response to general query")

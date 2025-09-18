from pydantic import BaseModel, Field

# response of ai agent after executing code in a sandbox
class CodeAgentResponse(BaseModel):
    files: list[str]

# response of llm after sending a regualr query
class QueryResponse(BaseModel):
    response: str = Field(description="LLM Query Response")

from pydantic import BaseModel, Field

class OpenAIResponse(BaseModel):
    content: str = Field(description="open ai llm response to general query")

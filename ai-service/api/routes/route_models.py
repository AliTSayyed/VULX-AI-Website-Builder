from pydantic import BaseModel, Field

'''
OPEN AI AGENT
'''
class OpenAIResponse(BaseModel):
    content: str = Field(description="open ai llm response to general query")


'''
SANDBOX
'''
class SandboxResponse(BaseModel):
    id: str = Field(description="sandbox id")
    url: str = Field(description="sandbox url")

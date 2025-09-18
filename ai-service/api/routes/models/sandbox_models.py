from pydantic import BaseModel, Field
from services.models.sandbox_models import WriteFileData
from e2b_code_interpreter import WriteInfo
from typing import List, Optional

class CreateSandboxResponse(BaseModel):
    id: str = Field(..., description="sandbox id")
    url: str = Field(..., description="sandbox url")

class ListSandboxResponse(BaseModel):
    path: Optional[str]= Field(..., description="input path specified")
    files: List[WriteInfo] = Field(..., description="list of files from specified path")

class ReadSandboxResponse(BaseModel):
    path: str = Field(..., description="input path specified")
    content: str = Field(..., description="file content of path specified")

class ExecuteSandboxResponse(BaseModel):
    command: str = Field(..., description="input terminal command")
    stdout:str = Field(..., description="on stdout message")
    stderr:str = Field(..., description="on stderr message")

class WriteSandboxRequest(BaseModel):
    write_data:List[WriteFileData] = Field(..., description="list of file paths and content")

class WriteSandboxResponse(BaseModel):
    files_written_to:List[WriteInfo] = Field(..., description="successfully files written to")
    write_data:List[WriteFileData] = Field(..., description="list of file paths and content")


from pydantic import BaseModel, Field
from typing import List

# Extract relevant terminal output
class TerminalInfo(BaseModel):
    stdout: str = Field(..., description="output of successful command")
    stderr: str = Field(..., description="output of failed command")

# data for writing to a file used in the sandbox
class WriteFileData(BaseModel):
    path: str = Field(..., description="file path to write to")
    data: str = Field(..., description="content to write to the file")

# inputs for code agent sandbox related tools
class ReadToolInput(BaseModel) :
    sandbox_id:str = Field(..., description="id used to connect to sandbox")
    path:str = Field(..., description="path of file to read") 

class ListToolInput(BaseModel):
    sandbox_id: str = Field(..., description="id used to connect to sandbox")
    path: str = Field(..., description="directory path to list files from")

class WriteToolInput(BaseModel):
    sandbox_id: str = Field(..., description="id used to connect to sandbox")
    write_data: List[WriteFileData] = Field(..., description="list of files to write with path and content")
    
class CommandToolInput(BaseModel):
    sandbox_id: str = Field(..., description="id used to connect to sandbox")
    command: str = Field(..., description="terminal command to execute")

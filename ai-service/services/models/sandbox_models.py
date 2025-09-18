from pydantic import BaseModel, Field

# Extract relevant terminal output
class TerminalInfo(BaseModel):
    stdout: str = Field(..., description="output of successful command")
    stderr: str = Field(..., description="output of failed command")

# data for writing to a file used in the sandbox
class WriteFileData(BaseModel):
    path: str = Field(..., description="file path to write to")
    data: str = Field(..., description="content to write to the file")

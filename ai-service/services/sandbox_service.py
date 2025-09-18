from e2b_code_interpreter import Sandbox, WriteInfo, EntryInfo, CommandResult
from langchain.tools import BaseTool
from services.models.sandbox_models import TerminalInfo, WriteFileData
from pydantic import Field
from typing import List

'''
SandboxService: Handles E2B code execution sandbox operations
Creates secure sandboxes for running untrusted code
Manages sandbox lifecycle and file operations
Provides custom tools for agents to interact with the sandbox 
'''

class SandboxService():
    def __init__(self):
       self.tools = [self.SandboxReadTool(sandbox_service=self)] 
    
    def get_tools(self):
        return self.tools

    def create(self, template_id:str) -> Sandbox:
        sbx = Sandbox.create(template=template_id) # By default the sandbox is alive for 5 minutes
        return sbx
    
    def list_files(self, sandbox_id:str, path:str = "/home/user/") -> List[WriteInfo]:
        sbx:Sandbox = Sandbox.connect(sandbox_id=sandbox_id)
        sandbox_files:List[EntryInfo] = sbx.files.list(path)
        files:List[WriteInfo] = []
        for sandbox_file in sandbox_files:
            files.append(
                WriteInfo(
                    name=sandbox_file.name,
                    type=sandbox_file.type,
                    path=sandbox_file.path
                )
            )
        return files
    
    def read_file(self, sandbox_id:str, path:str) -> str:
        sbx = Sandbox.connect(sandbox_id=sandbox_id) 
        file_conent: str = sbx.files.read(path=path)
        return file_conent

    def write_files(self, sandbox_id: str, write_data:List[WriteFileData]) -> List[WriteInfo]:
        sbx = Sandbox.connect(sandbox_id)
        dict_data = [item.model_dump() for item in write_data] # converts pydantic model into a proper dict data structure
        result:List[WriteInfo] = sbx.files.write_files(dict_data)
        return result
    
    def execute_terminal_command(self, sandbox_id:str, command:str) -> TerminalInfo:
        sbx = Sandbox.connect(sandbox_id)
        result:CommandResult = sbx.commands.run(cmd=command)
        return TerminalInfo(
            stdout=result.stdout,
            stderr=result.stderr
        )
    
    class SandboxReadTool(BaseTool):
        name: str = "read_sandbox_files"
        description: str = "Read all the files in the sandbox. To access the sandbox, the first paramater must be the sandbox_id"
        sandbox_service: 'SandboxService' = Field(exclude=True)
        def _run(self, sandbox_id:str, path:str) -> str:
            file_content = self.sandbox_service.read_file(sandbox_id=sandbox_id, path=path)
            return 

"""
{
    "name": "page.tsx",
    "type": "file",
    "path": "/home/user/app/page.tsx"
  }
"""

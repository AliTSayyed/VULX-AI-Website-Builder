from langchain.callbacks.base import BaseCallbackHandler
from services.models.sandbox_models import WriteFileData
from typing import Dict, Any, List
from services.models.callback_models import CodeAgentCallBackResult
'''
When a coding agent uses tools to update files or execute commands,
we want to store those inputs (specifically after a sucessful tool use)
this way we know what files changed without relying on the AI to remeber 
what action(s) it took. This call back object is made per code request
'''

class CodeAgentCallBack(BaseCallbackHandler):
    def __init__(self):
        print("CALLBACK OBJECT CREATED")
        self.updated_files: Dict[str, str] = {} 
        self.commands_executed:List[str] = [] 
    
    def on_tool_end(self, output:str, **kwargs) -> None:
        print(f"CALLBACK TRIGGERED TOOL NAME: {kwargs.get('name', 'unknown')}")
        print(f"CALLBACK TRIGGERED Inputs: {kwargs.get('inputs', 'unknown')}")
        print(f"CALLBACK TRIGGERED OUTPUT: {output}")
        tool_name = kwargs.get("name", "")
        inputs = kwargs.get("inputs", {})

        if "failed to" not in output and "error" not in output: # only capturing successful tool usage 
            if tool_name == "write_sandbox_files":
                self._capture_file_writes(inputs)
            elif tool_name == "execute_sandbox_command":
                self._capture_command(inputs)
        
    def _capture_file_writes(self, inputs:Dict[str, Any]) -> None:
        write_data:List[WriteFileData] = inputs.get("write_data", []) # write data is the input name of write_sandbox_tool param

        for file in write_data:
            print(f"writing the folliwng {file.path}:{file.data}")
            self.updated_files[file.path] = file.data
    
    def _capture_command(self, inputs:Dict[str, Any]) -> None:
        command = inputs.get("command", "")
        if command:
            print(f"running this command: {command}")
            self.commands_executed.append(command)
    
    def get_result(self) -> CodeAgentCallBackResult:
        return CodeAgentCallBackResult(
            updated_files=self.updated_files,
            commands_executed=self.commands_executed
        ) 

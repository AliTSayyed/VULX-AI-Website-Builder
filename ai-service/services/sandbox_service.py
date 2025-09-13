from fastapi import Depends
from e2b_code_interpreter import Sandbox
from api.config import settings
from typing import Annotated

'''
SandboxService: Handles E2B code execution sandbox operations
Creates secure sandboxes for running untrusted code
Manages sandbox lifecycle and file operations
Provides dependency injection for FastAPI routes
'''
class SandboxService():
    def __init__(self):
        self.template_id = settings.e2b_sandbox_template_1 # subject to change if there are more templates

    def create(self) -> Sandbox:
        sbx = Sandbox.create(template=self.template_id) # By default the sandbox is alive for 5 minutes
        files = sbx.files.list("/")
        print(files)
        return sbx

    # TODO execute code inside of sandbox, pass in the code in the params
    def execute_code(self, id: str):
        sbx = Sandbox.connect(sandbox_id=id)
        return

def get_sandbox_service() -> SandboxService:
    return SandboxService()

sandbox_service_dependency = Annotated[SandboxService, Depends(get_sandbox_service)]

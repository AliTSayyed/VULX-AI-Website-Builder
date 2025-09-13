from e2b_code_interpreter import Sandbox

'''
SandboxService: Handles E2B code execution sandbox operations
Creates secure sandboxes for running untrusted code
Manages sandbox lifecycle and file operations
Provides dependency injection for FastAPI routes
'''
class SandboxService():
    def __init__(self):
        return

    def create(self, template_id:str) -> Sandbox:
        sbx = Sandbox.create(template_id=template_id) # By default the sandbox is alive for 5 minutes
        return sbx

    def execute_code(self, id: str):
        return

    def execute_terminal_command():
        return

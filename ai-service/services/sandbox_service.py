from e2b_code_interpreter import Sandbox
from api.config import settings

def createSandbox() -> Sandbox:
    sbx = Sandbox.create(template=settings.e2b_sandbox_template_1) # By default the sandbox is alive for 5 minutes

    files = sbx.files.list("/")
    print(files)
    return sbx

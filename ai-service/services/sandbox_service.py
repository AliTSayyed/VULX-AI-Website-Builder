from e2b_code_interpreter import Sandbox
from api.config import settings

def createSanbox():
    sbx = Sandbox.create(api_key=settings.e2b_api_key) # By default the sandbox is alive for 5 minutes
    execution = sbx.run_code("print('hello world')") # Execute Python inside the sandbox
    print(execution.logs)

    files = sbx.files.list("/")
    print(files)
    return None 

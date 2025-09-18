from fastapi import APIRouter, HTTPException
from loguru import logger
from api.routes.models.sandbox_models import (CreateSandboxResponse, 
ListSandboxResponse, ReadSandboxResponse, ExecuteSandboxResponse, 
WriteSandboxRequest, WriteSandboxResponse)
from api.dependencies import sandbox_service_dependency
from api.config import settings
from services.models.sandbox_models import TerminalInfo
from e2b_code_interpreter import Sandbox, WriteInfo
from typing import List
'''
This route is for the Golang service.
It will call this route to get a sandbox id / url which will be stored.
Subsequent calls to the ai services will use this sandbox id to execute
code in the correct sandbox.
If a sandbox id expires use this route to create a new sandbox id to run code in.
'''

router = APIRouter(
    prefix="/sandbox",
    tags=["Sandbox Service"]
)

@router.post("/")
async def create_sandbox(sandbox_service: sandbox_service_dependency) -> CreateSandboxResponse:
    try:
        logger.info("creating sandbox")
        sbx:Sandbox = sandbox_service.create(
            template_id=settings.e2b_sandbox_nextjs_template_id
            ) # currently only creating a nextjs sandbox  
        
        return CreateSandboxResponse(
            id=sbx.sandbox_id,
            url=sbx.get_host(3000)
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to create sandbox: {str(e)}")

@router.get("/{sandbox_id}/files")
async def list_sandbox_files(sandbox_id:str, path:str, sandbox_service: sandbox_service_dependency) -> ListSandboxResponse:
    try:
        logger.info("listing files in sandbox")

        # no need to access these paths
        forbidden_paths = ['/', '/root', '/etc', '/sys', '/proc']
        if path in forbidden_paths:
            raise HTTPException(status_code=403, detail=f"Access to path '{path}' is not allowed")

        files:List[WriteInfo] = sandbox_service.list_files(
            sandbox_id=sandbox_id, 
            path=path
        )  
        return ListSandboxResponse(
            path=path,
            files=files
        ) 

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to list files from sandbox: {str(e)}")
    
@router.get("/{sandbox_id}/file")
async def read_sandbox_file(sandbox_id:str, path:str, sandbox_service: sandbox_service_dependency) -> ReadSandboxResponse:
    try:
        logger.info("reading files in sandbox")

        file_content:str = sandbox_service.read_file(
            sandbox_id=sandbox_id, 
            path=path
        )  
        return ReadSandboxResponse(
            path=path,
            content=file_content
        ) 

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to read files from sandbox: {str(e)}")

@router.post("/{sandbox_id}/command")
async def execute_sandbox_command(sandbox_id:str, command:str, sandbox_service: sandbox_service_dependency) -> ExecuteSandboxResponse:
    try:
        logger.info("executing command in sandbox")

        result:TerminalInfo = sandbox_service.execute_terminal_command(
            sandbox_id=sandbox_id, 
            command=command
        )  
        return ExecuteSandboxResponse(
            command=command,
            stdout=result.stdout,
            stderr=result.stderr,
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to execute terminal command in sandbox: {str(e)}")

@router.post("/{sandbox_id}/files")
async def write_sandbox_files(sandbox_id:str, files:WriteSandboxRequest, sandbox_service: sandbox_service_dependency) -> WriteSandboxResponse:
    try: 

        logger.info(f"writing {len(files.write_data)} file(s) to sandbox")

        result:List[WriteInfo] = sandbox_service.write_files(sandbox_id=sandbox_id, write_data=files.write_data)
        return WriteSandboxResponse(
            files_written_to=result,
            write_data=files.write_data
        ) 

    except Exception as e:
            raise HTTPException(status_code=500, detail=f"Failed to write files in sandbox: {str(e)}")

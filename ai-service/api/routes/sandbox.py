from fastapi import APIRouter, HTTPException
from api.routes.models.sandbox_models import (
    CreateSandboxResponse,
    ListSandboxResponse,
    ReadSandboxResponse,
    ExecuteSandboxResponse,
    WriteSandboxRequest,
    WriteSandboxResponse,
)
from api.dependencies import sandbox_service_dependency
from api.config import settings
from services.models.sandbox_models import TerminalInfo
from e2b_code_interpreter import Sandbox, WriteInfo
from typing import List
from utils.logging import logger

"""
Create route is for the Golang service.
It will call this route to get a sandbox id / url which will be stored.
Subsequent calls to the ai services will use this sandbox id to execute
code in the correct sandbox.
If a sandbox id expires use this route to create a new sandbox id to run code in.
Other routes are for development
"""

router = APIRouter(prefix="/sandbox", tags=["Sandbox Service"])


@router.post("/")
async def create_sandbox(
    sandbox_service: sandbox_service_dependency,
) -> CreateSandboxResponse:
    logger.info("sandbox_creation_started")
    try:
        sbx: Sandbox = sandbox_service.create(
            template_id=settings.e2b_sandbox_nextjs_template_id
        )  # currently only creating a nextjs sandbox
        logger.info("sandbox_creation_completed")
        return CreateSandboxResponse(id=sbx.sandbox_id, url=sbx.get_host(3000))
    except Exception as e:
        logger.error(
            "sandbox_creation_failed",
            error_type=type(e).__name__,
            error=str(e),
            exc_info=True,
        )
        raise HTTPException(
            status_code=500, detail=f"Failed to create sandbox: {str(e)}"
        )


@router.get("/{sandbox_id}/files")
async def list_sandbox_files(
    sandbox_id: str, path: str, sandbox_service: sandbox_service_dependency
) -> ListSandboxResponse:
    logger.info("file_listing_started", path=path)
    try:
        files: List[WriteInfo] = sandbox_service.list_files(
            sandbox_id=sandbox_id, path=path
        )

        logger.info("file_listing_completed", path=path, file_count=len(files))

        return ListSandboxResponse(path=path, files=files)

    except Exception as e:
        logger.error(
            "file_listing_failed",
            path=path,
            error_type=type(e).__name__,
            error=str(e),
            exc_info=True,
        )
        raise HTTPException(
            status_code=500, detail=f"Failed to list files from sandbox: {str(e)}"
        )


@router.get("/{sandbox_id}/file")
async def read_sandbox_file(
    sandbox_id: str, path: str, sandbox_service: sandbox_service_dependency
) -> ReadSandboxResponse:
    logger.info("file_read_started", path=path)
    try:
        file_content: str = sandbox_service.read_file(sandbox_id=sandbox_id, path=path)

        logger.info("file_read_completed", path=path, content_length=len(file_content))

        return ReadSandboxResponse(path=path, content=file_content)

    except Exception as e:
        logger.error(
            "file_read_failed",
            path=path,
            error_type=type(e).__name__,
            error=str(e),
            exc_info=True,
        )
        raise HTTPException(
            status_code=500, detail=f"Failed to read files from sandbox: {str(e)}"
        )


@router.post("/{sandbox_id}/command")
async def execute_sandbox_command(
    sandbox_id: str, command: str, sandbox_service: sandbox_service_dependency
) -> ExecuteSandboxResponse:
    logger.info("command_execution_started", command=command)
    try:
        result: TerminalInfo = sandbox_service.execute_terminal_command(
            sandbox_id=sandbox_id, command=command
        )

        logger.info(
            "command_execution_completed",
            command=command,
            stdout_length=len(result.stdout) if result.stdout else 0,
            stderr_length=len(result.stderr) if result.stderr else 0,
            has_errors=bool(result.stderr),
        )

        return ExecuteSandboxResponse(
            command=command,
            stdout=result.stdout,
            stderr=result.stderr,
        )

    except Exception as e:
        logger.error(
            "command_execution_failed",
            command=command,
            error_type=type(e).__name__,
            error=str(e),
            exc_info=True,
        )
        raise HTTPException(
            status_code=500,
            detail=f"Failed to execute terminal command in sandbox: {str(e)}",
        )


@router.post("/{sandbox_id}/files")
async def write_sandbox_files(
    sandbox_id: str,
    files: WriteSandboxRequest,
    sandbox_service: sandbox_service_dependency,
) -> WriteSandboxResponse:
    logger.info("file_write_started", file_count=len(files.write_data))
    try:
        result: List[WriteInfo] = sandbox_service.write_files(
            sandbox_id=sandbox_id, write_data=files.write_data
        )

        logger.info(
            "file_write_completed",
            files_requested=len(files.write_data),
            files_written=len(result),
        )

        return WriteSandboxResponse(
            files_written_to=result, write_data=files.write_data
        )

    except Exception as e:
        logger.error(
            "file_write_failed",
            file_count=len(files.write_data),
            error_type=type(e).__name__,
            error=str(e),
            exc_info=True,
        )
        raise HTTPException(
            status_code=500, detail=f"Failed to write files in sandbox: {str(e)}"
        )

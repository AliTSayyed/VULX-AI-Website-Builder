from fastapi import APIRouter, HTTPException
from loguru import logger
from pydantic import BaseModel, Field
from services.sandbox_service import sandbox_service_dependency 
from route_models import SandboxResponse
'''
This route is for the Golang service.
It will call this route to get a sandbox id / url which will be stored.
Subsequent calls to the ai services will use this sandbox id to execute
code in the correct sandbox.
If a sandbox id expires use this route to create a new sandbox id to run code in.
'''

router = APIRouter(
    prefix="/sandbox"
)

@router.get("/create")
async def create_sandbox(sandbox_service: sandbox_service_dependency):
    try:
        logger.info("creating sandbox")

        sbx = sandbox_service.create() 
        
        return SandboxResponse(
            id=sbx.sandbox_id,
            url=sbx.get_host(3000)
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to create sandbox: {str(e)}")

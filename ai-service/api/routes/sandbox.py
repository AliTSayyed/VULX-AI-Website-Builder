from fastapi import APIRouter, HTTPException
from services import sandbox_service
from loguru import logger
from pydantic import BaseModel

router = APIRouter(
    prefix="/sandbox"
)

class SandboxResponse(BaseModel):
    id: str 
    url: str

@router.get("/create")
async def create_sandbox():
    try:
        sbx = sandbox_service.createSandbox()
        
        return SandboxResponse(
            id=sbx.sandbox_id,
            url=sbx.get_host(3000)
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to create sandbox: {str(e)}")

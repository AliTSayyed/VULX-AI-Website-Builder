from fastapi import APIRouter
from services import sandbox_service
router = APIRouter(
    prefix="/sandbox"
)

@router.get("/create")
def get_E2B_sandbox():
   sandbox_service.createSandbox() 
    
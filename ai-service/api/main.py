from fastapi import FastAPI, APIRouter
from api.routes.sandbox import router as sandbox_router

# create server
ai_service = FastAPI()

# add api/v1 prefix to all enpoints
api_v1_router = APIRouter(prefix="/api/v1")
api_v1_router.include_router(sandbox_router)

# main server only knows the api v1 router
ai_service.include_router(api_v1_router)


from fastapi import FastAPI, APIRouter
from api.routes.healthz import router as health_router
from api.routes.sandbox import router as sandbox_router
from fastapi.middleware.cors import CORSMiddleware

# create server
ai_service = FastAPI()

# add api/v1 prefix to all enpoints
api_v1_router = APIRouter(prefix="/ai-service/v1")
api_v1_router.include_router(health_router)
api_v1_router.include_router(sandbox_router)

# main server only knows the api v1 router
ai_service.include_router(api_v1_router)

# need to configure for production 
ai_service.add_middleware(
    CORSMiddleware,
    allow_origins=["http://api:8080"], 
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

from fastapi import FastAPI, APIRouter
from api.routes.healthz import router as health_router
from api.routes.sandbox import router as sandbox_router
from api.routes.openai import router as openai_router
from api.routes.google import router as google_router
from api.routes.anthropic import router as anthropic_router
from fastapi.middleware.cors import CORSMiddleware
from api.config import settings
from utils.logging import LoggingMiddleWare

# env
env = settings.environment
PRODUCTION = env == "production"

# create server
ai_service = FastAPI(
    docs_url=None if PRODUCTION else "/docs",
    redoc_url=None if PRODUCTION else "/redoc",
    openapi_url=None if PRODUCTION else "/openapi.json",
)

# add api/v1 prefix to all enpoints
api_v1_router = APIRouter(prefix="/ai-service/v1")
api_v1_router.include_router(health_router)
api_v1_router.include_router(sandbox_router)
api_v1_router.include_router(openai_router)
api_v1_router.include_router(google_router)
api_v1_router.include_router(anthropic_router)

# main server only knows the api v1 router
ai_service.include_router(api_v1_router)

# need to configure for production
ai_service.add_middleware(
    CORSMiddleware,
    allow_origins=["http://api:8080"],  # allow requests from golang server
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["*"],
)

# log all http requests
ai_service.add_middleware(LoggingMiddleWare)

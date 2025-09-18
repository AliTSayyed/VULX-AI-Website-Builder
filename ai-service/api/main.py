from fastapi import FastAPI, APIRouter, Request
from starlette.responses import Response
from api.routes.healthz import router as health_router
from api.routes.sandbox import router as sandbox_router
from api.routes.openai import router as openai_router
from fastapi.middleware.cors import CORSMiddleware
from loguru import logger
import time

# create server
ai_service = FastAPI()

# add api/v1 prefix to all enpoints
api_v1_router = APIRouter(prefix="/ai-service/v1")
api_v1_router.include_router(health_router)
api_v1_router.include_router(sandbox_router)
api_v1_router.include_router(openai_router)

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

 # Log  request
@ai_service.middleware("http")
async def log_requests(request: Request, call_next):
    start_time = time.time()
    logger.info(f"Request: {request.method} {request.url}")
    response: Response = await call_next(request)
    duration = time.time() - start_time
    logger.info(f"Response: {request.method} {request.url} {request.json}- Status: {response.status_code} - Duration: {duration:.3f}s")
    return response

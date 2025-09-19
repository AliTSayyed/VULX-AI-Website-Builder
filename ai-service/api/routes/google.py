from fastapi import APIRouter, HTTPException
from api.routes.models.ai_models import AICodeAgentRequest, AICodeAgentResponse, AIRequest, AIResponse
from api.dependencies import google_service_dependency, google_code_agent_service_dependency
from services.models.ai_models import CodeAgentData
from loguru import logger
import traceback
'''
This route will be used by the Golang service.
It is tailored for responses by Google specifically.
The related Google service will handle running code in the sandbox or a general query.
# TODO  add user messages context in the param for a better answer.
'''
router = APIRouter (
    prefix="/google",
    tags=["Google Service"],
)


@router.post("/{sandbox_id}/code")
async def code_agent_request(sandbox_id:str, request:AICodeAgentRequest, google_code_agent: google_code_agent_service_dependency):
    try:
        result:CodeAgentData = await google_code_agent.process_code_request(sandbox_id=sandbox_id, user_message=request.message)
        return AICodeAgentResponse(
            human_message=request.message,
            summary=result.summary,
            commands=result.commands,
            files=result.files
        ) 
    except Exception as e:
        logger.error(f"Full error: {traceback.format_exc()}")
        raise HTTPException(status_code=500, detail=f"google code agent failed: {str(e)}") 
    
@router.post("/query")
async def query_request(request: AIRequest, google_service: google_service_dependency ) -> AIResponse:
    try:
        result = await google_service.process_query_request(user_message=request.message)
        return AIResponse(content=result)

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"google query failed: {str(e)}") 


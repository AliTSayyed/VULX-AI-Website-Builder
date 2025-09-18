from fastapi import APIRouter, HTTPException
from api.routes.models.openai_models import OpenAICodeAgentRequest, OpenAICodeAgentResponse, OpenAIRequest, OpenAIResponse
from api.dependencies import openai_service_dependency, openai_code_agent_service_dependency
from loguru import logger
import traceback
'''
This route will be used by the Golang service.
It is tailored for responses by OpenAI specifically.
The related OpenAI service will handle running code in the sandbox or a general query.
# TODO  add user messages context in the param for a better answer.
'''
router = APIRouter (
    prefix="/openai",
    tags=["Openai Service"],
)


@router.post("/{sandbox_id}/code")
async def code_agent_request(sandbox_id:str, request:OpenAICodeAgentRequest, openai_code_agent: openai_code_agent_service_dependency):
    try:
        result = await openai_code_agent.process_code_request(sandbox_id=sandbox_id, user_message=request.message)
        return OpenAICodeAgentResponse(
            human_message=request.message,
            summary=result.summary,
            commands=result.commands,
            files=result.files
        ) 
    except Exception as e:
        logger.error(f"Full error: {traceback.format_exc()}")
        raise HTTPException(status_code=500, detail=f"openai code agent failed: {str(e)}") 
    
@router.post("/query")
async def query_request(request: OpenAIRequest, openai_service: openai_service_dependency ) -> OpenAIResponse:
    try:
        result = await openai_service.process_query_request(user_message=request.message)
        return OpenAIResponse(content=result)

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"openai query failed: {str(e)}") 


from fastapi import APIRouter, HTTPException
from api.routes.models.ai_models import (
    AICodeAgentRequest,
    AICodeAgentResponse,
    AIRequest,
    AIResponse,
)
from api.dependencies import (
    google_service_dependency,
    google_code_agent_service_dependency,
)
from services.models.ai_models import CodeAgentData
from utils.logging import logger

"""
This route will be used by the Golang service.
It is tailored for responses by Google specifically.
The related Google service will handle running code in the sandbox or a general query.
# TODO  add user messages context in the param for a better answer.
"""
router = APIRouter(
    prefix="/google",
    tags=["Google Service"],
)


@router.post("/{sandbox_id}/code")
async def code_agent_request(
    sandbox_id: str,
    request: AICodeAgentRequest,
    google_code_agent: google_code_agent_service_dependency,
) -> AICodeAgentResponse:
    logger.info(
        "google_code_agent_started",
        sandbox_id=sandbox_id,
        message_length=len(request.message),
    )
    try:
        result: CodeAgentData = await google_code_agent.process_code_request(
            sandbox_id=sandbox_id, user_message=request.message
        )

        logger.info(
            "google_code_agent_completed",
            sandbox_id=sandbox_id,
            commands_executed=len(result.commands),
            files_modified=len(result.files),
        )

        return AICodeAgentResponse(
            human_message=request.message,
            summary=result.summary,
            commands=result.commands,
            files=result.files,
        )

    except Exception as e:
        logger.error(
            "google_code_agent_failed",
            sandbox_id=sandbox_id,
            error_type=type(e).__name__,
            error=str(e),
            exc_info=True,
        )
        raise HTTPException(
            status_code=500, detail=f"google code agent failed: {str(e)}"
        )


@router.post("/query")
async def query_request(
    request: AIRequest, google_service: google_service_dependency
) -> AIResponse:
    logger.info("google_query_started", message_length=len(request.message))
    try:
        result: str = await google_service.process_query_request(
            user_message=request.message
        )

        logger.info("google_query_completed", message_length=len(request.message))

        return AIResponse(content=result)

    except Exception as e:
        logger.error(
            "google_query_failed",
            error_type=type(e).__name__,
            error=str(e),
            exc_info=True,
        )
        raise HTTPException(status_code=500, detail=f"google query failed: {str(e)}")

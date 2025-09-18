from fastapi import APIRouter, HTTPException
from api.routes.models.openai_models import OpenAIResponse
from api.dependencies import openai_service_dependency, openai_code_agent_service_dependency
'''
This route will be used by the Golang service.
It is tailored for responses by OpenAI specifically.
The related OpenAI service will handle running code in the sandbox or a general query.
'''
router = APIRouter (
    prefix="/openai",
    tags=["Openai Service"]
)


# TODO after a sand box is created from the create sandbox route, we need to send a request to this coding agent
# this agent will need the user query and the id of the sandbox to execute its code on
@router.post("/{sandbox_id}/code")
async def code_agent_request(sandbox_id:str, message:str, openai_code_agent: openai_code_agent_service_dependency):
    try:
        # openai_code_agent.process_code_request(sandbox_id=id, user_message=message)
       return 
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"openai code agent failed: {str(e)}") 
    
# if a user sends a regular question then ai model should be queried.
# TODO  add user messages context in the param for a better answer.
@router.post("/query")
async def query_request(message: str, openai_service: openai_service_dependency ) -> OpenAIResponse:
    try:
        result = await openai_service.process_query_request(user_message=message)
        return OpenAIResponse(content=result)

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"openai query failed: {str(e)}") 


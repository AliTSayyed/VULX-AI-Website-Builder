from fastapi import APIRouter, HTTPException
from clients.openai_client import openai_dependency
from services.openai_service import OpenAICodeAgentService, OpenAIService
from services.sandbox_service import sandbox_service_dependency 
from route_models import OpenAIResponse
'''
This route will be used by the Golang service.
It is tailored for responses by OpenAI specifically.
The related OpenAI service will handle running code in the sandbox or a general query.
'''
router = APIRouter (
    prefix="/openai-agent"
)


# TODO after a sand box is created from the create sandbox route, we need to send a request to this coding agent
# this agent will need the user query and the id of the sandbox to execute its code on
@router.post("/code/{id}")
async def code_agent_request(id:id, message:str, openai: openai_dependency):
    try:
        openai_code_agent_service = OpenAICodeAgentService(llm=openai.get_client(), agent=openai.get_code_agent(), sandbox=sandbox_service_dependency)
        openai_code_agent_service.process_code_agent_request(sandbox_id=id, user_message=message)
        pass
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"openai code agent failed : {str(e)}") 
    
# if a user sends a regular question then ai model should be queried.
# TODO  add user messages context in the param for a better answer.
@router.post("/query")
async def query_request(message: str, openai: openai_dependency) -> OpenAIResponse:
    try:
        openai_service = OpenAIService(llm=openai.get_client())
        result = openai_service.process_query_request(user_message=message)
        return OpenAIResponse(content=result)

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"openai query failed : {str(e)}") 


from fastapi import Depends
from typing import Annotated
from services.sandbox_service import SandboxService
from clients.openai_client import OpenAIClient
from services.openai_services import OpenAICodeAgentService, OpenAIService
'''
This file will instantiate all the necessary objects that will be passed around
(Dependency Injection) in various services and routes.
Avoids creating many instances of the same object.
Annotated[Type, Metadata]
'''

# create the sandbox service once (holds interations with e2b sandbox)
def get_sandbox_service() -> SandboxService:
    return SandboxService()

sandbox_service_dependency = Annotated[SandboxService, Depends(get_sandbox_service)]

# create the open ai client object (holds connection to openai llm)
def get_openai_client() -> OpenAIClient:
    return OpenAIClient()

openai_dependency = Annotated[OpenAIClient, Depends(get_openai_client)]

# create open ai coding agent 
def get_openai_code_agent_service(
        openai: openai_dependency,
        sandbox: sandbox_service_dependency) -> OpenAICodeAgentService:  
    return OpenAICodeAgentService(openai_client=openai, sandbox_service=sandbox)

openai_code_agent_service_dependency = Annotated[OpenAICodeAgentService, Depends(get_openai_code_agent_service)]

# create general openai llm service
def get_openai_service(openai: openai_dependency) -> OpenAIService:
    return OpenAIService(openai_client=openai)

openai_service_dependency = Annotated[OpenAIService, Depends(get_openai_service)]



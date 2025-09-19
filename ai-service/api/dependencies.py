from fastapi import Depends
from typing import Annotated
from services.sandbox_service import SandboxService
from clients.openai_client import OpenAIClient
from clients.google_client import GoogleClient
from services.ai_services import CodeAgentService, GeneralAIService
from services.models.ai_models import AIClient
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
        sandbox: sandbox_service_dependency) -> CodeAgentService:
    ai_client = AIClient(openai_client=openai)  
    return CodeAgentService(llm=ai_client, sandbox_service=sandbox)

openai_code_agent_service_dependency = Annotated[CodeAgentService, Depends(get_openai_code_agent_service)]

# create general openai llm service
def get_openai_service(openai: openai_dependency) -> GeneralAIService:
    ai_client = AIClient(openai_client=openai)  
    return GeneralAIService(llm=ai_client)

openai_service_dependency = Annotated[GeneralAIService, Depends(get_openai_service)]

# create the google ai client object (holds connection to google ai llm)
def get_google_client() -> GoogleClient:
    return GoogleClient()

google_dependency = Annotated[GoogleClient, Depends(get_google_client)]

# create google coding agent 
def get_google_code_agent_service(
        google: google_dependency,
        sandbox: sandbox_service_dependency) -> CodeAgentService: 
    ai_client = AIClient(google_client=google)  
    return CodeAgentService(llm=ai_client, sandbox_service=sandbox)

google_code_agent_service_dependency = Annotated[CodeAgentService, Depends(get_google_code_agent_service)]

# create general google llm service
def get_google_service(google: google_dependency) -> GeneralAIService:
    ai_client = AIClient(google_client=google)  
    return GeneralAIService(llm=ai_client)

google_service_dependency = Annotated[GeneralAIService, Depends(get_google_service)]




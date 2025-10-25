from fastapi import Depends
from typing import Annotated
from services.sandbox_service import SandboxService
from clients.openai_client import OpenAIClient
from clients.google_client import GoogleClient
from clients.anthropic_client import AnthropicClient
from services.ai_services import CodeAgentService, GeneralAIService
from services.models.ai_models import AIClient
from utils.logging import logger
from functools import lru_cache

"""
This file will instantiate all the necessary objects that will be passed around
(Dependency Injection) in various services and routes.
Avoids creating many instances of the same object.
Annotated[Type, Metadata]
"""


# create the sandbox service once (holds interations with e2b sandbox)
@lru_cache()
def get_sandbox_service() -> SandboxService:
    logger.info("sandbox_service_client_created")
    return SandboxService()


sandbox_service_dependency = Annotated[SandboxService, Depends(get_sandbox_service)]


# create the open ai client object (holds connection to openai llm)
@lru_cache()
def get_openai_client() -> OpenAIClient:
    logger.info("openai_client_created")
    return OpenAIClient()


openai_dependency = Annotated[OpenAIClient, Depends(get_openai_client)]


# create open ai coding agent
@lru_cache()
def get_openai_code_agent_service(
    openai: openai_dependency, sandbox: sandbox_service_dependency
) -> CodeAgentService:
    logger.info("openai_code_agent_service_client_created")
    ai_client = AIClient(openai_client=openai)
    return CodeAgentService(llm=ai_client, sandbox_service=sandbox)


openai_code_agent_service_dependency = Annotated[
    CodeAgentService, Depends(get_openai_code_agent_service)
]


# create general openai llm service
@lru_cache()
def get_openai_service(openai: openai_dependency) -> GeneralAIService:
    logger.info("openai_general_service_client_created")
    ai_client = AIClient(openai_client=openai)
    return GeneralAIService(llm=ai_client)


openai_service_dependency = Annotated[GeneralAIService, Depends(get_openai_service)]


# create the google ai client object (holds connection to google ai llm)
@lru_cache()
def get_google_client() -> GoogleClient:
    logger.info("google_client_created")
    return GoogleClient()


google_dependency = Annotated[GoogleClient, Depends(get_google_client)]


# create google coding agent
@lru_cache()
def get_google_code_agent_service(
    google: google_dependency, sandbox: sandbox_service_dependency
) -> CodeAgentService:
    logger.info("google_code_agent_service_client_created")
    ai_client = AIClient(google_client=google)
    return CodeAgentService(llm=ai_client, sandbox_service=sandbox)


google_code_agent_service_dependency = Annotated[
    CodeAgentService, Depends(get_google_code_agent_service)
]


# create general google llm service
@lru_cache()
def get_google_service(google: google_dependency) -> GeneralAIService:
    logger.info("google_general_service_client_created")
    ai_client = AIClient(google_client=google)
    return GeneralAIService(llm=ai_client)


google_service_dependency = Annotated[GeneralAIService, Depends(get_google_service)]


# create the anthropic client object (holds connection to claude llm)
@lru_cache()
def get_anthropic_client() -> AnthropicClient:
    logger.info("anthropic_client_created")
    return AnthropicClient()


anthropic_dependency = Annotated[AnthropicClient, Depends(get_anthropic_client)]


# create the anthropic coding agent
@lru_cache()
def get_anthropic_code_agent_service(
    anthropic: anthropic_dependency, sandbox: sandbox_service_dependency
) -> CodeAgentService:
    logger.info("anthropic_code_agent_service_client_created")
    ai_client = AIClient(anthropic_client=anthropic)
    return CodeAgentService(llm=ai_client, sandbox_service=sandbox)


anthropic_code_agent_service_dependency = Annotated[
    CodeAgentService, Depends(get_anthropic_code_agent_service)
]


# create general anthropic llm service
@lru_cache()
def get_anthropic_service(anthropic: anthropic_dependency) -> GeneralAIService:
    logger.info("anthropic_general_service_client_created")
    ai_client = AIClient(anthropic_client=anthropic)
    return GeneralAIService(llm=ai_client)


anthropic_service_dependency = Annotated[
    GeneralAIService, Depends(get_anthropic_service)
]

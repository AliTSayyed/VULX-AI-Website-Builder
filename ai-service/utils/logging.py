from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.responses import Response
from structlog.stdlib import BoundLogger
import structlog
import uuid

logger: BoundLogger = structlog.get_logger()


class LoggingMiddleWare(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next) -> Response:
        structlog.contextvars.clear_contextvars()
        structlog.contextvars.bind_contextvars(
            request_id=str(uuid.uuid4()),
        )

        logger.info("request_received", method=request.method, path=request.url.path)

        response = await call_next(request)

        logger.info("request_completed", status_code=response.status_code)

        return response

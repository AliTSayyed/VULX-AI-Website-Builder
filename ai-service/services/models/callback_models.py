from pydantic import BaseModel
from typing import Dict, List


class CodeAgentCallBackResult(BaseModel):
    updated_files: Dict[str, str]
    commands_executed: List[str]

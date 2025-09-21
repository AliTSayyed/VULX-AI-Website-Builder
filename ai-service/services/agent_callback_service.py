from langchain.callbacks.base import BaseCallbackHandler
from typing import Dict, Any, List
from services.models.callback_models import CodeAgentCallBackResult
from utils.logging import logger

"""
When a coding agent uses tools to update files or execute commands,
we want to store those inputs (specifically after a sucessful tool use)
this way we know what files changed without relying on the AI to remeber 
what action(s) it took. This call back object is made per code request
"""


class CodeAgentCallBack(BaseCallbackHandler):
    def __init__(self) -> None:
        self.updated_files: Dict[str, str] = {}
        self.commands_executed: List[str] = []

        # store agent tool inputs in pending, if tool was successful then return outputs to user
        self.pending_files: Dict[str, str] = {}
        self.pending_commands: List[str] = []

    def on_tool_start(
        self,
        serialized: dict[str, Any],
        input_str: str,
        *,
        inputs: dict[str, Any],
        **kwargs,
    ) -> None:
        logger.debug("agent_tool_started", inputs=inputs)

        if not inputs:
            return
        if "write_data" in inputs:
            self._capture_file_writes(inputs)
        elif "command" in inputs:
            self._capture_command(inputs)

    def on_tool_end(self, output: str, **kwargs: Any) -> None:
        tool_name = kwargs.get("name", "")
        success = "failed to" not in output and "error" not in output
        if not success:
            logger.warning(
                "agent_tool_failed", tool_name=tool_name, output_preview=output[:100]
            )
        else:
            logger.debug("agent_tool_completed", tool_name=tool_name, success=True)

        if tool_name == "write_sandbox_files" or tool_name == "execute_sandbox_command":
            # Move pending to final on success
            if "failed to" not in output and "error" not in output:
                self.updated_files.update(self.pending_files)
                self.commands_executed.extend(self.pending_commands)

        # Clear pending regardless
        self.pending_files.clear()
        self.pending_commands.clear()

    def _capture_file_writes(self, inputs: Dict[str, Any]) -> None:
        write_data = inputs.get(
            "write_data", []
        )  # write data is the input name of write_sandbox_tool param

        for file in write_data:
            self.pending_files[file["path"]] = file["data"]

    def _capture_command(self, inputs: Dict[str, Any]) -> None:
        command = inputs.get("command", "")
        if command:
            self.pending_commands.append(command)

    def get_result(self) -> CodeAgentCallBackResult:
        logger.debug(
            "callback_results_retrieved",
            total_files=len(self.updated_files),
            total_commands=len(self.commands_executed),
        )

        return CodeAgentCallBackResult(
            updated_files=self.updated_files, commands_executed=self.commands_executed
        )

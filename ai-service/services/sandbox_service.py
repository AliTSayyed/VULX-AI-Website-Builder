from e2b_code_interpreter import Sandbox, WriteInfo, EntryInfo, CommandResult
from langchain.tools import BaseTool
from services.models.sandbox_models import (
    TerminalInfo,
    WriteEntry,
    ReadToolInput,
    ListToolInput,
    WriteToolInput,
    CommandToolInput,
)
from pydantic import BaseModel, Field
from typing import List, Type
from utils.logging import logger

"""
SandboxService: Handles E2B code execution sandbox operations
Creates secure sandboxes for running untrusted code
Manages sandbox lifecycle and file operations
Provides custom tools for agents to interact with the sandbox 
None of these are async, forces the llm to do one action at a time.
"""


class SandboxService:
    def __init__(self) -> None:
        try:
            self.tools: List[BaseTool] = [
                self.SandboxListTool(sandbox_service=self),
                self.SandboxReadTool(sandbox_service=self),
                self.SandboxWriteTool(sandbox_service=self),
                self.SandboxCommandTool(sandbox_service=self),
            ]

        except Exception as e:
            logger.error(
                "sandbox_service_initialization_failed",
                error_type=type(e).__name__,
                error=str(e),
                exc_info=True,
            )
            raise

    def get_tools(self) -> List[BaseTool]:
        return self.tools

    def create(self, template_id: str) -> Sandbox:
        sbx = Sandbox.create(
            template=template_id
        )  # By default the sandbox is alive for 5 minutes
        return sbx

    def list_files(self, sandbox_id: str, path: str = "/home/user/") -> List[WriteInfo]:
        # path check
        forbidden_paths = ["/", "/root", "/etc", "/sys", "/proc"]
        if path in forbidden_paths:
            raise Exception(f"do not access the following path: {path} in the sandbox")

        sbx = Sandbox.connect(sandbox_id=sandbox_id)
        sandbox_files: List[EntryInfo] = sbx.files.list(path)
        files: List[WriteInfo] = []
        for sandbox_file in sandbox_files:
            files.append(
                WriteInfo(
                    name=sandbox_file.name,
                    type=sandbox_file.type,
                    path=sandbox_file.path,
                )
            )
        return files

    def read_file(self, sandbox_id: str, path: str) -> str:
        sbx = Sandbox.connect(sandbox_id=sandbox_id)
        file_content: str = sbx.files.read(path=path)
        return file_content

    def write_files(
        self, sandbox_id: str, write_data: List[WriteEntry]
    ) -> List[WriteInfo]:
        sbx = Sandbox.connect(sandbox_id)
        dict_data = [
            item.model_dump() for item in write_data
        ]  # converts pydantic model into a proper dict data structure for the sandbox api
        result: List[WriteInfo] = sbx.files.write_files(files=dict_data)  # type: ignore
        return result

    def execute_terminal_command(self, sandbox_id: str, command: str) -> TerminalInfo:
        sbx = Sandbox.connect(sandbox_id)
        result: CommandResult = sbx.commands.run(cmd=command)
        return TerminalInfo(stdout=result.stdout, stderr=result.stderr)

    class SandboxListTool(BaseTool):
        name: str = "list_sandbox_files"
        description: str = "List all files and directories in a sandbox directory. Provide sandbox_id and optional directory path."
        sandbox_service: "SandboxService" = Field(exclude=True)
        args_schema: Type[BaseModel] = ListToolInput

        def _run(self, sandbox_id: str, path: str) -> str:
            try:
                files: List[WriteInfo] = self.sandbox_service.list_files(
                    sandbox_id=sandbox_id, path=path
                )

                file_list = []
                for file in files:
                    file_list.append(f"{file.type}: {file.name} (path: {file.path})")

                files_text = "\n".join(file_list)
                return f"Successfully listed files from sandbox {sandbox_id}\nDirectory: {path}\nTotal items: {len(files)}\nFiles and directories:{files_text}"
            except Exception as e:
                return f"failed to list files in '{path}' from sandbox {sandbox_id}. error: {str(e)}"

    class SandboxReadTool(BaseTool):
        name: str = "read_sandbox_file"
        description: str = "Read a single file in the sandbox. To access the sandbox, the first parameter must be the sandbox_id and the second must be the path of the file."
        sandbox_service: "SandboxService" = Field(exclude=True)
        args_schema: Type[BaseModel] = ReadToolInput

        def _run(self, sandbox_id: str, path: str) -> str:
            try:
                file_content: str = self.sandbox_service.read_file(
                    sandbox_id=sandbox_id, path=path
                )
                return f"Successfully read file from sandbox {sandbox_id}\nPath: {path}\nContent: {file_content}"
            except Exception as e:
                return f"failed to read file '{path}' from sandbox {sandbox_id}. error: {str(e)}"

    class SandboxWriteTool(BaseTool):
        name: str = "write_sandbox_files"
        description: str = "Write one or more files to the sandbox. Provide sandbox_id and a list of files with their paths and content."
        sandbox_service: "SandboxService" = Field(exclude=True)
        args_schema: Type[BaseModel] = WriteToolInput

        def _run(self, sandbox_id: str, write_data: List[WriteEntry]) -> str:
            try:
                result: List[WriteInfo] = self.sandbox_service.write_files(
                    sandbox_id=sandbox_id, write_data=write_data
                )

                written_files = []
                for file in result:
                    written_files.append(
                        f"Successfully wrote {file.type}: {file.name} to: {file.path}"
                    )

                files_text = "\n".join(written_files)
                return f"Successfully wrote {len(result)} file(s) to sandbox {sandbox_id}\nFiles written:\n{files_text}"
            except Exception as e:
                return f"failed to write files to sandbox {sandbox_id}. error: {str(e)}"

    class SandboxCommandTool(BaseTool):
        name: str = "execute_sandbox_command"
        description: str = "Execute a terminal command in the sandbox. Provide sandbox_id and the command to run."
        sandbox_service: "SandboxService" = Field(exclude=True)
        args_schema: Type[BaseModel] = CommandToolInput

        def _run(self, sandbox_id: str, command: str) -> str:
            try:
                result: TerminalInfo = self.sandbox_service.execute_terminal_command(
                    sandbox_id=sandbox_id, command=command
                )

                output_parts = [f"Successfully executed in sandbox {sandbox_id}"]
                output_parts.append(f"Command: {command}")

                if result.stdout:
                    output_parts.append(f"\nStdout:\n{result.stdout}")

                if result.stderr:
                    output_parts.append(f"\nStderr:\n{result.stderr}")

                if not result.stdout and not result.stderr:
                    output_parts.append("\nCommand completed with no output.")

                return "\n".join(output_parts)
            except Exception as e:
                return f"failed to execute command '{command}' in sandbox {sandbox_id}. error: {str(e)}"

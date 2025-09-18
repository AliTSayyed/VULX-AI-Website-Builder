4. then create the related agent tools
5. properly prompt the agent to use the tools

create route and service models
properly log requests and services
properly handle exceptions
convert relevant functions to async

# Sandbox Service & Agent Tools Implementation TODO

## 4. Agent Tools

- [ ] Create `SandboxListTool` that calls the list files endpoint
- [ ] Create `SandboxReadTool` that calls the read file endpoint
- [ ] Create `SandboxWriteTool` that calls the write file endpoint
- [ ] Create `SandboxShellTool` that calls the execute command endpoint
- [ ] Update `SandboxService.get_agent_tools()` to return these tools

## 5. Integration

- [ ] Test FastAPI endpoints manually
- [ ] Test agent tools with OpenAI code agent
- [ ] Update `OpenAICodeAgentService` to use the new tools
- [ ] Test end-to-end flow from route to agent to sandbox

# GOLANG BACKEND STREAMING FLOW (SIMPLIFIED)

1. INITIAL REQUEST
   Frontend → Golang Handler (POST /api/messages)
   ├─ Handler starts Temporal Workflow
   ├─ Handler returns workflowID to Frontend
   └─ Frontend immediately calls SSE endpoint with workflowID

2. STREAMING CONNECTION
   Frontend → SSE Endpoint (GET /api/stream?workflow_id=xxx)
   ├─ SSE connection established
   └─ Waiting for chunks from Python via Golang

3. TEMPORAL WORKFLOW STEPS

   Step 1: Create Sandbox
   ├─ Workflow → Activity: CreateSandbox()
   ├─ Activity → E2B: Create NextJS sandbox
   ├─ E2B → Activity: Return sandboxID, URL
   └─ Activity → Workflow: Sandbox ready

   Step 2: Call AI Agent (WITH DIRECT STREAMING)
   ├─ Workflow → Activity: CallAIAgent(sandboxID, userMessage, workflowID)
   ├─ Activity → Python Service: POST /ai/agent (includes workflowID)
   │
   └─ STREAMING LOOP (BYPASSES TEMPORAL):
   ├─ Python Service runs LangChain agent
   ├─ Agent executes tools on E2B (write files, commands)
   ├─ For each chunk: Python → Golang REST: POST /api/stream/{workflowID}/chunk
   ├─ Golang → SSE Manager: Forward chunk directly
   ├─ SSE → Frontend: Stream chunk immediately
   └─ Repeat until agent complete
   │
   ├─ Python Service → Activity: Return final response
   └─ Activity → Workflow: AI processing complete

   Step 3: Save Results
   ├─ Workflow → Activity: SaveMessage(response, sandboxID)
   ├─ Activity → Database: Save complete response
   └─ Activity → Workflow: Saved

4. CLEANUP
   ├─ Workflow completes
   ├─ SSE connection closes
   └─ Frontend shows final result

KEY COMPONENTS:

- RPC Endpoint: Starts workflow, returns workflowID
- Chunk Endpoint: Receives chunks from Python, forwards directly to SSE
- SSE Endpoint: Streams real-time updates to frontend
- Temporal Client: Orchestrates main workflow (not streaming)
- Activities: CreateSandbox, CallAIAgent, SaveMessage

DIRECT STREAMING FLOW:
Python AI Service → Golang REST API → SSE Manager → Frontend
(Temporal handles orchestration, not streaming)

Frontend ←→ Golang (Connect RPC) ✓
Golang ←→ Python (REST) ✓
Python → Golang (HTTP chunks for streaming) ✓

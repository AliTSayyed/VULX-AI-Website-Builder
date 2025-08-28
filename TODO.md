# TODO

1. Code Generation & Storage
   AI generates code (in my application)
   Store that code in the database
   Code is associated with the chat/session
   AI should be aware of what the code files and content are

2. Code Execution
   Send the AI-generated code to the E2B sandbox (using the sandbox ID)
   E2B executes the code in the isolated environment
   E2B returns the results/output
   The same chat should have the same sandbox ID but the AI needs to show the right 'project' that the user asks.
   A user may have multiple 'projects' created in the same chat and the AI should be able to update the correct files when asked.
   Chache sandbox ids so we dont look them up on every chat message.

- E2B Provides
  Execution environment (not code generation)
  Results/output of the executed code
  URLs for web applications (if the code creates a web server/app)
  File system access (for file operations, data persistence within the sandbox)

3. Display to User
   Code: Show code from my own database
   Output/Results: Show this from E2B's response
   E2B can provide a URL to the running application

4. I Handle
   Code storage in the database
   Session management (mapping chats to sandbox IDs)
   UI for displaying both code and results
   Code versioning/history if needed
   Temporal Workflows for managing long conversations
   Langchain to control the output of AI and what models are available.
   Need to determine correct context for the AI on every call, may need intermediate AI calls to determine the best context for the actually code AI call.

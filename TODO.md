# TODO

### Infrastructure & DevOps

- [ ] **Create GitHub Actions CI/CD automation**
  - [ ] Set up local environment configuration
  - [ ] Set up production environment configuration
  - [ ] Configure automated builds and deployments
  - [ ] Set up environment-specific secrets management
  - (Note: Dev environment is low priority for this project)

### Golang API

#### Schema & Domain Layer

- [ ] **Create Messages schema, domain, service, and handlers**

  - [ ] Design messages table schema in PostgreSQL
  - [ ] Implement message domain entities
  - [ ] Create message service with business logic
  - [ ] Build message handlers (gRPC/REST endpoints)
  - [ ] **Implement SSE (Server-Sent Events) connection for streaming AI responses**
    - User should see AI response stream in real-time as it's generated
    - Need to handle SSE connection lifecycle and error cases
  - [ ] **Implement cursor-based pagination for get messages**
    - Use "before" cursor for infinite scroll up (loading older messages)
    - Return cursor with each response for next pagination call

- [ ] **Create Projects schema, domain, service, and handlers**

  - [ ] Design projects table schema in PostgreSQL
  - [ ] Implement project domain entities
  - [ ] Create project service with business logic
  - [ ] Build project handlers (gRPC/REST endpoints)
  - [ ] **Auto-update project timestamp when new message is sent**
    - Update `updated_at` field on the project whenever a message is added
  - [ ] **Implement get all projects with ordering by last updated**
    - Order by `updated_at` DESC (most recently updated first)
    - Add pagination with limit and offset for scroll-down loading
    - Return projects list for dashboard display

- [ ] **Create Codebase schema, domain, service, and handlers**
  - [ ] Design codebase table schema with version history support
  - [ ] Each save should create a new version entry (immutable history)
  - [ ] Link versions to messages/timestamps for user to see iterations
  - [ ] Implement codebase domain entities
  - [ ] Create codebase service for version management
  - [ ] Build codebase handlers for retrieving specific versions
  - [ ] Implement endpoint to download codebase version as ZIP file

#### Sandbox Management

- [ ] **Implement Redis caching for sandbox IDs**

  - [ ] Use project ID as Redis key
  - [ ] Store sandbox ID as value with TTL
  - [ ] Set TTL to 30-60 minutes (decide between 30min or 1hr)

- [ ] **Build sandbox lifecycle management**
  - [ ] **Pre-expiration refresh logic**
    - When user sends message, check if sandbox TTL < 5 minutes remaining
    - If expiring soon, create new sandbox and swap sandbox ID in Redis
    - Update project with new sandbox ID before old one expires
    - This provides seamless experience for active users
  - [ ] **Cold start handling**
    - When user opens project after long period (sandbox expired)
    - Create new sandbox and inform user to wait for activation
    - Display loading state during sandbox creation (normal UX, same as Bolt)

#### Workflow Orchestration

- [ ] **Determine Temporal workflow usage**
  - [ ] Decide if Temporal is only for POST message requests or broader use cases
  - [ ] Implement workflow for message processing:
    - Credit deduction
    - Sandbox creation/retrieval
    - AI service call
    - Codebase version creation
    - Response streaming
  - [ ] Handle retries and error cases in workflow
  - [ ] Define workflow timeout and compensation logic

### Frontend (Next.js)

#### Landing Page (Logged Out)

- [ ] **Build home page for unauthenticated users**
  - [ ] Create hero section with animated chat textbox
    - Implement typing animation that writes text then deletes it
    - Cycle through multiple example prompts (e.g., "Create a login form", "Build a dashboard", "Make a landing page")
    - Make textbox interactive
  - [ ] **Implement login modal trigger on textbox interaction**
    - Show login modal when user hovers over or clicks textbox
    - Modal should have OAuth Google login option
  - [ ] **Add navigation bar**
    - Login button on top right
    - Additional navigation items (features, pricing, docs, etc.) to fill space
    - Make navbar feel complete and professional

#### Dashboard (Logged In)

- [ ] **Build main dashboard layout**

  - [ ] **Left sidebar**

    - "New Project" button at top (navigates to fresh chat screen)
    - "Recents" tab showing recent projects ordered by last updated
    - "Projects" tab showing all projects with scroll-down pagination
    - Each project item shows title and last updated timestamp

  - [ ] **Center panel - Chat interface**

    - Message list with infinite scroll up (load older messages)
    - Message input textbox at bottom
    - Display AI responses with streaming (SSE connection)
    - Show user messages and AI responses in conversation format

  - [ ] **Right panel - Sandbox view**
    - Default view: Live preview of sandbox (iframe or similar)
    - Toggle button to switch between "Preview" and "Code" view
    - **Preview view**: Always shows latest version only
      - Display notice: "Preview shows latest version only"
    - **Code view**: Show code files with syntax highlighting
      - Read-only code display (user cannot edit)
      - Version selector dropdown to switch between code versions
      - Each version switch makes separate API request (don't preload all versions)
      - Display current version number and timestamp
    - **Download button**: Download current version as ZIP file

#### State Management

- [ ] Set up React Query for data fetching and caching
- [ ] Implement optimistic updates for message sending
- [ ] Handle SSE connection state (connecting, open, error, closed)
- [ ] Manage project selection and active project state

### AI-Service (FastAPI)

#### Model Consistency & Reliability

- [ ] **Ensure all 3 models work consistently**

  - [ ] Test OpenAI code agent thoroughly
  - [ ] Test Google Gemini code agent thoroughly
  - [ ] Test Anthropic Claude code agent thoroughly
  - [ ] Normalize responses across different models
  - [ ] Handle model-specific quirks and limitations
  - [ ] Need to create message length limit for sending requests to llms

- [ ] **Error handling and retry logic**
  - [ ] Define error response format for Golang API
  - [ ] Implement retry logic for transient failures
  - [ ] **Coordinate with Temporal workflow in Golang API**
    - Determine what errors should trigger Temporal retries
    - Define retry policies (max attempts, backoff strategy)
    - Handle permanent failures vs. temporary failures
  - [ ] Add circuit breaker pattern for external API calls
  - [ ] Implement timeout handling for long-running agent operations
  - [ ] Log all errors for debugging and monitoring

## Backlog / Future Tasks

- [ ] Add tests for critical paths (auth, message streaming, sandbox lifecycle)
- [ ] Set up monitoring and alerting for production
- [ ] Implement rate limiting per user tier
- [ ] Add analytics for tracking usage patterns
- [ ] Optimize database queries with proper indexes
- [ ] Consider Redis Cluster for high availability
- [ ] Add WebSocket fallback for SSE if needed
- [ ] Implement collaborative features (share projects)

## Questions to Resolve

- **Sandbox TTL**: Should sandboxes last 30 minutes or 1 hour?
- **Temporal scope**: Is Temporal only for message POST requests, or should it orchestrate other workflows too?
- **Version history limits**: Should we limit how many versions we keep per project? Or keep all versions forever?
- **Error UX**: How should we display errors to users when AI service fails? Retry automatically or let user retry manually?
- **Preview refresh**: Should preview auto-refresh when new code version is created, or require manual refresh?

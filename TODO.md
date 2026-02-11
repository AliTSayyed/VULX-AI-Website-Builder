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
  - [ ] **Store file structure and contents for WebContainer**
    - Store complete file tree with paths and content
    - Support efficient retrieval for WebContainer boot
    - Enable diff generation between versions

#### WebContainer State Management

- [ ] **Implement Redis caching for WebContainer state**

  - [ ] Use project ID as Redis key
  - [ ] Store serialized file system state as value with TTL
  - [ ] Set TTL to 30-60 minutes (decide between 30min or 1hr)
  - [ ] Cache file structure for quick WebContainer initialization

- [ ] **Build WebContainer lifecycle management**
  - [ ] **Handle state persistence**
    - Save WebContainer file system state to Redis on updates
    - Load cached state when user returns to project
    - Invalidate cache when new code version is created
  - [ ] **Cold start handling**
    - When user opens project after cache expiration
    - Load latest codebase version from PostgreSQL
    - Initialize WebContainer with fresh file system
    - Display loading state during initialization

#### Workflow Orchestration

- [ ] **Determine Temporal workflow usage**
  - [ ] Decide if Temporal is only for POST message requests or broader use cases
  - [ ] Implement workflow for message processing:
    - Credit deduction
    - AI service call
    - Codebase version creation
    - WebContainer state update
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

  - [ ] **Right panel - WebContainer IDE**
    - **Default view: Live preview**
      - Render WebContainer preview in iframe
      - Auto-refresh when new code version is generated
      - Display loading state during WebContainer boot
    - **Toggle between "Preview" and "Code" views**
    - **Code view features:**
      - File tree explorer (collapsible folders)
      - Monaco editor for code editing (syntax highlighting, autocomplete)
      - User can edit files and see changes in real-time
      - Save button to persist manual edits as new version
      - Version selector dropdown to switch between versions
        - Each version switch loads different file structure from API
        - Don't preload all versions (fetch on demand)
      - Display current version number and timestamp
    - **Download button**: Download current version as ZIP file
    - **WebContainer integration:**
      - Initialize WebContainer on project load
      - Mount file system from API/cache
      - Handle hot module reload for edits
      - Terminal output display (optional feature)

#### State Management

- [ ] Set up React Query for data fetching and caching
- [ ] Implement optimistic updates for message sending
- [ ] Handle SSE connection state (connecting, open, error, closed)
- [ ] Manage project selection and active project state
- [ ] **WebContainer state management**
  - Track WebContainer boot status
  - Sync file system changes with backend
  - Handle WebContainer errors and recovery

### AI-Service (FastAPI)

#### Model Consistency & Reliability

- [ ] **Ensure all 3 models work consistently**

  - [ ] Test OpenAI code agent thoroughly
  - [ ] Test Google Gemini code agent thoroughly
  - [ ] Test Anthropic Claude code agent thoroughly
  - [ ] Normalize responses across different models
  - [ ] Handle model-specific quirks and limitations
  - [ ] Create message length limit for sending requests to LLMs
  - [ ] **Return structured file outputs**
    - Return complete file tree structure (paths + content)
    - Ensure consistent format across all AI models
    - Include metadata (file types, line counts, etc.)

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

- [ ] Add tests for critical paths (auth, message streaming, WebContainer lifecycle)
- [ ] Set up monitoring and alerting for production
- [ ] Implement rate limiting per user tier
- [ ] Add analytics for tracking usage patterns
- [ ] Optimize database queries with proper indexes
- [ ] Consider Redis Cluster for high availability
- [ ] Add WebSocket fallback for SSE if needed
- [ ] Implement collaborative features (share projects)
- [ ] Add terminal emulator in WebContainer view
- [ ] Support npm install and package management in WebContainer
- [ ] Implement file upload for custom assets
- [ ] Add syntax error detection before WebContainer execution

## Questions to Resolve

- **WebContainer state TTL**: Should cached state last 30 minutes or 1 hour?
- **Temporal scope**: Is Temporal only for message POST requests, or should it orchestrate other workflows too?
- **Version history limits**: Should we limit how many versions we keep per project? Or keep all versions forever?
- **Error UX**: How should we display errors to users when AI service fails? Retry automatically or let user retry manually?
- **Manual edits**: When user edits code in WebContainer, should we auto-save versions or require explicit save action?
- **WebContainer persistence**: Should user edits in WebContainer persist if they refresh the page, or only saved versions persist?
- **Preview refresh**: Should preview auto-refresh when new code version is created, or require manual refresh?

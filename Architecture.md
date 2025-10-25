# Architecture

## Overview

This project is a SaaS platform that enables users to generate frontend code through AI interaction, similar to Bolt and Lovable. Users can ask an AI to create frontend components and receive both the generated code and a live visual preview of that frontend. The platform supports multiple AI models (OpenAI, Google Gemini, and Anthropic Claude Sonnet), allowing users to choose their preferred model for code generation. Each query consumes credits from the user's account, providing a usage-based monetization model.

The core value proposition is giving users the ability to rapidly prototype and generate frontend code through natural language instructions, with immediate visual feedback of the results.

## High-Level Architecture

The application is built as a **monorepo containing multiple microservices**, each responsible for a distinct part of the system. This architecture allows for independent scaling, deployment, and development of each service while maintaining a cohesive codebase.

### Microservices

1. **Golang API** - The main backend API that handles business logic, authentication, user management, and orchestration
2. **Next.js Frontend** - The web interface where users interact with the AI and view generated code
3. **FastAPI AI-Service** - A Python-based service dedicated to AI model interactions and sandbox management
4. **Proto Folder** - Contains Protocol Buffer definitions for RPC communication between frontend and backend

### Communication Patterns

- **Frontend ↔ Backend**: gRPC + REST (using Connect RPC with Vanguard for dual protocol support)
- **Backend ↔ AI-Service**: REST-based HTTP communication
- **Protocol Buffers**: Generated using Buf for type-safe communication contracts

### Infrastructure Services (Docker Compose)

The entire application stack runs via Docker Compose with the following services:

- **PostgreSQL** - Primary database for application data and Temporal workflow storage
- **Redis** - Session storage for login tokens and temporary sandbox ID caching
- **Temporal** - Workflow orchestration service for managing long-running processes
- **Caddy** - Reverse proxy providing HTTPS certificates, enabling secure JWT cookie storage and serving OpenAI documentation
- **AI-Service** - FastAPI container for AI interactions
- **Frontend** - Next.js application container
- **Backend API** - Golang service container

Each microservice has its own Dockerfile for containerization and independent deployment.

## Golang API Architecture

The backend follows **Hexagonal Architecture (Ports and Adapters) with Domain-Driven Design (DDD)** principles. This architecture pattern keeps business logic isolated from external dependencies, making the codebase maintainable, testable, and flexible to change.

### Hexagonal DDD Structure

**Core Concept**: The business logic sits at the center (the "hexagon") and is completely isolated from external concerns. All external interactions happen through well-defined interfaces (ports), which are implemented by adapters.

**Layers (from inside out)**:

1. **Domain Layer (The Core)**

   - Contains pure business logic and rules
   - Defines entities (core business objects like User, Project, Credit, etc.)
   - No dependencies on frameworks, databases, or external services
   - This is the "what" of the application - what the business does

2. **Application Services Layer**

   - Orchestrates use cases and workflows
   - Coordinates between domain entities and ports
   - Contains the application's use case logic (e.g., "create a new AI generation request")
   - Depends only on domain and port interfaces

3. **Ports (Interfaces)**

   - Define contracts for external dependencies
   - **Inbound Ports**: Interfaces for incoming requests (e.g., HTTP handlers call these)
   - **Outbound Ports**: Interfaces for outgoing operations (e.g., database, external APIs)
   - These are the "promises" that adapters must fulfill

4. **Adapters**
   - Concrete implementations of ports
   - **Inbound Adapters (Handlers/Controllers)**:
     - API layer that receives requests from frontend
     - Uses endpoints declared in proto files
     - Calls application services to execute business logic
   - **Outbound Adapters**:
     - PostgreSQL adapter for database operations
     - OAuth adapter for Google authentication
     - Temporal adapter for workflow service
     - AI-Service adapter for REST calls to FastAPI
     - Redis adapter for caching

### Dependency Injection

All dependencies are wired together at application startup using dependency injection. This means:

- Services receive their dependencies through constructors
- No service creates its own dependencies
- Easy to swap implementations (e.g., mock database for testing)
- Clear dependency graph that prevents circular dependencies

### gRPC + REST Compatibility

We use **Connect RPC with Vanguard** to achieve dual-protocol support:

- Single handler implementation serves both gRPC and REST clients
- Vanguard automatically generates REST endpoints from gRPC service definitions
- This is currently the best approach for Go services that need both protocols
- Frontend can use either protocol depending on the context

**Why this matters**: Traditional gRPC requires gRPC-web for browsers, which has limitations. Connect + Vanguard gives us native browser support via REST while keeping efficient gRPC for backend-to-backend calls.

## Authentication & Authorization

### JWT-Based Authentication (Hybrid Approach)

We use a hybrid JWT approach optimized for applications with 100+ users where convenience is prioritized over maximum security (appropriate for most SaaS applications).

**Token Lifecycle**:

- JWT tokens are valid for 7 days
- Stored in httpOnly cookies (protected from JavaScript access, immune to XSS)
- Automatically attached to requests by the browser
- No refresh token needed - the long-lived token auto-refreshes transparently

### Authentication Flow

**Login/Signup**:

1. Frontend sends credentials to server
2. Server validates credentials
3. Server creates 7-day JWT
4. Server returns JWT in httpOnly cookie via Set-Cookie header
5. Browser automatically stores cookie

**Normal API Request**:

1. Frontend makes API call
2. Browser automatically attaches JWT cookie
3. Server validates JWT
4. Server processes request and returns response

**Automatic Token Refresh** (happens invisibly):

1. Frontend makes any API call
2. Server validates JWT and checks expiration
3. If token has less than 42 hours remaining:
   - Server processes the request normally
   - Server generates a NEW 7-day JWT
   - Server includes new token in response Set-Cookie header
   - Browser automatically updates cookie
4. User never knows refresh happened

**Token Expiration** (>7 days inactive):

1. Frontend makes API call with expired token
2. Server returns 401 Unauthenticated
3. Frontend interceptor detects 401
4. Frontend clears local cache
5. Frontend redirects to login page

**Logout**:

1. User clicks logout
2. Frontend calls logout endpoint
3. Server adds hashed token to blacklist database (stored until natural expiry)
4. Server clears cookie via Set-Cookie with past expiry date
5. Browser deletes cookie
6. Frontend clears cache and redirects to login

**Key Insight**: The refresh happens on the backend during normal requests, not through a separate frontend flow. The user experiences seamless authentication without ever seeing "token refreshing" messages.

### OAuth Integration (Google)

Currently supporting Google OAuth for passwordless authentication. Password-based auth can be added in the future.

**OAuth Flow (Step-by-Step)**:

1. **Homepage (Unauthenticated)**

   - User sees "Login with Google" button

2. **Initiate OAuth**

   - User clicks "Login with Google"
   - Frontend calls `beginAccountAuth()` RPC endpoint
   - Backend generates Google OAuth URL with state parameter
   - Backend returns OAuth URL to frontend

3. **Redirect to Google**

   - Frontend executes `window.location.href = Google OAuth URL`
   - Browser navigates away from our app to Google

4. **Google Consent Screen**

   - User sees requested permissions
   - **Success Path**: User clicks "Allow"
   - **Error Path**: User clicks "Deny" or closes window

5. **Google Redirects Back**

   - **Success**: `/auth/callback?code=ABC123&state=xyz`
   - **Error**: `/auth/callback?error=access_denied`

6. **AuthCallback Component Mounts**

   - React component loads at callback URL
   - `useEffect` runs automatically on mount

7. **Process Callback**

   - useEffect examines URL parameters
   - **Has code**: Call `finishAccountAuth()` RPC
   - **Has error or no code**: Redirect to `/login`

8. **Backend Processing** (success path only):

   - Validate state parameter (CSRF protection)
   - Exchange authorization code for Google access token
   - Use access token to fetch user profile from Google
   - Check if user exists in database:
     - If exists: Update last login timestamp
     - If new: Create user record
   - Generate 7-day JWT for our application
   - Return JWT + user profile to frontend

9. **Frontend Handles Response**

   - Browser stores JWT cookie automatically
   - Update React Query cache with user data
   - Navigate to `/dashboard`

10. **Final Destination**
    - **Success**: User lands on authenticated dashboard
    - **Failure**: User returns to login page to try again

**Why OAuth**: Eliminates password management, improves security, reduces friction in signup flow, and leverages trusted identity providers.

## AI-Service (FastAPI)

The AI-Service is a Python-based FastAPI server that handles all AI model interactions and manages code generation sandboxes. This service exists separately from the Golang API because E2B (the sandbox provider) only has a Python SDK, not a Golang SDK.

### Why FastAPI?

- Python is required for E2B sandbox SDK
- LangChain (used for AI agents) has the most mature Python implementation
- FastAPI provides excellent performance for Python web services
- Easier to manage AI/ML dependencies in Python ecosystem

### Endpoints

The service exposes separate endpoints for three AI providers:

1. **OpenAI Endpoints**

   - `/openai/query` - General AI questions
   - `/openai/code-agent` - Code generation with sandbox interaction

2. **Google Gemini Endpoints**

   - `/gemini/query` - General AI questions
   - `/gemini/code-agent` - Code generation with sandbox interaction

3. **Anthropic Claude Endpoints**
   - `/claude/query` - General AI questions
   - `/claude/code-agent` - Code generation with sandbox interaction

**Query Endpoints**: Accept any general question and return AI response (no sandbox interaction)

**Code Agent Endpoints**: Accept a sandbox ID and use AI agents with tools to read, write, and execute code in that sandbox

### Sandbox Management (E2B)

**Sandbox Creation**:

- `/sandbox/create` endpoint creates new E2B sandboxes
- Each sandbox requires a template ID
- Template is a pre-configured Next.js environment built with Docker
- Template Dockerfile is pushed to E2B account and receives a template ID
- This template ID is specified when creating new sandboxes

**Important**: To use a different sandbox environment (e.g., React instead of Next.js), you must:

1. Create a new Dockerfile with that environment
2. Push it to E2B as a template
3. Use the new template ID when creating sandboxes

**Sandbox Template**: Currently using a Next.js template that includes:

- Node.js runtime
- Next.js framework pre-installed
- Common dependencies
- Development server configuration

### LangChain AI Agents

We use LangChain to build AI agents that can interact with sandboxes programmatically.

**Agent Architecture**:

- Each AI model (OpenAI, Gemini, Claude) has its own agent configuration
- Agents are given tools to interact with sandboxes
- When given a sandbox ID, agents can:
  - **Read files**: View existing code in the sandbox
  - **Write files**: Create or modify files
  - **List files**: See directory structure
  - **Execute commands**: Run shell commands (npm install, npm run dev, etc.)

**Sandbox Service**:

- Encapsulates all E2B API interactions
- Provides clean interface for sandbox operations
- Methods include: connect, read_file, write_file, list_files, execute_command, disconnect

**Tools for Agents**:

- The sandbox service methods are converted into LangChain tools
- Each tool has a clear description for the AI to understand when to use it
- Tools are registered with the agent at initialization
- Agent autonomously decides which tools to use based on user request

### Callback System for Agent Tracking

**Problem**: AI agents don't always accurately report what they did

**Solution**: Custom callback service that logs all agent actions

**How it works**:

- Callback hooks intercept every tool invocation
- Logs are created before and after each tool call
- Records: tool name, input parameters, output, timestamp, success/failure
- Provides ground truth of what the agent actually did vs. what it claims

**What we return to users**:

- Only the successful write operations
- List of files created/modified
- Generated code content
- NOT the agent's narrative of what it did (unreliable)

**Why this matters**: We can audit agent behavior, debug issues, and provide accurate feedback to users about what changed in their sandbox.

### Dependency Injection Pattern

Following the same pattern as the Golang API:

- All client objects (OpenAI, Gemini, Claude, E2B) are instantiated once at startup
- Clients are passed to services via dependency injection
- Prevents creating new client connections for each request
- Improves performance and resource management
- Makes testing easier (can inject mock clients)

### Communication with Golang API

- Golang API makes REST calls to AI-Service endpoints
- JSON request/response format
- Golang API adapter handles serialization and error handling
- Sandbox IDs are passed from Golang API to AI-Service for agent operations

## Frontend (Next.js)

The frontend is built with Next.js and uses Shadcn UI components for a modern, accessible user interface.

**Current State**: Basic connection to backend established, but UI is still in development phase.

**Key Technologies**:

- **Next.js**: React framework with server-side rendering and routing
- **Shadcn UI**: High-quality, customizable component library built on Radix UI
- **Connect RPC**: Type-safe client for calling backend services
- **React Query**: For data fetching, caching, and state management

**Planned Features** (to be built out):

- Dashboard for managing AI generations
- Code editor with live preview
- Model selection interface
- Credit usage tracking
- Project management
- Authentication UI (login, signup, OAuth callback)

**Authentication Integration**:

- JWT cookies automatically attached to requests
- React Query cache stores user data
- Protected routes redirect unauthenticated users
- OAuth callback handler at `/auth/callback`

## Proto

The proto folder contains all Protocol Buffer definitions that define the contract between frontend and backend.

**Purpose**:

- Defines message structures for requests and responses
- Ensures type safety across frontend and backend
- Single source of truth for API contracts
- Automatically generates TypeScript and Go code

**What's Defined Here**:

- User authentication messages (login, signup, OAuth)
- AI query and code generation requests/responses
- Sandbox creation and management messages
- Credit system messages
- Error response structures

**Build Process**:

- Use Buf to generate code from proto files
- Generated TypeScript code for frontend
- Generated Go code for backend
- Buf handles versioning and breaking change detection

**Benefits of Protocol Buffers**:

- Strong typing prevents runtime errors
- Smaller payload size than JSON
- Built-in backward compatibility rules
- Clear documentation through structured definitions
- Works seamlessly with gRPC and REST (via Connect + Vanguard)

## Key Architectural Decisions

### Why Monorepo?

- All services in one repository for easier development
- Shared proto definitions between services
- Simplified dependency management
- Atomic commits across multiple services
- Easier to maintain consistency

### Why Microservices?

- Independent scaling (AI-Service can scale separately from API)
- Language flexibility (Python for AI, Go for performance)
- Isolated failures (AI-Service crash doesn't bring down auth)
- Easier to maintain bounded contexts

### Why Hexagonal Architecture?

- Business logic is protected and testable
- Easy to swap implementations (change database, change AI provider)
- Clear separation of concerns
- Prevents framework lock-in

### Why JWT in Cookies?

- Protection against XSS attacks (httpOnly)
- Automatic attachment to requests (no manual headers)
- Works across subdomains
- Simple revocation via blacklist

### Why Separate AI-Service?

- E2B only has Python SDK
- LangChain is most mature in Python
- Isolates heavy AI dependencies from main API
- Can scale AI processing independently

### Why Temporal?

- Complex workflows for code generation (multi-step processes)
- Built-in retry logic and error handling
- Visibility into long-running operations
- Can resume workflows after crashes

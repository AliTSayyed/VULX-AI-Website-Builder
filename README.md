# AI Frontend Code Generator SaaS

A SaaS platform that enables users to generate frontend code through AI interaction. Users can ask AI to create frontend components and receive both the generated code and a live visual preview, similar to Bolt and Lovable.

## Features (In Progress)

- **Multiple AI Models**: Choose between OpenAI, Google Gemini, or Anthropic Claude Sonnet
- **Live Sandbox Preview**: See your generated code running in real-time
- **Version History**: Track all iterations of your code generations
- **OAuth Authentication**: Secure OAuth login with JWT sessions
- **Credit-Based System**: Pay-per-use model for AI generations
- **Code Export**: Download any version of your generated code as a ZIP file

## Tech Stack

### Backend

- **Golang** - Main API with hexagonal DDD architecture
- **FastAPI (Python)** - AI service for LangChain agents and E2B sandboxes
- **PostgreSQL** - Primary database
- **Redis** - Session storage and sandbox caching
- **Temporal** - Workflow orchestration
- **Protocol Buffers** - Type-safe communication contracts
- **Connect RPC + Vanguard** - Dual gRPC/REST protocol support

### Frontend

- **Next.js** - React framework
- **Shadcn UI** - Component library
- **React Query** - Data fetching and state management

### Infrastructure

- **Docker & Docker Compose** - Containerization
- **Caddy** - Reverse proxy with HTTPS
- **E2B** - Sandboxed code execution environments
- **Buf** - Protocol buffer code generation

## Architecture

This project follows hexagonal (ports and adapters) architecture with domain-driven design principles. See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed architectural documentation.

## Prerequisites

Install the following tools:

- **Git** - Version control
- **Golang** - `^1.24.2`
- **Docker** - Container runtime
- **Docker Compose** - Multi-container orchestration
- **Make** - Build automation
- **Node.js** - `^18.x` (for frontend)
- **Python** - `^3.11` (for AI service)
- **Buf CLI** - For generating protocol buffers

## Getting Started

### 1. Clone the repository

```bash
git clone <repository-url>
cd <project-directory>
```

### 2. Set up environment variables

Secrets live in a sibling directory, not in this repo: `../AI-Website-Builder-Secrets/`. Compose
reads `.api-env` and `.ai-service-env` from there — nothing starts without it. Required keys are
enumerated in `api/.env.example` and `ai-service/.env.example`.

`.api-env` must additionally set the three URLs the two-host local setup depends on:

```
APP_URL=https://local.app.vulx.ai
API_URL=https://local.api.vulx.ai
REDIRECT_URL=https://local.app.vulx.ai/auth/callback
```

`APP_URL`/`API_URL` are the CORS allowlist (`AllowCredentials: true` against explicit origins —
never a wildcard). `REDIRECT_URL` must be byte-identical to the redirect URI registered in Google
Cloud Console (step 5 below), or login fails with `redirect_uri_mismatch`.

Also create `app/.env.local` (gitignored, not committed):

```
NEXT_PUBLIC_API_URL=https://local.api.vulx.ai
```

This is inlined into the frontend bundle at build time; changing it later needs
`docker compose restart app`, not a rebuild.

### 3. `/etc/hosts`

The stack serves two hosts through Caddy, both required for login to work (the session cookie needs
a real `.vulx.ai`-domain HTTPS origin, and Google will not accept a `localhost`-free HTTP redirect
URI):

```
127.0.0.1 local.api.vulx.ai
127.0.0.1 local.app.vulx.ai
```

### 4. Generate Protocol Buffers

```bash
make gen
```

### 5. Google Cloud Console

Add to the OAuth client's **Authorized redirect URIs**:

```
https://local.app.vulx.ai/auth/callback
```

Must exactly match `REDIRECT_URL` above — a trailing-slash mismatch is enough to break it. No
**Authorized JavaScript origins** entry is needed; the app never loads Google's JS SDK.

### 6. Start all services

```bash
make
```

This will start:

- PostgreSQL database
- Redis cache
- Temporal server
- Caddy reverse proxy
- Golang API
- FastAPI AI service
- Next.js frontend

### 7. Trust Caddy's local CA

Without this, the browser shows a certificate interstitial (or a Google OAuth redirect fails
silently on it). Once the stack is up:

```bash
make trust
```

This copies Caddy's root cert out of the container and trusts it in the macOS system keychain
(prompts for your password). `make nuke` drops the Caddy data volume and regenerates the CA, so
re-run `make trust` after any nuke. `make trust-rm` removes a previously trusted Caddy CA from the
keychain, for cleanup.

### 8. Access the application

- **Frontend**: https://local.app.vulx.ai
- **API**: https://local.api.vulx.ai
- **Temporal UI**: http://localhost:8081

## License

Proprietary - All Rights Reserved

Copyright (c) 2025. This code is private and confidential.

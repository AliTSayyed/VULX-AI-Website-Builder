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

Copy the example env files and fill in your credentials:

```bash
cp .env.example .env
# Edit .env with your API keys and configuration
```

### 3. Generate Protocol Buffers

```bash
make gen
```

### 4. Start all services

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

### 5. Access the application

- **Frontend**: https://localhost (via Caddy, add local.vulx.ai to etc/hosts on 127.0.0.1)
- **API**: https://localhost/api
- **Temporal UI**: http://localhost:8080

## License

Proprietary - All Rights Reserved

Copyright (c) 2025. This code is private and confidential.

# Intelligent Tasking Engine

An enterprise-grade, multi-regional Agentic Tasking Engine built on Go, React, and Bazel 8, powered by Google's Model Context Protocol (MCP) and Gemini AI.

---

## Project Overview

The **Intelligent Tasking Engine** orchestrates operational retail and task workflows, combining static rule-based scheduling with dynamic AI agent execution.

- **Administrative MDM & Schema CRUD:** Fully managed entities including Users, Roles, Locations, Task Templates, Certifications, and SOPs.
- **AI-Assisted Task Execution:** Real-time query resolution via vector search against localized SOP context windows (`pgvector` on AlloyDB).
- **Human-in-the-Loop Compliance:** Explicit override mechanisms that trigger transactional audit ledgers using custom GORM Hooks.
- **Model Context Protocol (MCP):** Fully loaded agent definitions loaded from the `/pkg/agents` module, exposed hermetically over JSON-RPC.

For a deep dive into technical boundaries, database schemas, and spatial mapping domains, read the [Architectural Design Document](docs/architecture.md).

---

## Repository Architecture & Directory Layout

The workspace follows the standard Go project layout (`/cmd`, `/internal`, `/pkg`, `/api`) combined with Bazel 8 rules to build both backend services and React client interfaces.

| Directory | Description |
| :--- | :--- |
| [`/cmd/server`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/cmd/server) | Entry point for the Gin API Server and JSON-RPC MCP listeners. |
| [`/pkg/persistence`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/persistence) | GORM configuration, GCS connectivity, pgvector client hooks, and transactional DB instances. |
| [`/pkg/model`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/model) | Core domain entities and schemas mapped via GORM tags. |
| [`/pkg/server`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/server) | Gin web router, custom JWT authentication validation middlewares, and route registration. |
| [`/pkg/service`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/service) | Clean, isolated service layers encapsulating operational task flows and peer-to-peer handshakes. |
| [`/pkg/agents`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/agents) | Gemini Model Context Protocol (MCP) agent tool implementations. |
| [`/web/console`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/web/console) | React-based Admin Operations Dashboard built using pnpm, Vite, and TypeScript. |
| [`/web/render-test`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/web/render-test) | React-based rendering validation workbench. |
| [`/docs`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs) | Core architecture diagrams, specs, and schemas. |

---

## Build & Development Workflows

The entire repository is driven via **Bazel 8** for fast, hermetic builds.

### 1. Local Monorepo Setup
Initialize and resolve all Bazel modules and Node dependencies:
```bash
bazel mod tidy
```

### 2. Starting the Concurrent Development Environment
To start both frontend applications and the Go backend API server simultaneously in a parallel dev session, run the `dev_server` target:
```bash
bazel run //:dev_server
```
- **Go API Backend:** Binds to `:8080`
- **Admin Console SPA:** Binds to `:5173`
- **Render Test Workbench:** Binds to `:5174` (dynamically negotiated)

### 3. Independent Component Builds & Tests

#### Build & Test the Backend Server
```bash
# Build server binary
bazel build //cmd/server

# Run server tests
bazel test //pkg/...
```

#### Build React Frontend Applications
```bash
# Build the Admin Console
bazel build //web/console:build

# Build the Render Test Workbench
bazel build //web/render-test:build
```

---

## Technical & Styling Guidelines

- **Go Development:** Private business logic is strictly kept under `/internal` or private submodules. Multi-stage containerization packages Go binaries inside secure distroless containers.
- **TypeScript/React Styling:** Functional arrow syntax components wrap Material 3 `ThemeProvider` definitions. Tailwind CSS is reserved for localized layouts.
- **Python Tooling:** Any external automation or helper scripts must be executed exclusively via `uv` under the designated python environment.

# Enterprise Task Engine

An enterprise-grade, multi-regional Task Orchestration Engine built on Go, React, and Bazel 8, powered by Google's Model Context Protocol (MCP) and Gemini AI.

---

## Project Overview

The **Enterprise Task Engine** orchestrates operational retail and task workflows, combining static rule-based scheduling with dynamic AI agent execution.

- **Administrative MDM & Schema CRUD:** Fully managed entities including Users, Roles, Locations, Task Templates, Certifications, and SOPs.
- **AI-Assisted Task Execution:** Real-time query resolution via vector search against localized SOP context windows (`pgvector` on AlloyDB).
- **Human-in-the-Loop Compliance:** Explicit override mechanisms that trigger transactional audit ledgers using custom GORM Hooks.
- **Model Context Protocol (MCP):** Fully loaded agent definitions loaded from the `/pkg/agents` module, exposed hermetically over JSON-RPC.

For a deep dive into technical boundaries, database schemas, and spatial mapping domains, read the **[Architectural Design Document](docs/content/docs/architecture.md)**, **[Event-Driven Operations Specification](docs/content/docs/events.md)**, and the **[Background Job Scheduler Daemon Guide](docs/content/docs/scheduler.md)**.

---

## Repository Architecture & Directory Layout

The workspace follows the standard Go project layout (`/cmd`, `/internal`, `/pkg`, `/api`) combined with Bazel 8 rules to build both backend services and React client interfaces.

| Directory | Description |
| :--- | :--- |
| [`/cmd/server.go`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/cmd/server.go) | Entry point for the Gin API Server and JSON-RPC MCP listeners. |
| [`/pkg/persistence`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/persistence) | GORM configuration, GCS connectivity, pgvector client hooks, and transactional DB instances. |
| [`/pkg/model`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/model) | Core domain entities and schemas mapped via GORM tags. |
| [`/pkg/api`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/api) | Gin web router, custom JWT authentication validation middlewares, and route registration. |
| [`/pkg/service`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/service) | Clean, isolated service layers encapsulating operational task flows and peer-to-peer handshakes. |
| [`/pkg/agents`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/agents) | Gemini Model Context Protocol (MCP) agent tool implementations. |
| [`/web/console`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/web/console) | React-based Admin Operations Dashboard built using pnpm, Vite, and TypeScript. |
| [`/web/render-test`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/web/render-test) | React-based rendering validation workbench. |
| [`/apps/gtasks`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/apps/gtasks) | Native Android associate portal app (Kotlin/Compose). |
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
To start both frontend applications and the Go backend API server simultaneously in a parallel dev session, run the `dev_portal` target:
```bash
bazel run //:dev_portal
```

#### Starting Only the Go API Backend & MCP Server
If you only need to run the Go API backend and JSON-RPC MCP server (for example, when testing and developing with the native Android gateway in `apps/gtasks`), run the isolated `dev_server` target:
```bash
bazel run //:dev_server
```
- **Go API Backend:** Binds to `:8080`
- **Admin Console SPA:** Binds to `:5173`
- **Render Test Workbench:** Binds to `:5174` (dynamically negotiated)

#### Live Hot-Swapping with `ibazel`
To enable dynamic in-browser Hot Module Replacement (HMR) on local file edits without manually restarting targets, run the dev cluster using `ibazel` (the interactive Bazel watcher):

* **Option A (Zero-Install via npx):**
  ```bash
  npx @bazel/ibazel run //:dev_server
  ```
* **Option B (Global CLI installation):**
  ```bash
  npm install -g @bazel/ibazel
  ibazel run //:dev_server
  ```

The console frontend target has been pre-configured with `tags = ["ibazel_notify_changes"]` to permit direct sandbox notification and instant Vite HMR sweeps.

### 3. Serving the Documentation Site
To build and spin up the local Hugo server with live hot-reloading to browse the documentation locally under a rich visual style, run:
```bash
bazel run //docs:serve
```
- **Documentation Server:** Serves at `http://localhost:1313/`

### 4. Native Android Handset App (GTasks)
The native Android app is located in `/apps/gtasks`. It communicates with the Go API Gateway and serves as the primary task execution portal for storefront associates.

* **Open in Android Studio:** Open the `/apps/gtasks` subdirectory directly in Android Studio to enable full Gradle indexing.
* **Compile via Bazel:**
  ```bash
  bazel build //apps/gtasks:build
  ```
* **Install on Connected Device/Emulator (via ADB):**
  ```bash
  adb install -r apps/gtasks/app/build/outputs/apk/debug/app-debug.apk
  ```
* **Tunnel API Traffic for Physical USB Devices:**
  ```bash
  adb reverse tcp:8080 tcp:8080
  ```

For a step-by-step setup guide covering developer certificates, OAuth 2.0 client IDs, and Android Studio SDK configurations, read the **[Android Handset App Guide](docs/content/docs/android_setup.md)** and the **[Google Cloud & OAuth Setup Guide](docs/content/docs/cloud_setup.md)**.

### 5. Independent Component Builds & Tests

#### Build & Test the Backend Server
```bash
# Build server binary
bazel build //cmd:server

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

---

## Google Workspace Directory OU Structure & Test Identity Matrix

To test enterprise RBAC, multi-store manager overlaps, and regional grounding workflows, the local sandbox environment features a fully provisioned Google Workspace directory mapping 553 user accounts across 109 physical locations.

For the full reference documentation detailing:
* **Organizational Unit (OU) Hierarchy topologies**
* **Test Personnel account naming conventions**
* **Active Regional Managers & Geofootprints matrix**
* **Complete 109 Storefront Location Lookup table**
* **Standard Role Access profiles and Database Role ID assignments**

Read the **[Workspace Directory & Store Mapping Specifications](docs/content/docs/store_information.md)** in the internal specifications portal.




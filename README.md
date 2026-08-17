# Enterprise Task Engine

An enterprise-grade, multi-regional Task Orchestration Engine built on Go, React, Python ADK, and Android Jetpack Compose, driven by Bazel 8 and powered by Google's Model Context Protocol (MCP), Gemini AI, on-device Gemma, and Agentic User Interfaces (A2UI).

---

## Project Overview

The **Enterprise Task Engine** orchestrates physical retail and operational workflows, combining static rule-based scheduling with dynamic AI agent execution:

- **Administrative MDM & Schema CRUD:** Fully managed entities including Users, Roles, Organizations, Sites, Locations, Assets, Task Templates, Certifications, and SOPs.
- **AI-Assisted Task Execution & RAG:** Real-time query resolution via vector similarity search against localized SOP context windows (`pgvector` HNSW cosine indexing on AlloyDB).
- **Agentic User Interface (A2UI):** Dynamic, server-driven UI cards (v0.8.0 Stable) rendered natively across both web (React 19) and mobile (Android Jetpack Compose).
- **Model Context Protocol (MCP):** Full suite of agent tools exposed over JSON-RPC 2.0 via HTTP and SSE.
- **Speech Intelligence & Neural Translation:** Bidirectional associate-customer voice translation using Google Cloud Speech-to-Text, Cloud Translation, HD Text-to-Speech voices, and Google Cloud Chirp 3 Instant Custom Voice Cloning.
- **Dual-Engine LLM Tier on Mobile:** Seamless switching between cloud Vertex AI Gemini models and local on-device Gemma 2B INT4 inference via MediaPipe GenAI for offline resilience.
- **Distributed Concurrency & Locking:** Split-brain-proof leader election via PostgreSQL session advisory locks (`key: 5555`) pinned to dedicated `sql.Conn` TCP sockets, paired with `FOR UPDATE SKIP LOCKED` parallel worker queues and dead-letter watchdog recovery.

---

## Repository Architecture & Directory Layout

The workspace follows standard multi-language layout patterns integrated under Bazel 8:

| Directory | Description |
| :--- | :--- |
| [`/cmd/server`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/cmd/server) | Entry point for the Go Gin API Server and JSON-RPC MCP listeners. |
| [`/cmd/task_agent`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/cmd/task_agent) | Python FastAPI ADK Agent service with A2UI transpiler and SVG blueprint engine. |
| [`/cmd/db_loader`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/cmd/db_loader) | Automated database schema migrator and relational seed data loader. |
| [`/cmd/db_diagnose`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/cmd/db_diagnose) | Command-line diagnostic and connectivity validation utility. |
| [`/pkg/persistence`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/persistence) | GORM configuration, pgvector client hooks, repositories, and transactional DB pools. |
| [`/pkg/model`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/model) | Core domain entities, GORM tags, event types, and audit schema models. |
| [`/pkg/api`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/api) | Gin web router, OAuth/IAP middlewares, MCP handlers, and route controllers. |
| [`/pkg/service`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/pkg/service) | Business logic for task execution, RAG SOP vector indexing, scheduling, and voice translation. |
| [`/web/console`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/web/console) | React 19 Admin & Operations Console built with Vite, TypeScript, and Tailwind CSS. |
| [`/web/agentic`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/web/agentic) | Standalone speech-enabled Agentic UI cockpit with interactive digital twin map. |
| [`/web/render-test`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/web/render-test) | React-based rendering validation workbench. |
| [`/apps/gtasks`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/apps/gtasks) | Native Android associate handset portal app (Kotlin / Jetpack Compose M3). |
| [`/docs`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs) | Hugo documentation site root (Hugo configurations & Bazel rules). |

---

## Technical Specifications Index

For detailed architectural specifications, schemas, workflows, and deployment guides, browse the internal documentation portal:

| Specification Manual | Operational Domain | Key Technical Focus |
| :--- | :--- | :--- |
| **[Architectural Design Document](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/architecture.md)** | Core Architecture | Multi-tier topology, Gin API Gateway, AlloyDB, pgvector HNSW, Otel |
| **[App & UI Workflows](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/apps_and_ui.md)** | Client Ecosystem | React Admin Console, GTasks Native Android, Dual-Engine LLM |
| **[A2UI Architecture & Engine](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/a2ui_architecture.md)** | Dynamic UI Tier | A2UI v0.8 contracts, MCP tools, React & Compose rendering engines |
| **[Voice & Speech Intelligence](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/voice_and_translation.md)** | Speech Intelligence | Cloud STT, Neural Translation, TTS HD voices, Chirp 3 Voice Cloning |
| **[Event-Driven Mechanics](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/events.md)** | Event Ingestion | ARTS XML / Schema.org taxonomies, BATCH vs ADHOC alert triggers |
| **[Distributed Scheduler Daemon](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/scheduler.md)** | Concurrency & Queue | Postgres advisory locks (`key: 5555`), `SKIP LOCKED`, dead-letter watchdog |
| **[Workspace Directory & Stores](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/store_information.md)** | Identity & Hierarchy | Google Workspace OU structure, 109 storefronts, 553 test profiles |
| **[Google Cloud & OAuth Setup](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/cloud_setup.md)** | Infrastructure Setup | GCP project provisioning, OAuth 2.0 consent screens, Client IDs |
| **[Android Handset App Guide](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/android_setup.md)** | Mobile Deployment | Android Studio setup, ADB reverse port forwarding, APK installs |
| **[A2UI Development Lessons](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/a2ui_lessons.md)** | UI Integration Notes | Typographic schemas, column stretching, rate-limit mitigation |
| **[Governance & Licensing](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/governance_and_licensing.md)** | Governance & Decisions | Apache 2.0 terms, authors, design decisions, failure domain nuances |
| **[Contributing Guide](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/contributing.md)** | Community & Workflows | CLA, Google Code of Conduct, PR workflow, testing requirements |

---

## Build & Development Workflows

The entire repository is driven via **Bazel 8** for fast, hermetic builds.

### 1. Local Monorepo Setup
First, initialize your local environment configuration by copying the template:
```bash
cp example.env.local.toml .env.local.toml
```
Update **[`.env.local.toml`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/example.env.local.toml)** with your local PostgreSQL credentials, Google Cloud Project ID, and OAuth 2.0 Client ID (`.env.local.toml` is ignored by Git).

Next, initialize and resolve all Bazel modules and Node dependencies:
```bash
bazel mod tidy
```

### 2. Starting the Concurrent Development Environment
To start the frontend applications and the Go backend API server simultaneously in a parallel dev session:
```bash
bazel run //:dev_server
```
- **Go API Backend & MCP Server:** Binds to `:8080`
- **Admin Console SPA:** Binds to `:5173`
### 3. Documentation Site & Single-Page Bundler
- **Serve Local Documentation Portal (Hot-Reloading):**
  ```bash
  bazel run //docs:serve
  ```
  Serves at `http://localhost:1313/`.
- **Generate Single-Page Markdown & LLM Text Bundle:**
  ```bash
  bazel run //docs:generate_single_page
  # or build hermetically into bazel-bin/docs/
  bazel build //docs:bundle
  ```
  Generates `docs/docs_bundle.md` and `docs/llms-full.txt`.

### 4. Native Android Handset App (GTasks)
The native Android app is located in `/apps/gtasks`:
- **Build APK via Bazel:**
  ```bash
  bazel build //apps/gtasks:build
  ```
- **Install on Connected Device/Emulator (via ADB):**
  ```bash
  adb install -r apps/gtasks/app/build/outputs/apk/debug/app-debug.apk
  ```
- **Tunnel API Traffic for Physical USB Devices:**
  ```bash
  adb reverse tcp:8080 tcp:8080
  ```

### 5. Backend Server & Unit Tests
```bash
# Build server binary
bazel build //cmd:server

# Run all backend unit tests hermetically
bazel test //test/...
```

---

## Contributing & Community

We welcome contributions from Google open source contributors and internal engineers. Please review our **[Contributing Guide](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/docs/content/docs/contributing.md)** before submitting pull requests:
- Must sign a Google **[Contributor License Agreement (CLA)](https://cla.developers.google.com/)**
- Strictly follow **[Google Open Source Community Guidelines](https://opensource.google/conduct/)**
- All code must include an Apache 2.0 license header and pass Bazel verification (`bazel test //...`)

---

## Authors & Licensing

- **Author & Principal Architect:** Ryan McGuinness (`rmcguinness@google.com`, `rrmcguinness`)
- **Project Owners:** See **[`owners.txt`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/owners.txt)**
- **Organization:** Google LLC
- **License:** Apache License, Version 2.0 (see **[`LICENSE`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/LICENSE)** and **[`NOTICE`](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/NOTICE)**)

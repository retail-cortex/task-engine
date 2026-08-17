---
title: "Enterprise Task Engine Docs"
weight: 1
bookFlatSection: true
---

# Enterprise Task Engine

An enterprise-grade, multi-regional Task Orchestration Engine built on Go, React, Python ADK, and Android Jetpack Compose, built with Bazel 8 and powered by Google's Model Context Protocol (MCP), Gemini AI, on-device Gemma, and Agentic User Interfaces (A2UI).

---

## Master Specifications Index

| Specification Manual | Operational Domain | Key Technical Focus |
| :--- | :--- | :--- |
| **[Architectural Design Document]({{< ref "architecture.md" >}})** | Core Architecture | Multi-tier topology, Gin API Gateway, AlloyDB, pgvector HNSW, Otel |
| **[App & UI Workflows]({{< ref "apps_and_ui.md" >}})** | Client Ecosystem | React Admin Console, GTasks Native Android, Dual-Engine LLM |
| **[A2UI Architecture & Engine]({{< ref "a2ui_architecture.md" >}})** | Dynamic UI Tier | A2UI v0.8 contracts, MCP tools, React & Compose rendering engines |
| **[Voice & Speech Intelligence]({{< ref "voice_and_translation.md" >}})** | Speech Intelligence | Cloud STT, Neural Translation, TTS HD voices, Chirp 3 Voice Cloning |
| **[Event-Driven Mechanics]({{< ref "events.md" >}})** | Event Ingestion | ARTS XML / Schema.org taxonomies, BATCH vs ADHOC alert triggers |
| **[Distributed Scheduler Daemon]({{< ref "scheduler.md" >}})** | Concurrency & Queue | Postgres advisory locks (`key: 5555`), `SKIP LOCKED`, dead-letter watchdog |
| **[Workspace Directory & Stores]({{< ref "store_information.md" >}})** | Identity & Hierarchy | Google Workspace OU structure, 109 storefronts, 553 test profiles |
| **[Google Cloud & OAuth Setup]({{< ref "cloud_setup.md" >}})** | Infrastructure Setup | GCP project provisioning, OAuth 2.0 consent screens, Client IDs |
| **[Android Handset App Guide]({{< ref "android_setup.md" >}})** | Mobile Deployment | Android Studio setup, ADB reverse port forwarding, APK installs |
| **[A2UI Development Lessons]({{< ref "a2ui_lessons.md" >}})** | UI Integration Notes | Typographic schemas, column stretching, rate-limit mitigation |
| **[Governance & Licensing]({{< ref "governance_and_licensing.md" >}})** | Governance & Decisions | Apache 2.0 terms, authors, design decisions, failure domain nuances |

---

## Repository Architecture & Directory Layout

| Directory | Description |
| :--- | :--- |
| [`/cmd/server`](../../cmd/server) | Entry point for the Go Gin API Server and JSON-RPC MCP listeners. |
| [`/cmd/task_agent`](../../cmd/task_agent) | Python FastAPI ADK Agent service with A2UI transpiler and SVG blueprint engine. |
| [`/cmd/db_loader`](../../cmd/db_loader) | Automated database schema migrator and relational seed data loader. |
| [`/cmd/db_diagnose`](../../cmd/db_diagnose) | Command-line diagnostic and connectivity validation utility. |
| [`/pkg/persistence`](../../pkg/persistence) | GORM configuration, pgvector client hooks, repositories, and transactional DB pools. |
| [`/pkg/model`](../../pkg/model) | Core domain entities, GORM tags, event types, and audit schema models. |
| [`/pkg/api`](../../pkg/api) | Gin web router, OAuth/IAP middlewares, MCP handlers, and route controllers. |
| [`/pkg/service`](../../pkg/service) | Business logic for task execution, RAG SOP vector indexing, scheduling, and translation. |
| [`/web/console`](../../web/console) | React 19 Admin & Operations Console built with Vite, TypeScript, and Tailwind CSS. |
| [`/web/agentic`](../../web/agentic) | Standalone speech-enabled Agentic UI cockpit with interactive digital twin map. |
| [`/web/render-test`](../../web/render-test) | React-based rendering validation workbench. |
| [`/apps/gtasks`](../../apps/gtasks) | Native Android associate handset portal app (Kotlin / Jetpack Compose M3). |
| [`/docs`](../../docs) | Hugo documentation site root (Hugo configurations & Bazel rules). |

---

## Quickstart Commands

```bash
# Start backend API server and MCP listener
bazel run //:dev_server

# Start Hugo documentation portal locally at http://localhost:1313
bazel run //docs:serve

# Generate unified single-page markdown bundle & llms-full.txt
bazel run //docs:generate_single_page
# or build hermetically into bazel-bin/docs/
bazel build //docs:bundle

# Execute full backend test suites hermetically
bazel test //test/...

# Build React Admin Console
bazel build //web/console:build

# Build GTasks Android APK
bazel build //apps/gtasks:build
```

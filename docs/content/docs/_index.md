---
title: "Technical Specifications Portal"
weight: 1
bookFlatSection: true
---

# Technical Specifications & Architecture Portal

Welcome to the internal, design-level reference manual and architectural specifications portal for the **Enterprise Task Engine**.

Choose from the following specification manuals:

### Core Architecture & System Design
* **[Architectural Design Document]({{< ref "architecture.md" >}})** — Master system topology, technology stacks, API routing, and database schemas.
* **[App & UI Workflows]({{< ref "apps_and_ui.md" >}})** — Deep dive into the React Admin Console, GTasks Native Android app, Dual-Engine LLM reasoning, and web sandboxes.
* **[A2UI Architecture & Rendering Engine]({{< ref "a2ui_architecture.md" >}})** — A2UI v0.8 protocol contracts, MCP tool integrations, Python ADK transpiler, and multi-platform rendering engines.
* **[Voice Translation & Speech Intelligence]({{< ref "voice_and_translation.md" >}})** — Google Cloud STT, Neural Translation, TTS HD voices, and Chirp 3 instant custom voice cloning.

### Operational Engines & Workflows
* **[Event Ingestion System Specifications]({{< ref "events.md" >}})** — ARTS XML and Schema.org taxonomies, BATCH vs ADHOC styles, and streaming alert triggers.
* **[Distributed Scheduler & Recovery Engine]({{< ref "scheduler.md" >}})** — PostgreSQL session advisory locking (`key: 5555`), `SKIP LOCKED` queues, and dead-letter watchdog recovery.
* **[Workspace Directory & Store Footprints]({{< ref "store_information.md" >}})** — Google Workspace OU hierarchy, role mappings, and complete 109-store test matrix.

### Setup & Infrastructure Guides
* **[Google Cloud Platform & OAuth Setup]({{< ref "cloud_setup.md" >}})** — GCP project setup, OAuth 2.0 consent screens, and client credentials.
* **[Android Handset App Developer Guide]({{< ref "android_setup.md" >}})** — Android Studio setup, ADB reverse port forwarding, and handset deployment.
* **[A2UI Development & Integration Lessons]({{< ref "a2ui_lessons.md" >}})** — Historical lessons learned, typographic constraints, and 429 rate-limit mitigations.

### Governance & Standards
* **[Governance, Licensing & Architectural Decisions]({{< ref "governance_and_licensing.md" >}})** — Apache 2.0 terms, maintainers, architectural design trade-offs, and technical nuances.
* **[Contributing Guide]({{< ref "contributing.md" >}})** — Contributor License Agreement (CLA) requirements, Code of Conduct, style guides, PR workflow, and testing verification.

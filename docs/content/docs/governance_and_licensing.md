---
title: "Governance & Licensing"
weight: 60
---

# Governance, Licensing, Authorship & Architectural Decisions
## Enterprise Task Engine: Engineering Charter & Nuance Specifications (v5.0)

> [!NOTE]
> This document specifies the project governance, legal licensing, authorship, architectural decision rationale, and failure domain nuances governing the Enterprise Task Engine codebase.

---

## 1. Project Authorship & Maintainers

- **Engineering Organization:** Google LLC
- **Project Lead & Architecture:** Task Engine Engineering Team / Ryan McGuinness
- **Repository:** `gemini_task_engine` (Google Solutions Engineering)

---

## 2. Licensing & Legal Attributions

The Enterprise Task Engine is distributed under the **Apache License, Version 2.0**.

### A. License Summary
```text
Copyright 2026 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

### B. Third-Party Software Notices ([`NOTICE`](../../NOTICE))
This product includes software developed at Google LLC (https://www.google.com/) and incorporates open-source libraries under compliant licenses (MIT, Apache 2.0, BSD-3-Clause):
- **Go Ecosystem:** Gin Web Framework, GORM ORM, pgvector-go, OpenTelemetry Go SDK, Google Cloud Client Libraries, modenv.
- **Python Ecosystem:** FastAPI, Uvicorn, Google Agent Development Kit (ADK), Google GenAI SDK, pg8000, httpx.
- **Node/React Ecosystem:** React 19, TypeScript, Vite, Tailwind CSS, Material 3 Theme Provider, Lucide Icons.
- **Android Ecosystem:** Android Jetpack Compose, Material 3, Retrofit 2, OkHttp, Coil SVG, Google MediaPipe Tasks GenAI.

---

## 3. Architectural Design Decisions & Trade-Off Analysis

The architecture of the Enterprise Task Engine was engineered to maximize deployment simplicity, operational throughput, and data consistency while minimizing external infrastructure dependencies.

| Decision Domain | Selected Approach | Rationale & Trade-Off Analysis |
| :--- | :--- | :--- |
| **Distributed State & Leader Election** | PostgreSQL Advisory Locks (`key: 5555`) pinned to `sql.Conn` | Eliminates the operational overhead and failure modes of dedicated ZooKeeper/Redis clusters. Self-heals on crash. |
| **Task Queue & Workload Dispatch** | Transactional GORM `FOR UPDATE SKIP LOCKED` row locking | Provides ACID guarantees and prevents duplicate task claims without requiring separate Kafka/RabbitMQ message brokers. |
| **Grounded AI & SOP Vector Search** | In-Database `pgvector` with HNSW Cosine Index inside AlloyDB | Eliminates dual-write sync drift between SQL DB and external vector DBs (Pinecone, Milvus). |
| **User Interface Framework** | Server-Driven A2UI Protocol (v0.8.0) over MCP JSON-RPC 2.0 | Enables dynamic card layouts, real-time forms, and actions without deploying new client releases to App Stores. |
| **Mobile On-Device Resilience** | Dual-Engine LLM: Remote Gemini + Local Gemma 2B MediaPipe | Guarantees operational uptime for floor associates inside shielded retail basements. |
| **Monorepo Build Automation** | Hermetic Bazel 8 with Bzlmod (`MODULE.bazel`) | Fast incremental compilation, reproducible container builds, and unified cross-language CI. |

---

## 4. Comprehensive Technical Nuances & Failure Domain Handling

### A. Advisory Lock TCP Connection Pinning (`sql.Conn`)
- **The Problem:** Standard connection pools (e.g., GORM's default database pool) recycle TCP sockets across queries. Session-bound PostgreSQL advisory locks (`pg_try_advisory_lock`) are bound strictly to the physical TCP connection. If the pool switches connections, the lock is silently lost, causing split-brain leader elections.
- **The Solution:** In [`scheduler.go`](../../pkg/service/scheduler.go), the scheduler explicitly allocates and holds a single, dedicated connection socket (`sqlDB.Conn(ctx)`). All lock attempts, heartbeats, and release checks execute over this pinned socket.

### B. SKIP LOCKED Worker Claiming & Lock Tagging
- **The Problem:** Multiple horizontal worker nodes polling a centralized SQL table encounter deadlock contention or race conditions if using naive `SELECT ... WHERE status = 'PENDING'` queries.
- **The Solution:** Workers execute `SELECT ... FOR UPDATE SKIP LOCKED LIMIT 5` inside an atomic GORM transaction, tagging claimed rows with `locked_at = NOW()` and `locked_by = node-ID`. Concurrent workers skip locked rows with zero wait time.

### C. Watchdog Dead-Letter Recovery Loops
- **The Problem:** If a worker node crashes mid-execution (OOM kill, network disconnect), rows remain locked in `IN_PROGRESS` indefinitely.
- **The Solution:** The Leader node runs a watchdog loop every 30 seconds scanning for tasks where `locked_at < NOW() - 5m`. If `retry_count < max_retries`, the watchdog resets locks to `PENDING` and increments the retry count. If retries exceed 3, the task is routed to `DEAD_LETTER` status to alert operational supervisors.

### D. SOP Document Drift Detection (ETag & SHA-256 Checksums)
- **The Problem:** Standard Operating Procedures (SOPs) hosted on external URLs or web storage may change without notification, resulting in stale vector embeddings.
- **The Solution:** The Leader node issues periodic HTTP `HEAD` requests. If `ETag` or `Last-Modified` headers have changed, or if a binary file's SHA-256 fingerprint has drifted, the old `SOPProcess` is deactivated and a new background re-indexing pipeline is launched automatically.

### E. Hermetic Base64 Vector Inlining for Mixed Content
- **The Problem:** In secure production environments hosted over HTTPS, embedding dynamic blueprint SVG images from local HTTP routes (e.g., `http://localhost:8081/api/v1/blueprint`) triggers browser Mixed Content blocking, causing floor plans to disappear.
- **The Solution:** The A2UI transpiler dynamically renders the SVG in-memory, base64-encodes the content, and outputs a self-contained Data URI (`data:image/svg+xml;base64,...`) directly in the card envelope.

### F. Degraded Offline Mode & Mock Database Fallback
- **The Problem:** In local developer environments or during CI runs without a running PostgreSQL instance, server bootstrap fails on connection dial.
- **The Solution:** If PostgreSQL is unreachable on boot, `persistence.InitDB` logs a warning and initializes `persistence.NewOfflineDB`, allowing the Gin API server to start in degraded mode. The frontend console renders an `OfflineOverlay` visual indicator alerting developers.

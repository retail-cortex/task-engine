---
title: "Architectural Specification"
weight: 10
---

# Architectural Design Document
## Enterprise Task Engine: Master Specifications Manual (v5.0)

> [!NOTE]
> This master architectural specifications document catalogues the topological design patterns, database schemas, and platform integrations implemented inside the Enterprise Task Engine monorepo.

---

## 1. Executive Summary & Core Use Cases

The Enterprise Task Engine is a multi-node, horizontally scalable operational platform designed to coordinate, automate, and verify physical retail operations in real-time. The system bridges core administrative configuration layers (Master Data Management) with standard dynamic workforce execution pipelines, backed by Google GenAI Vertex AI models (Gemini) and interactive agentic user interfaces (A2UI).

```
 ┌─────────────────────────────────────────────────────────┐
 │               INTERACTIVE CLIENT COCKPIT                │
 │    Glassmorphic A2UI React Console / Gemini Coach       │
 └───────────┬─────────────────────────────────▲───────────┘
             │                                 │
             │ (JSON-RPC MCP API Triggers)     │ (Grounded UI Card Layouts)
             ▼                                 │
 ┌─────────────────────────────────────────────┴───────────┐
 │                  CORE GO BACKEND SERVER                 │
 │     Gin REST APIs, MCP Server, Distributed Scheduler    │
 └───────────┬─────────────────────────────────▲───────────┘
             │                                 │
             │ (Atomic transaction SQL blocks) │ (Data Streams / Row Checks)
             ▼                                 │
 ┌─────────────────────────────────────────────┴───────────┐
 │                 ALLOYDB STATE DATABASE                  │
 │      HNSW pgvector, SKIP LOCKED queues, advisory locks  │
 └─────────────────────────────────────────────────────────┘
```

The system addresses four primary operational domains:
1. **Master Data Management (MDM):** Provides robust administrative APIs for defining structural retail boundaries: Organizations, Sites (Stores/Warehouses), Locations (fixtures/shelves coordinates maps), Roles, Assets, Certifications, and SOPs templates.
2. **Intelligent Shift Initialization:** Automatically provisions a shift session context pool (`shift_agent_sessions`) upon user clock-in, backing the associate's Gemini ADK live preview agent coach session.
3. **Task Auto-Generation & Prioritized Queues:** Coordinates scheduled business calendar materialization sweeps (**BATCH** chron cycles) and dynamic streaming alert ingestions (**ADHOC** sensor alarms, till out alerts, register call bells). Enforces strict asset certifications constraint checks.
4. **Grounded AI-Assisted Operations & RAG:** Real-time dynamic alerts calculate vector embeddings, query database-indexed SOP text chunks utilizing pgvector HNSW similarity algorithms, and inject grounded SOP instructions directly inside active task checklists. Exposes rich structured layouts dynamically using Agentic UI cards (A2UI).

---

## 2. System Technology Stack

The platform is designed to run hermetically across containerized environments under GKE, Cloud Run, and local developer sandboxes using the following stacks:

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                                SYSTEM STACK                                  │
├──────────────────┬───────────────────────────────────────────────────────────┤
│ Backend Platform │ Go (Golang) Gin API Server Engine, GORM postgres ORM     │
├──────────────────┼───────────────────────────────────────────────────────────┤
│ AI Intelligence  │ Model Context Protocol (MCP) Server, Vertex AI Gemini     │
├──────────────────┼───────────────────────────────────────────────────────────┤
│ Database Tier    │ AlloyDB / PostgreSQL Cluster, pgvector HNSW indexes       │
├──────────────────┼───────────────────────────────────────────────────────────┤
│ User Interfaces  │ React 19 Frontend Dashboard SPA, Vite TypeScript, A2UI    │
├──────────────────┼───────────────────────────────────────────────────────────┤
│ Infrastructure   │ Terraform IaC, Docker, Google Cloud Run / GKE, Otel Trace │
└──────────────────┴───────────────────────────────────────────────────────────┘
```

---

## 3. High-Performance Concurrency & Distributed State Locking

To support dozens of active horizontal microservice replicas without requiring complex external cache coordinators (Redis, ZooKeeper), the database cluster serves as the unified system lock voter and scheduling state manager.

### A. Distributed Leader Election & Advisory Locking
Leader election limits chron sweeps, document updates scans, and lock timeout watchdogs to a single active server replica, avoiding double sweeps:
* **PostgreSQL Advisory Locks:** Cluster replicas bid for leadership on startup and checks loops by requesting a session-bound Postgres advisory lock on key `5555`.
* **TCP Connection Pinned Sockets (`sql.Conn`):** Standard pooled database clients recycle net connections, causing session locks to be dropped. The scheduler prevents this by reserving a single dedicated TCP socket direct from the driver pool context (`sqlDB.Conn`), pinning the advisory lock exclusively to this dedicated connection loop context.
* **Self-Healing Failover:** If the active leader node crashes, Postgres automatically drops the TCP pool after standard keep-alive intervals, releasing the advisory lock key instantly so worker nodes bid on leadership in their next `15s` check sweep.

### B. Concurrency-Safe SKIP LOCKED Parallel Workers
To allow replica worker nodes to pull and complete checklist tasks in parallel with zero lock contention or duplicate executions:
* **SKIP LOCKED transaction boundaries:** Workers claim pending operations under GORM transactions executing row-level locks pings:
  `SELECT * FROM task_executions WHERE status = 'PENDING' AND ... LIMIT ? FOR UPDATE SKIP LOCKED`
* **Logical Lock Tagging:** Claimed task rows are updated inside the transaction to `IN_PROGRESS` and tagged with the node's unique ID (`locked_by = node-ID`) and locked timestamp. Once transaction locks release on commit, the logical lock keeps the rows isolated until execution is marked `COMPLETED`.

### C. Watchdog Dead Letter Recovery Loops
A periodic watchdog sweep (Leader node only, every `30s`) scans task rows stuck in `IN_PROGRESS` state longer than standard lock timeout boundaries (`5m`):
* **Lock Recovery:** If the task's `RetryCount` is less than `MaxRetries` (default: 3), the watchdog logs the error, increments `RetryCount`, resets locks (`locked_at = NULL`, `locked_by = NULL`), and enqueues the row back to the queue under a `PENDING` status.
* **Dead Letter Routing:** If the task fails 3 times, it is permanently isolated by moving the status to the terminal `'DEAD_LETTER'` queue, alerting administrative operations diagnostics channels.

---

## 4. API Endpoints Architecture

### A. Administrative Data Management (CRUD Targets)
Administrative endpoints fall under `/api/v1/admin/*` and require elevated supervisor or system manager credentials.

**Organizations, Users, and RBAC**
* `GET /api/v1/admin/users` - Query user details and compliance certifications status.
* `POST /api/v1/admin/roles` - Seed standard organizational roles.
* `PUT /api/v1/admin/users/{id}/roles` - Map user identities to roles.
* `POST /api/v1/admin/organizations` - Register organizational tenants.
* `GET /api/v1/admin/organizations` - List active tenant organizations.
* `PUT /api/v1/admin/organizations/{id}/users` - Map user to corporate organizations boundary.

**Spatial Nodes, Fixtures, and Assets**
* `POST /api/v1/admin/sites` - Register retail sites (Store #1000 - Dallas Texas).
* `POST /api/v1/admin/locations` - Register fixtures, registers, or backroom rack coordinate points.
* `POST /api/v1/admin/assets` - Register physical machinery and tools.
* `PATCH /api/v1/admin/assets/{id}/status` - Update asset status states (`AVAILABLE`, `BROKEN`, `MAINTENANCE`).

**SOP Documents & Scheduling**
* `POST /api/v1/admin/sops` - Intake standard SOP files (DOCX, PDF, HTML, XLSX). Triggers RAG chunking embeds pipelines.
* `POST /api/v1/admin/tasks/templates` - Create task definitions, including checklist JSON matrices.
* `POST /api/v1/admin/tasks/templates/{id}/rules` - Map Maker/Checker supervisor approval schemas.
* `POST /api/v1/admin/availabilities` - Map employee availability calendars and RFC 5545 RRULE shift patterns.

**Background Job Scheduler Diagnostics**
* `GET /api/v1/admin/scheduler/status` - Exposes multi-node leader state, active worker counts, processed/failed task metrics, and last error diagnostics.
* `POST /api/v1/admin/scheduler/trigger` - Force materializes scheduled batch sweeps calendar checks.

### B. Operational Tasking & Agent Interface
Operational routes are used by floor associates, client dashboards, and Vertex AI ADK MCP servers.

* `GET /health/readiness` - Diagnostic target to verify active AlloyDB pools connectivity.
* `GET /static/*` - Served natively by the Go server filesystem engine to resolve static SOP documents templates.
* `POST /api/v1/sessions/shift/{shiftId}/chat` - Primary Vertex AI Gemini MCP JSON-RPC conversational endpoint.
* `GET /api/v1/organizations/{orgId}/sites/{siteId}/tasks` - Fetches prioritized, active site queues (filterable by local spatial location coordinates).
* `PATCH /api/v1/organizations/{orgId}/sites/{siteId}/tasks/{id}/status` - Mutates checklist task statuses, executing GORM hook ledgers.
* `POST /api/v1/organizations/{orgId}/sites/{siteId}/tasks/{id}/override` - Bypasses constraints by submitting supervisor justification tags.
* `POST /api/v1/organizations/{orgId}/sites/{siteId}/trades` - Initiates peer-to-peer shift and task trades.

---

## 5. PostgreSQL/AlloyDB Relational Database Schema

### Database Extensions
```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector"; -- pgvector extension mapping
```

### A. Organization, Spatial & Identity Schemas
```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    o_auth_provider VARCHAR(50) NOT NULL,
    o_auth_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    UNIQUE (o_auth_provider, o_auth_id)
);

CREATE TABLE user_organizations (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (organization_id, user_id)
);

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    site_type VARCHAR(50) NOT NULL DEFAULT 'STORE',
    altitude_meters DECIMAL(10, 4) NOT NULL DEFAULT 0.0,
    icao_code VARCHAR(10) NOT NULL DEFAULT '',
    address TEXT,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);

CREATE TABLE user_sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, site_id)
);

CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES locations(id) ON DELETE CASCADE DEFAULT NULL,
    name VARCHAR(255) NOT NULL,
    location_type VARCHAR(50) NOT NULL DEFAULT 'FIXTURE',
    location_function_type VARCHAR(50) NOT NULL DEFAULT 'DISPLAY',
    x DECIMAL(10, 4) NOT NULL DEFAULT 0.0,
    y DECIMAL(10, 4) NOT NULL DEFAULT 0.0,
    z DECIMAL(10, 4) NOT NULL DEFAULT 0.0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### B. Assets & Certification Schemas
```sql
CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    asset_tag VARCHAR(100) UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'AVAILABLE',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);

CREATE TABLE certifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    issuer VARCHAR(255),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);

CREATE TABLE user_certifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    certification_id UUID NOT NULL REFERENCES certifications(id) ON DELETE CASCADE,
    issued_date TIMESTAMPTZ NOT NULL,
    expiration_date TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    UNIQUE (user_id, certification_id)
);
```

### C. Grounded SOP Documents & RAG Schemas
```sql
CREATE TABLE sops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    canonical_url VARCHAR(1024),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb, -- Maps ETags, file checksums, expiration dates
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sop_processes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sop_id UUID NOT NULL REFERENCES sops(id) ON DELETE CASCADE,
    chunking_strategy VARCHAR(100) NOT NULL,
    embedding_model VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sop_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sop_id UUID NOT NULL REFERENCES sops(id) ON DELETE CASCADE,
    sop_process_id UUID NOT NULL REFERENCES sop_processes(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    content TEXT NOT NULL,
    embedding VECTOR(768) NOT NULL, -- pgvector structural field mapping (768 dimensions)
    UNIQUE (sop_process_id, chunk_index)
);

-- pgvector Index Definition utilizing HNSW algorithms
CREATE INDEX ON sop_chunks USING hnsw (embedding vector_cosine_ops);

CREATE TABLE task_sops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    sop_id UUID NOT NULL REFERENCES sops(id) ON DELETE CASCADE,
    UNIQUE (task_id, sop_id)
);
```

### D. Workforces Scheduling, Shifts & Materialized Events
```sql
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    site_id UUID REFERENCES sites(id) ON DELETE SET NULL,
    task_id UUID REFERENCES tasks(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL DEFAULT 'RetailShift', -- Classifications
    event_style VARCHAR(20) NOT NULL DEFAULT 'BATCH', -- BATCH or ADHOC category maps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_event_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    rrule TEXT, -- RFC 5545 Recurrence string rules (Standard Scheduled Batch runs)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, event_id)
);

CREATE TABLE user_event_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES user_event_schedules(id) ON DELETE CASCADE,
    instance_start_date TIMESTAMPTZ NOT NULL,
    instance_end_date TIMESTAMPTZ NOT NULL,
    event_status VARCHAR(50) NOT NULL DEFAULT 'EventScheduled', -- Schema.org align
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (schedule_id, instance_start_date)
);
```

### E. Task Execution Queues, Audits & Trades ledgers
```sql
CREATE TABLE tasks (
    id PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    task_type VARCHAR(50) NOT NULL DEFAULT 'STANDARD',
    target_role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    priority INT NOT NULL DEFAULT 3,
    step_order INT DEFAULT 0,
    estimated_duration_minutes INT,
    checklist_template JSONB DEFAULT '[]'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);

CREATE TABLE task_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    is_consumable BOOLEAN NOT NULL DEFAULT FALSE,
    quantity_required INT DEFAULT 1,
    is_hard_blocker BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (task_id, asset_id)
);

CREATE TABLE task_approval_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    required_role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    timing VARCHAR(50) NOT NULL DEFAULT 'POST_EXECUTION',
    is_strict BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE task_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_template_id UUID NOT NULL REFERENCES tasks(id) ON DELETE RESTRICT,
    parent_execution_id UUID REFERENCES task_executions(id) ON DELETE CASCADE,
    execution_type VARCHAR(50) NOT NULL DEFAULT 'STANDARD',
    subject_execution_id UUID REFERENCES task_executions(id) ON DELETE CASCADE,
    initiator_id UUID REFERENCES users(id) ON DELETE SET NULL,
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_instance_id UUID NOT NULL REFERENCES user_event_instances(id) ON DELETE CASCADE,
    description TEXT DEFAULT NULL, -- Enriched dynamic RAG compliance context text
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    priority INT NOT NULL DEFAULT 3,
    due_at TIMESTAMPTZ,
    prerequisite_execution_id UUID REFERENCES task_executions(id) ON DELETE SET NULL,
    decision VARCHAR(50),
    completed_at TIMESTAMPTZ,
    checklist_state JSONB DEFAULT '{}'::jsonb,
    override_flags JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    -- Multi-node scheduler transaction locks tracking fields
    locked_at TIMESTAMPTZ DEFAULT NULL,
    locked_by VARCHAR(100) DEFAULT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    last_error TEXT DEFAULT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);

-- Mapped index optimized for SKIP LOCKED transaction polls speed
CREATE INDEX idx_executions_locked ON task_executions(status, locked_at);

CREATE TABLE task_execution_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_execution_id UUID NOT NULL REFERENCES task_executions(id) ON DELETE CASCADE,
    changed_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action_type VARCHAR(50) NOT NULL,
    previous_state JSONB,
    new_state JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE task_trades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_execution_id UUID NOT NULL REFERENCES task_executions(id) ON DELETE CASCADE,
    initiator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    proposed_assignee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);
```

### F. Shift Coach Sessions context
```sql
CREATE TABLE shift_agent_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shift_instance_id UUID NOT NULL REFERENCES user_event_instances(id) ON DELETE CASCADE,
    message_history JSONB NOT NULL DEFAULT '[]'::jsonb,
    session_context JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    UNIQUE (assignee_id, shift_instance_id)
);
```

---

## 6. Interactive Frontend & A2UI Cockpit (`web/console`)

The floor associate's client operations cockpit is designed as a **Glassmorphic A2UI Operations Console** built on React, Vite, and tailwind typography. Exposes the following visual tiers:
1. **Glassmorphic Dashboard cockpit:** Radial glow dark themes, tailored neon priority glows indicators, interactive lists grid panels, and full Dev mock fallbacks.
2. **actions column styling (flex justify-end):** Mapped natively to prevent horizontal wrapping.
3. **A2UI Spatial coordinates map:** Integrated SVG architectural blueprint coordinates map of the store layout displaying pulsating focus target beacons.
4. **Hanna conversational coach chat & Suggestions chips:** Direct chat dialog returns coaching sarcasms and enqueues **dynamic visual A2UI Cards**:
   * **Cash Drop Card:** displays cash drawer drops checklists, security vault coordinates, and provides a direct supervisor bypass/verify trigger button.
   * **Shift Trade Card:** renders side-by-side shift comparison grids with direct makerchecker action swap buttons.
   * **METAR Wind Card:** visualizes airport wind audit parameters.

---

## 7. Infrastructure, Deployment & Observability

### A. Shaded distroless containers on GKE / Cloud Run
Backend compiled Go statically-linked executable binaries package directly inside minimal distroless Docker container configurations to keep production attack surfaces minimal. Leverages GCP Cloud Run scalability, scaling to zero upon inactive cycles.

### B. Otel & Structured Logs Correlation
OpenTelemetry maps correlation hooks throughout database transactions and AI sessions. JSON trace telemetry is written out, routing structured correlation IDs seamlessly across GORM audit engines, MCP tool execution runs, and Cloud Logging stacks.

### C. Terraform IaC Modules provisioning AlloyDB
Standard Terraform HCL maps corporate subnets, AlloyDB clusters, and pgvector HNSW index configurations cleanly. State files are securely locked and persisted under a primary GCS backend bucket:

```hcl
terraform {
  required_version = ">= 1.5.0"
  backend "gcs" {
    bucket = "nexus-tasking-tfstate"
    prefix = "env/prod"
  }
  required_providers {
    google = { source = "hashicorp/google", version = "~> 5.0" }
  }
}

resource "google_storage_bucket" "sop_bucket" {
  name                        = "retail-tasking-sops-${var.env}"
  location                    = "US"
  uniform_bucket_level_access = true
}

resource "google_alloydb_cluster" "primary" {
  cluster_id = "tasking-db-cluster"
  location   = "us-central1"
  network    = google_compute_network.vpc.id
}

resource "google_cloud_run_v2_service" "api_backend" {
  name     = "tasking-api-backend"
  location = "us-central1"
  template {
    containers {
      image = "us-docker.pkg.dev/${var.project}/repo/backend:latest"
      liveness_probe {
        http_get {
          path = "/health/readiness"
        }
      }
      env {
        name  = "DB_HOST"
        value = google_alloydb_instance.primary.ip_address
      }
    }
  }
}
```

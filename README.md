# FlowForge

**A distributed workflow orchestration platform built with Go, React, PostgreSQL, Redis, and RabbitMQ.**

FlowForge is a workflow automation MVP that lets users define multi-step workflows, validate their dependency graph, queue executions, and inspect run status and output. Its backend separates the HTTP API from background workers and uses PostgreSQL as the source of truth, RabbitMQ for execution messages, and Redis leases to coordinate workers.

> **Project status:** MVP / active development. The current executor processes steps sequentially. Some capabilities represented in the data model or architecture are planned rather than fully implemented; see [Current capabilities and limitations](#current-capabilities-and-limitations).

---

## Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Technology stack](#technology-stack)
- [How execution works](#how-execution-works)
- [Supported step types](#supported-step-types)
- [Getting started](#getting-started)
- [Running the frontend](#running-the-frontend)
- [Configuration](#configuration)
- [API reference](#api-reference)
- [Example workflow](#example-workflow)
- [Data model](#data-model)
- [Reliability and delivery semantics](#reliability-and-delivery-semantics)
- [Current capabilities and limitations](#current-capabilities-and-limitations)
- [Security notes](#security-notes)
- [Repository structure](#repository-structure)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

## Overview

Workflow orchestration becomes difficult when a process contains multiple dependent tasks, asynchronous execution, retries, and failures. FlowForge explores these problems through a small distributed system with independently running API and worker processes.

With the current implementation, you can:

- Create and list workflow definitions.
- Validate step identifiers, dependency references, and dependency cycles.
- Submit workflow runs for asynchronous execution.
- Execute supported steps in dependency order.
- Configure per-step retry attempts, linear backoff, and timeouts.
- Persist workflow definitions and run status in PostgreSQL.
- Queue execution requests through RabbitMQ.
- Use Redis leases to reduce concurrent processing of the same run.
- Inspect workflow run history and execution status from a React dashboard.

## Architecture

```text
                 React + TypeScript UI
                          |
                          | HTTP / JSON
                          v
                   Go HTTP API
                    /        \
                   v          v
             PostgreSQL    RabbitMQ
            (source of       |
              truth)         | execution message
                             v
                       Go Worker
                         /     \
                        v       v
                    PostgreSQL  Redis
                    (run state) (worker lease)
```

### Component responsibilities

| Component | Responsibility |
|---|---|
| React + TypeScript + Vite | Workflow dashboard, creation of a sample workflow, run history, and execution status |
| Go API | HTTP endpoints, request validation, workflow persistence, and run submission |
| Go worker | Consumes queued jobs and executes workflow steps |
| PostgreSQL | Stores workflow definitions and workflow run input, output, and status |
| RabbitMQ | Delivers workflow execution messages to workers |
| Redis | Provides a time-limited lease to reduce duplicate concurrent processing |

PostgreSQL is the authoritative store. Queue messages contain workflow and run identifiers; workers reload the persisted workflow and input when processing a message.

## Technology stack

**Backend**
- Go 1.23
- net/http
- PostgreSQL with pgx
- RabbitMQ with amqp091-go
- Redis with go-redis
- UUID-based workflow and run identifiers

**Frontend**
- React
- TypeScript
- Vite
- Lucide React

**Infrastructure**
- Docker
- Docker Compose
- PostgreSQL 17
- Redis 7
- RabbitMQ 3 with management plugin

## How execution works

1. A client sends a workflow definition to the API.
2. The API checks that the workflow contains steps, that each step has an ID and type, that IDs are unique, that dependencies refer to existing steps, and that the graph has no cycle.
3. The API saves the workflow definition in PostgreSQL.
4. A run request creates a run with the `queued` status and publishes a message to RabbitMQ.
5. A worker consumes the message and attempts to acquire a Redis lease for that run.
6. The worker reloads the run and workflow from PostgreSQL, changes the run status to `running`, and executes the steps in topological order.
7. The worker stores the final output and marks the run `succeeded` or `failed`.

### Workflow run states

```text
queued  ->  running  ->  succeeded
                     ->  failed
```

The database schema also allows `cancelled`, but a cancellation endpoint is not currently implemented.

## Supported step types

The engine currently implements these step types:

| Type | Behaviour |
|---|---|
| `noop` | Returns a simple completion message |
| `set` | Returns the step's configured JSON object |
| `delay` | Waits for the configured number of seconds, subject to the step timeout |
| `condition` | Compares an input value with a configured expected value and returns the comparison result |

A step can include:

- `depends_on`: IDs of steps that must complete first.
- `retry.max_attempts`: maximum number of attempts when a step fails.
- `retry.backoff_seconds`: base delay between retries. The current implementation increases the delay linearly by attempt number.
- `timeout_seconds`: per-attempt execution timeout.

The `condition` step currently produces a boolean result; it does not yet control branching or skip downstream steps. The `http` step is explicitly disabled in the MVP.

## Getting started

### Prerequisites

- Git
- Docker and Docker Compose
- Node.js and npm (to run the frontend locally)

### 1. Clone the repository

```bash
git clone https://github.com/Kushagrarai09/flowforge.git
cd flowforge
```

### 2. Start backend services

From the repository root:

```bash
docker compose up --build
```

Docker Compose starts:

- PostgreSQL on host port `5433`
- Redis on host port `6379`
- RabbitMQ on host port `5672`
- RabbitMQ management UI on host port `15672`
- FlowForge API on host port `8080`
- FlowForge worker

The PostgreSQL initialization script is mounted from `backend/migrations`. Docker Compose waits for PostgreSQL and RabbitMQ health checks before starting the API and worker.

Check the API health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

### 3. Run the frontend

Open a second terminal:

```bash
cd frontend
npm install
npm run dev
```

Open the local URL printed by Vite (usually `http://localhost:5173`).

The frontend uses `http://localhost:8080/api/v1` by default. To point it at another API URL, create a frontend `.env` file:

```env
VITE_API_URL=http://localhost:8080/api/v1
```

Restart the Vite development server after changing environment variables.

### 4. Stop the services

```bash
docker compose down
```

To also remove the local PostgreSQL volume and its stored data:

```bash
docker compose down -v
```

**Warning:** `docker compose down -v` permanently removes the Compose-managed database volume.

## Configuration

The backend reads these environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `DATABASE_URL` | `postgres://flowforge:flowforge@localhost:5433/flowforge?sslmode=disable` | PostgreSQL connection string |
| `REDIS_URL` | `localhost:6379` | Redis address |
| `AMQP_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection string |
| `PORT` | `8080` | API listening port |

Docker Compose overrides the connection values so that containers communicate using their Compose service names.

## API reference

Base URL: `http://localhost:8080/api/v1`

All request and response bodies use JSON.

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | API health check (outside the versioned base path) |
| `GET` | `/workflows` | List workflows (up to 100) |
| `POST` | `/workflows` | Create a workflow |
| `GET` | `/workflows/{workflow_id}` | Retrieve a workflow |
| `POST` | `/workflows/{workflow_id}/runs` | Queue a workflow run |
| `GET` | `/workflows/{workflow_id}/runs` | List runs for a workflow (up to 100) |
| `GET` | `/runs/{run_id}` | Retrieve a run |

### Create a workflow

`POST /api/v1/workflows`

Request body:

```json
{
  "name": "Example workflow",
  "description": "A small two-step workflow",
  "definition": {
    "steps": [
      {
        "id": "prepare",
        "name": "Prepare",
        "type": "set",
        "config": {
          "message": "Ready"
        }
      },
      {
        "id": "wait",
        "name": "Wait",
        "type": "delay",
        "depends_on": ["prepare"],
        "config": {
          "seconds": 1
        },
        "retry": {
          "max_attempts": 2,
          "backoff_seconds": 1
        },
        "timeout_seconds": 5
      }
    ]
  }
}
```

A successful creation returns HTTP `201` and the stored workflow, including its generated ID.

### Start a workflow run

`POST /api/v1/workflows/{workflow_id}/runs`

The request body is the input object available to steps. For example:

```json
{
  "customer_id": "customer-123",
  "priority": "high"
}
```

The body can be empty. A successful request returns HTTP `202` with the newly queued run.

### Inspect a run

`GET /api/v1/runs/{run_id}`

The response includes the run ID, workflow ID, status, input, output, error (when present), and creation time. A run is asynchronous: poll this endpoint to observe its status transition.

## Example workflow

The following definition uses a condition step to compare an input value. It demonstrates dependency ordering and the current condition behaviour; it does not implement conditional branching.

```json
{
  "steps": [
    {
      "id": "check_priority",
      "name": "Check priority",
      "type": "condition",
      "config": {
        "input_key": "priority",
        "equals": "high"
      }
    },
    {
      "id": "finish",
      "name": "Finish",
      "type": "noop",
      "depends_on": ["check_priority"]
    }
  ]
}
```

The executor runs `check_priority` before `finish`. The first step's result is included in the run output under its step ID.

## Data model

The initial database migration creates three tables:

- **`workflows`** — workflow name, description, JSONB definition, version, and timestamps.
- **`workflow_runs`** — workflow reference, status, JSONB input and output, error text, and execution timestamps.
- **`step_runs`** — a schema for per-step status, attempt, input/output, error, and timestamps.

The current worker updates workflow-level run records. Although `step_runs` is present in the schema, the current executor does not yet write per-step execution records.

## Reliability and delivery semantics

FlowForge demonstrates several distributed-systems concepts:

- Durable RabbitMQ queues and persistent messages.
- Asynchronous execution outside the HTTP request lifecycle.
- PostgreSQL-backed workflow definitions and run state.
- Redis leases to reduce concurrent processing of the same run.
- Retry and timeout handling at the step executor level.
- Dependency validation before execution.

**Delivery is at least once, not exactly once.** A worker may perform an operation and crash before acknowledging its RabbitMQ message. Redis leases reduce concurrent duplicate processing but do not guarantee exactly-once side effects. Future steps that call external services should use idempotency keys and, where appropriate, an outbox/inbox design or database fencing.

## Current capabilities and limitations

This repository is an MVP, and the following distinctions are important:

| Area | Current implementation |
|---|---|
| DAG validation | Implemented using topological sorting; rejects duplicate IDs, unknown dependencies, and cycles |
| DAG execution | Sequential topological execution; parallel step execution is not implemented |
| Retries and timeouts | Implemented in the step executor |
| Conditional logic | Comparison result is returned; dynamic branching is not implemented |
| HTTP tasks | Not enabled; the engine returns an explicit error for the `http` type |
| Step-level history | Table exists, but worker does not persist `step_runs` records |
| Cancellation | Database status supports `cancelled`; no cancellation API/worker flow yet |
| Dashboard | Lists workflows, creates a predefined two-step sample workflow, and displays runs; no visual DAG editor |
| Authentication and authorization | Not implemented |
| Exactly-once execution | Not guaranteed; queue delivery is at least once |

## Security notes

This project is intended for local development and learning, not deployment as-is to an untrusted public environment.

Before production use:

- Replace the example database and broker credentials with secrets.
- Do not expose PostgreSQL, Redis, or RabbitMQ management ports publicly.
- Add authentication and authorization to API endpoints.
- Restrict CORS to trusted frontend origins.
- Add request size limits, input validation, and rate limiting.
- Use TLS for network connections and secure secret management.
- Review worker lease expiry, crash recovery, and idempotency for external side effects.

## Repository structure

```text
flowforge/
├── backend/
│   ├── cmd/
│   │   ├── api/                 # HTTP API entry point
│   │   └── worker/              # Background worker entry point
│   ├── internal/
│   │   ├── config/              # Environment-based configuration
│   │   ├── domain/              # Workflow, step, run, and message types
│   │   ├── engine/              # DAG validation and step execution
│   │   ├── httpapi/             # HTTP handlers and CORS middleware
│   │   ├── queue/               # RabbitMQ integration
│   │   └── store/               # PostgreSQL persistence
│   ├── migrations/              # Database initialization SQL
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── App.tsx              # Workflow dashboard
│   │   ├── main.tsx             # React entry point
│   │   └── style.css            # Dashboard styling
│   ├── package.json
│   └── vite.config.ts
├── docs/
│   └── architecture.md          # Architecture and execution notes
└── docker-compose.yml
```

## Roadmap

Potential next steps for evolving the MVP:

- [ ] Persist step-level execution status, output, and retry attempts.
- [ ] Add parallel execution for independent DAG branches with concurrency limits.
- [ ] Implement conditional branching and explicit skip semantics.
- [ ] Add HTTP task support with safe timeouts, retries, and secret handling.
- [ ] Add workflow cancellation and graceful worker shutdown.
- [ ] Improve message redelivery, dead-letter handling, and idempotent processing.
- [ ] Add authentication, authorization, and audit logging.
- [ ] Build a visual workflow editor and richer execution timeline.
- [ ] Add automated unit, integration, and end-to-end tests.
- [ ] Add metrics, structured logs, and distributed tracing.

## Contributing

Contributions and suggestions are welcome.

1. Fork the repository.
2. Create a feature branch.
3. Make a focused change and add tests where appropriate.
4. Run the relevant checks locally.
5. Open a pull request describing the change and its behaviour.

## License

No license file was present in the repository at the time this README was updated. Unless a license is added, standard copyright applies and reuse/redistribution permissions are not explicitly granted.

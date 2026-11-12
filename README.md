# FlowForge

FlowForge is a distributed workflow automation engine built with Go, React, PostgreSQL, Redis, and RabbitMQ. It enables users to create, schedule, and execute multi-step workflows using DAG orchestration, state machines, conditional branching, retry policies, timeout handling, distributed locks, and fault recovery.

## Project status

This repository contains an MVP foundation for workflow definition, DAG validation and execution, persistence, queue integration, and a React dashboard. Review the architecture and implementation notes before treating it as production-ready.

## Quick start

1. Copy `.env.example` to `.env` and adjust values as needed.
2. Start backend dependencies and services with `docker compose up --build`.
3. In a separate terminal, run `cd frontend && npm install && npm run dev`.
4. Open http://localhost:5173.

See `docs/architecture.md` and `docs/sample-workflow.json` for architecture details and an example workflow.

## Technology

- Go API and worker
- React + TypeScript frontend
- PostgreSQL persistence
- Redis coordination
- RabbitMQ messaging
- Docker Compose

## License

No license has been specified yet. All rights remain with the copyright holder.
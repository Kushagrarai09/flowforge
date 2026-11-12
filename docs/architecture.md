# FlowForge architecture notes

## Request lifecycle

1. A client submits a workflow definition to the API.
2. The API validates unique step IDs, dependency references and acyclicity, then persists the definition.
3. A run request creates a `queued` row in PostgreSQL and publishes a persistent message to RabbitMQ.
4. A worker consumes the message and attempts to acquire a Redis lease for that run ID.
5. The worker loads the workflow and input, marks the run `running`, and executes the DAG in dependency order.
6. The worker stores the final output and terminal status, then acknowledges the queue message.

## State model

Workflow run states currently implemented:

```text
queued -> running -> succeeded
                  -> failed
```

The database schema reserves `cancelled` as a terminal state for a future cancellation endpoint. Step-level states and attempts are represented by the `step_runs` table but are not yet written by the current worker.

## DAG validation

The engine uses Kahn's topological sort. A workflow is rejected if it contains duplicate IDs, references a missing dependency, or has a cycle. The current executor processes the resulting topological order sequentially. This gives deterministic dependency correctness; bounded parallel execution is a planned enhancement.

## Delivery semantics and reliability

RabbitMQ uses durable queues and persistent messages. This provides at-least-once delivery, not exactly-once execution. A worker can crash after performing an external side effect but before acknowledging the message. Future external-effect steps must therefore use idempotency keys or an outbox/inbox pattern. Redis locks reduce concurrent duplicate work but are not a substitute for idempotent execution or database fencing.

## Data ownership

PostgreSQL is the source of truth for workflow definitions and run state. RabbitMQ transports execution requests; Redis coordinates active workers. Messages contain identifiers only, so the worker reloads the latest persisted definition and input rather than trusting a large serialized job payload.

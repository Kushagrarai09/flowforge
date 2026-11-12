CREATE TABLE IF NOT EXISTS workflows (
 id UUID PRIMARY KEY,
 name TEXT NOT NULL,
 description TEXT NOT NULL DEFAULT '',
 definition JSONB NOT NULL,
 version INTEGER NOT NULL DEFAULT 1,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS workflow_runs (
 id UUID PRIMARY KEY,
 workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
 status TEXT NOT NULL CHECK (status IN ('queued','running','succeeded','failed','cancelled')),
 input JSONB NOT NULL DEFAULT '{}'::jsonb,
 output JSONB NOT NULL DEFAULT '{}'::jsonb,
 error TEXT NOT NULL DEFAULT '',
 started_at TIMESTAMPTZ,
 finished_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS step_runs (
 id BIGSERIAL PRIMARY KEY,
 run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
 step_id TEXT NOT NULL,
 status TEXT NOT NULL,
 attempt INTEGER NOT NULL DEFAULT 0,
 input JSONB NOT NULL DEFAULT '{}'::jsonb,
 output JSONB NOT NULL DEFAULT '{}'::jsonb,
 error TEXT NOT NULL DEFAULT '',
 started_at TIMESTAMPTZ,
 finished_at TIMESTAMPTZ,
 UNIQUE(run_id, step_id)
);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow_created ON workflow_runs(workflow_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_step_runs_run ON step_runs(run_id);

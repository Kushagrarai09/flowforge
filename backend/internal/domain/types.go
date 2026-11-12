package domain

import "encoding/json"

type Step struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Type           string         `json:"type"`
	DependsOn      []string       `json:"depends_on,omitempty"`
	Config         map[string]any `json:"config,omitempty"`
	Retry          *RetryPolicy   `json:"retry,omitempty"`
	TimeoutSeconds int            `json:"timeout_seconds,omitempty"`
}
type RetryPolicy struct {
	MaxAttempts    int `json:"max_attempts"`
	BackoffSeconds int `json:"backoff_seconds"`
}
type Definition struct { Steps []Step `json:"steps"` }
type Workflow struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Definition  Definition `json:"definition"`
	Version     int        `json:"version"`
	CreatedAt   string     `json:"created_at,omitempty"`
}
type Run struct {
	ID         string          `json:"id"`
	WorkflowID string          `json:"workflow_id"`
	Status     string          `json:"status"`
	Input      json.RawMessage `json:"input"`
	Output     json.RawMessage `json:"output"`
	Error      string          `json:"error,omitempty"`
	CreatedAt  string          `json:"created_at,omitempty"`
}
type ExecuteMessage struct {
	RunID      string `json:"run_id"`
	WorkflowID string `json:"workflow_id"`
}

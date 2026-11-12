package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kushagrarai09/flowforge/internal/domain"
	"time"
)

type Result map[string]any

func Validate(d domain.Definition) error {
	if len(d.Steps) == 0 { return errors.New("workflow must contain at least one step") }
	ids := map[string]bool{}
	for _, s := range d.Steps {
		if s.ID == "" || s.Type == "" { return errors.New("each step requires id and type") }
		if ids[s.ID] { return fmt.Errorf("duplicate step id: %s", s.ID) }
		ids[s.ID] = true
	}
	for _, s := range d.Steps { for _, dep := range s.DependsOn { if !ids[dep] { return fmt.Errorf("step %s depends on unknown step %s", s.ID, dep) } } }
	if _, e := topological(d); e != nil { return e }
	return nil
}
func topological(d domain.Definition) ([]domain.Step, error) {
	by := map[string]domain.Step{}; degree := map[string]int{}; next := map[string][]string{}
	for _, s := range d.Steps { by[s.ID] = s; degree[s.ID] = len(s.DependsOn); for _, dep := range s.DependsOn { next[dep] = append(next[dep], s.ID) } }
	q := []string{}; for id, n := range degree { if n == 0 { q = append(q, id) } }
	out := []domain.Step{}
	for len(q) > 0 { id := q[0]; q = q[1:]; out = append(out, by[id]); for _, n := range next[id] { degree[n]--; if degree[n] == 0 { q = append(q, n) } } }
	if len(out) != len(d.Steps) { return nil, errors.New("workflow contains a dependency cycle") }
	return out, nil
}
func Execute(ctx context.Context, d domain.Definition, input map[string]any) (Result, error) {
	order, e := topological(d); if e != nil { return nil, e }; results := Result{}
	for _, s := range order {
		select { case <-ctx.Done(): return results, ctx.Err(); default: }
		for _, dep := range s.DependsOn { if _, ok := results[dep]; !ok { return results, fmt.Errorf("dependency %s did not complete", dep) } }
		attempts := 1; backoff := 1
		if s.Retry != nil { if s.Retry.MaxAttempts > 0 { attempts = s.Retry.MaxAttempts }; if s.Retry.BackoffSeconds > 0 { backoff = s.Retry.BackoffSeconds } }
		var value any; var runErr error
		for i := 0; i < attempts; i++ {
			stepCtx := ctx; cancel := func() {}
			if s.TimeoutSeconds > 0 { stepCtx, cancel = context.WithTimeout(ctx, time.Duration(s.TimeoutSeconds)*time.Second) }
			value, runErr = runStep(stepCtx, s, input, results); cancel()
			if runErr == nil { break }
			if i < attempts-1 { select { case <-ctx.Done(): return results, ctx.Err(); case <-time.After(time.Duration(backoff*(i+1))*time.Second): } }
		}
		if runErr != nil { return results, fmt.Errorf("step %s failed after %d attempt(s): %w", s.ID, attempts, runErr) }
		results[s.ID] = value
	}
	return results, nil
}
func runStep(ctx context.Context, s domain.Step, input map[string]any, previous Result) (any, error) {
	switch s.Type {
	case "noop": return map[string]any{"message":"completed"}, nil
	case "set": return s.Config, nil
	case "delay":
		seconds := 1.0; if v, ok := s.Config["seconds"].(float64); ok { seconds = v }
		select { case <-ctx.Done(): return nil, ctx.Err(); case <-time.After(time.Duration(seconds*float64(time.Second))): return map[string]any{"delayed_seconds":seconds}, nil }
	case "condition":
		key, _ := s.Config["input_key"].(string); expected := s.Config["equals"]; actual := input[key]
		return map[string]any{"matched":fmt.Sprint(actual)==fmt.Sprint(expected),"actual":actual}, nil
	case "http": return nil, errors.New("http step is not enabled in the MVP; use noop, set, delay or condition")
	default: return nil, fmt.Errorf("unsupported step type %q", s.Type)
	}
}
func Encode(v any) json.RawMessage { b, _ := json.Marshal(v); return b }

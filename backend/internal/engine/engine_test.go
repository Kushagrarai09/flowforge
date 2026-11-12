package engine

import (
	"context"
	"github.com/kushagrarai09/flowforge/internal/domain"
	"testing"
)

func TestValidateRejectsCycle(t *testing.T) {
	d := domain.Definition{Steps: []domain.Step{{ID:"a",Type:"noop",DependsOn:[]string{"b"}},{ID:"b",Type:"noop",DependsOn:[]string{"a"}}}}
	if Validate(d)==nil { t.Fatal("expected cycle error") }
}
func TestExecuteDAG(t *testing.T) {
	d := domain.Definition{Steps: []domain.Step{{ID:"first",Type:"set",Config:map[string]any{"value":"ok"}},{ID:"second",Type:"noop",DependsOn:[]string{"first"}}}}
	got,e := Execute(context.Background(),d,map[string]any{}); if e!=nil { t.Fatal(e) }; if _,ok:=got["second"]; !ok { t.Fatal("dependent step did not execute") }
}
func TestValidateRejectsMissingDependency(t *testing.T) {
	d := domain.Definition{Steps: []domain.Step{{ID:"a",Type:"noop",DependsOn:[]string{"missing"}}}}
	if Validate(d)==nil { t.Fatal("expected missing dependency error") }
}

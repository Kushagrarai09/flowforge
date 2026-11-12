package httpapi

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/kushagrarai09/flowforge/internal/domain"
	"github.com/kushagrarai09/flowforge/internal/engine"
	"github.com/kushagrarai09/flowforge/internal/queue"
	"github.com/kushagrarai09/flowforge/internal/store"
	"net/http"
	"strings"
	"time"
)

type Server struct { Store *store.Store; Queue *queue.Queue }
func (s *Server) Handler() http.Handler {
	m:=http.NewServeMux()
	m.HandleFunc("GET /health",func(w http.ResponseWriter,r *http.Request){write(w,200,map[string]string{"status":"ok"})})
	m.HandleFunc("/api/v1/workflows",s.workflows); m.HandleFunc("/api/v1/workflows/",s.workflow); m.HandleFunc("GET /api/v1/runs/",s.run)
	return cors(m)
}
func (s *Server) workflows(w http.ResponseWriter,r *http.Request) {
	switch r.Method {
	case "GET":
		v,e:=s.Store.ListWorkflows(r.Context()); if e!=nil {fail(w,500,e);return}; write(w,200,v)
	case "POST":
		var in struct { Name string `json:"name"`; Description string `json:"description"`; Definition domain.Definition `json:"definition"` }
		if e:=json.NewDecoder(r.Body).Decode(&in);e!=nil {fail(w,400,e);return}
		if strings.TrimSpace(in.Name)=="" {fail(w,400,"name is required");return}
		if e:=engine.Validate(in.Definition);e!=nil {fail(w,400,e);return}
		v:=domain.Workflow{ID:uuid.NewString(),Name:in.Name,Description:in.Description,Definition:in.Definition,Version:1}
		if e:=s.Store.CreateWorkflow(r.Context(),v);e!=nil {fail(w,500,e);return}; write(w,201,v)
	default: write(w,405,map[string]string{"error":"method not allowed"})
	}
}
func (s *Server) workflow(w http.ResponseWriter,r *http.Request) {
	parts:=strings.Split(strings.Trim(r.URL.Path,"/"),"/"); if len(parts)<4 {write(w,404,map[string]string{"error":"not found"});return}; id:=parts[3]
	if len(parts)==4&&r.Method=="GET" {v,e:=s.Store.GetWorkflow(r.Context(),id);if e!=nil {fail(w,404,e);return};write(w,200,v);return}
	if len(parts)==5&&parts[4]=="runs"&&r.Method=="POST" {
		if _,e:=s.Store.GetWorkflow(r.Context(),id);e!=nil {fail(w,404,e);return}
		var input map[string]any; if e:=json.NewDecoder(r.Body).Decode(&input);e!=nil&&e.Error()!="EOF" {fail(w,400,e);return};if input==nil {input=map[string]any{}}
		run:=domain.Run{ID:uuid.NewString(),WorkflowID:id,Status:"queued",Input:engine.Encode(input),Output:engine.Encode(map[string]any{})}
		if e:=s.Store.CreateRun(r.Context(),run);e!=nil {fail(w,500,e);return}
		if e:=s.Queue.Publish(domain.ExecuteMessage{RunID:run.ID,WorkflowID:id});e!=nil {_=s.Store.UpdateRun(r.Context(),run.ID,"failed",e.Error(),nil);fail(w,503,e);return}
		write(w,202,run);return
	}
	if len(parts)==5&&parts[4]=="runs"&&r.Method=="GET" {v,e:=s.Store.ListRuns(r.Context(),id);if e!=nil {fail(w,500,e);return};write(w,200,v);return}
	write(w,404,map[string]string{"error":"not found"})
}
func (s *Server) run(w http.ResponseWriter,r *http.Request) {id:=strings.TrimPrefix(r.URL.Path,"/api/v1/runs/");v,e:=s.Store.GetRun(r.Context(),id);if e!=nil {fail(w,404,e);return};write(w,200,v)}
func write(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_=json.NewEncoder(w).Encode(v)}
func fail(w http.ResponseWriter,status int,e any){write(w,status,map[string]any{"error":e})}
func cors(next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.Header().Set("Access-Control-Allow-Origin","*");w.Header().Set("Access-Control-Allow-Headers","Content-Type, Authorization");w.Header().Set("Access-Control-Allow-Methods","GET, POST, OPTIONS");if r.Method=="OPTIONS"{w.WriteHeader(204);return};next.ServeHTTP(w,r)})}
var _=context.Background
var _=time.Second

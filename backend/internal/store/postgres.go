package store

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kushagrarai09/flowforge/internal/domain"
)
type Store struct{DB *pgxpool.Pool}
func New(ctx context.Context,url string)(*Store,error){p,e:=pgxpool.New(ctx,url);if e!=nil{return nil,e};if e=p.Ping(ctx);e!=nil{p.Close();return nil,e};return &Store{p},nil}
func(s *Store)CreateWorkflow(ctx context.Context,w domain.Workflow)error{b,_:=json.Marshal(w.Definition);_,e:=s.DB.Exec(ctx,`INSERT INTO workflows(id,name,description,definition) VALUES($1,$2,$3,$4)`,w.ID,w.Name,w.Description,b);return e}
func(s *Store)GetWorkflow(ctx context.Context,id string)(domain.Workflow,error){var w domain.Workflow;var b []byte;e:=s.DB.QueryRow(ctx,`SELECT id,name,description,definition,version,created_at::text FROM workflows WHERE id=$1`,id).Scan(&w.ID,&w.Name,&w.Description,&b,&w.Version,&w.CreatedAt);if e!=nil{return w,e};e=json.Unmarshal(b,&w.Definition);return w,e}
func(s *Store)ListWorkflows(ctx context.Context)([]domain.Workflow,error){rows,e:=s.DB.Query(ctx,`SELECT id,name,description,definition,version,created_at::text FROM workflows ORDER BY created_at DESC LIMIT 100`);if e!=nil{return nil,e};defer rows.Close();out:=[]domain.Workflow{};for rows.Next(){var w domain.Workflow;var b []byte;if e=rows.Scan(&w.ID,&w.Name,&w.Description,&b,&w.Version,&w.CreatedAt);e!=nil{return nil,e};if e=json.Unmarshal(b,&w.Definition);e!=nil{return nil,e};out=append(out,w)};return out,rows.Err()}
func(s *Store)CreateRun(ctx context.Context,r domain.Run)error{_,e:=s.DB.Exec(ctx,`INSERT INTO workflow_runs(id,workflow_id,status,input) VALUES($1,$2,$3,$4)`,r.ID,r.WorkflowID,r.Status,r.Input);return e}
func(s *Store)UpdateRun(ctx context.Context,id,status,errText string,out any)error{b,_:=json.Marshal(out);_,e:=s.DB.Exec(ctx,`UPDATE workflow_runs SET status=$2,error=$3,output=$4,started_at=COALESCE(started_at,now()),finished_at=CASE WHEN $2 IN ('succeeded','failed','cancelled') THEN now() ELSE finished_at END WHERE id=$1`,id,status,errText,b);return e}
func(s *Store)GetRun(ctx context.Context,id string)(domain.Run,error){var r domain.Run;e:=s.DB.QueryRow(ctx,`SELECT id,workflow_id,status,input,output,error,created_at::text FROM workflow_runs WHERE id=$1`,id).Scan(&r.ID,&r.WorkflowID,&r.Status,&r.Input,&r.Output,&r.Error,&r.CreatedAt);return r,e}
func(s *Store)ListRuns(ctx context.Context,wid string)([]domain.Run,error){rows,e:=s.DB.Query(ctx,`SELECT id,workflow_id,status,input,output,error,created_at::text FROM workflow_runs WHERE workflow_id=$1 ORDER BY created_at DESC LIMIT 100`,wid);if e!=nil{return nil,e};defer rows.Close();out:=[]domain.Run{};for rows.Next(){var r domain.Run;if e=rows.Scan(&r.ID,&r.WorkflowID,&r.Status,&r.Input,&r.Output,&r.Error,&r.CreatedAt);e!=nil{return nil,e};out=append(out,r)};return out,rows.Err()}
var ErrNotFound=errors.New("not found")

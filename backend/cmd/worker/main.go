package main

import (
	"context"
	"encoding/json"
	"github.com/kushagrarai09/flowforge/internal/config"
	"github.com/kushagrarai09/flowforge/internal/domain"
	"github.com/kushagrarai09/flowforge/internal/engine"
	"github.com/kushagrarai09/flowforge/internal/queue"
	"github.com/kushagrarai09/flowforge/internal/store"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	db, e := store.New(ctx, cfg.DatabaseURL)
	if e != nil {
		log.Fatal(e)
	}
	defer db.DB.Close()
	q, e := queue.Connect(cfg.AMQPURL)
	if e != nil {
		log.Fatal(e)
	}
	defer q.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	defer redisClient.Close()
	deliveries, e := q.Consume()
	if e != nil {
		log.Fatal(e)
	}
	log.Println("FlowForge worker started")
	for d := range deliveries {
		var msg domain.ExecuteMessage
		if e = json.Unmarshal(d.Body, &msg); e != nil {
			_ = d.Nack(false, false)
			continue
		}
		lock := redis.NewScript(`if redis.call('SET',KEYS[1],ARGV[1],'NX','EX',ARGV[2]) then return 1 else return 0 end`)
		locked, e := lock.Run(ctx, redisClient, []string{"flowforge:run:" + msg.RunID, msg.RunID, "3600"}).Int()
		if e != nil || locked == 0 {
			_ = d.Nack(false, true)
			time.Sleep(time.Second)
			continue
		}
		func() {
			defer redisClient.Del(ctx, "flowforge:run:"+msg.RunID)
			run, e := db.GetRun(ctx, msg.RunID)
			if e != nil {
				_ = d.Ack(false)
				return
			}
			if run.Status == "succeeded" || run.Status == "failed" || run.Status == "cancelled" {
				_ = d.Ack(false)
				return
			}
			w, e := db.GetWorkflow(ctx, msg.WorkflowID)
			if e != nil {
				_ = db.UpdateRun(ctx, msg.RunID, "failed", e.Error(), nil)
				_ = d.Ack(false)
				return
			}
			var input map[string]any
			_ = json.Unmarshal(run.Input, &input)
			_ = db.UpdateRun(ctx, msg.RunID, "running", "", map[string]any{})
			result, e := engine.Execute(ctx, w.Definition, input)
			if e != nil {
				_ = db.UpdateRun(ctx, msg.RunID, "failed", e.Error(), result)
			} else {
				_ = db.UpdateRun(ctx, msg.RunID, "succeeded", "", result)
			}
			_ = d.Ack(false)
		}()
	}
}

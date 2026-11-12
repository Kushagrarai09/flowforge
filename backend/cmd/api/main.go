package main

import (
	"context"
	"github.com/kushagrarai09/flowforge/internal/config"
	"github.com/kushagrarai09/flowforge/internal/httpapi"
	"github.com/kushagrarai09/flowforge/internal/queue"
	"github.com/kushagrarai09/flowforge/internal/store"
	"log"
	"net/http"
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
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: (&httpapi.Server{Store: db, Queue: q}).Handler(), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("FlowForge API listening on :%s", cfg.Port)
	log.Fatal(srv.ListenAndServe())
}

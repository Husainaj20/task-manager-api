package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/husainaj20/task-manager-api/internal/api"
	"github.com/husainaj20/task-manager-api/internal/service"
	"github.com/husainaj20/task-manager-api/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var st store.Store
	if os.Getenv("STORE") == "redis" {
		addr := os.Getenv("REDIS_ADDR")
		if addr == "" {
			addr = "localhost:6379"
		}
		r := store.NewRedisStore(addr, "taskmgr")
		st = r
	} else {
		ms := store.NewMemoryStore()
		st = ms
	}

	queue := service.NewQueue(8) // 8 workers by default
	defer queue.Stop()

	queue.SetProcessor(func(ctx context.Context, t *service.TaskWork) error {
		time.Sleep(150 * time.Millisecond)
		return st.UpdateStatus(ctx, t.ID, "done", t.Result)
	})

	h := api.New(st, queue)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: h.Router(),
	}

	go func() {
		log.Printf("server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	log.Println("server exited")
}

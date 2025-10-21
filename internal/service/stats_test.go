package service

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestStats_CountersAccurate(t *testing.T) {
	q := NewQueue(3)
	q.ConfigureRetry(3, 5*time.Millisecond, 2.0, 50*time.Millisecond, false)
	defer q.Stop()

	// behavior: first two tasks fail once then succeed, last one always fails to DLQ
	successAfter := map[string]int{"t1": 2, "t2": 2}
	atomicProcessed := int32(0)
	atomicDLQ := int32(0)

	q.SetDLQHandler(func(id string) {
		atomic.AddInt32(&atomicDLQ, 1)
	})

	q.SetProcessor(func(ctx context.Context, tw *TaskWork) error {
		if want, ok := successAfter[tw.ID]; ok {
			if tw.Attempts+1 >= want {
				atomic.AddInt32(&atomicProcessed, 1)
				return nil
			}
			return context.DeadlineExceeded
		}
		// default: always fail
		return context.DeadlineExceeded
	})

	q.Enqueue(&TaskWork{ID: "t1"})
	q.Enqueue(&TaskWork{ID: "t2"})
	q.Enqueue(&TaskWork{ID: "t-dlq"})

	ok := q.WaitIdle(2 * time.Second)
	if !ok {
		t.Fatalf("queue did not become idle")
	}

	_, inflight, processed, _, dlq := q.Stats()
	if inflight != 0 {
		t.Fatalf("expected inflight 0, got %d", inflight)
	}
	if processed != int64(atomic.LoadInt32(&atomicProcessed)) {
		t.Fatalf("processed mismatch: got %d, want %d", processed, atomic.LoadInt32(&atomicProcessed))
	}
	if dlq != int64(atomic.LoadInt32(&atomicDLQ)) {
		t.Fatalf("dlq mismatch: got %d, want %d", dlq, atomic.LoadInt32(&atomicDLQ))
	}
}

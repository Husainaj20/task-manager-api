package service

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueue_ProcessesTasks(t *testing.T) {
	var processed int32
	q := NewQueue(4)
	q.SetProcessor(func(ctx context.Context, t *TaskWork) error {
		atomic.AddInt32(&processed, 1)
		return nil
	})

	for i := 0; i < 10; i++ {
		q.Enqueue(&TaskWork{ID: "task"})
	}

	time.Sleep(100 * time.Millisecond)
	q.Stop()

	if processed != 10 {
		t.Errorf("expected 10 tasks processed, got %d", processed)
	}
}

func TestQueue_StopDrainsSafely(t *testing.T) {
	var processed int32
	q := NewQueue(2)
	q.SetProcessor(func(ctx context.Context, t *TaskWork) error {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&processed, 1)
		return nil
	})

	for i := 0; i < 5; i++ {
		q.Enqueue(&TaskWork{ID: "t"})
	}

	q.Stop()

	if processed != 5 {
		t.Errorf("expected 5 tasks processed after Stop, got %d", processed)
	}
}

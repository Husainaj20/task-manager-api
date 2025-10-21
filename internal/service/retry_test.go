package service

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetry_SucceedsBeforeMax(t *testing.T) {
	q := NewQueue(2)
	q.ConfigureRetry(4, 5*time.Millisecond, 2.0, 50*time.Millisecond, false)
	defer q.Stop()

	var attempts int32
	q.SetProcessor(func(ctx context.Context, tw *TaskWork) error {
		atomic.AddInt32(&attempts, 1)
		if atomic.LoadInt32(&attempts) <= 2 {
			return context.DeadlineExceeded
		}
		return nil
	})

	q.Enqueue(&TaskWork{ID: "t1"})
	// Poll attempts until expected or timeout
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&attempts) >= 3 {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Fatalf("expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

func TestRetry_ExceedsMax_GoesToDLQ(t *testing.T) {
	q := NewQueue(1)
	q.ConfigureRetry(3, 5*time.Millisecond, 2.0, 50*time.Millisecond, false)
	defer q.Stop()

	var failed int32
	var dlqCalled int32
	q.SetDLQHandler(func(id string) {
		atomic.AddInt32(&dlqCalled, 1)
	})
	q.SetProcessor(func(ctx context.Context, tw *TaskWork) error {
		atomic.AddInt32(&failed, 1)
		return context.DeadlineExceeded
	})

	q.Enqueue(&TaskWork{ID: "t-dlq"})
	dl := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(dl) {
		if atomic.LoadInt32(&dlqCalled) >= 1 {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	if atomic.LoadInt32(&dlqCalled) != 1 {
		t.Fatalf("expected dlq called once, got %d", atomic.LoadInt32(&dlqCalled))
	}
}

func TestBackoff_DelaysGrow(t *testing.T) {
	q := NewQueue(1)
	q.ConfigureRetry(4, 5*time.Millisecond, 2.0, 40*time.Millisecond, false)
	defer q.Stop()

	var mu2 sync.Mutex
	var attempts2 int32
	timestamps := []time.Time{}
	q.SetProcessor(func(ctx context.Context, tw *TaskWork) error {
		mu2.Lock()
		timestamps = append(timestamps, time.Now())
		mu2.Unlock()
		atomic.AddInt32(&attempts2, 1)
		if atomic.LoadInt32(&attempts2) < 4 {
			return context.DeadlineExceeded
		}
		return nil
	})

	q.Enqueue(&TaskWork{ID: "t-back"})
	dl2 := time.Now().Add(2 * time.Second)
	for time.Now().Before(dl2) {
		if atomic.LoadInt32(&attempts2) >= 4 {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	if atomic.LoadInt32(&attempts2) < 4 {
		t.Fatalf("expected 4 attempts, got %d", atomic.LoadInt32(&attempts2))
	}

	mu2.Lock()
	if len(timestamps) != 4 {
		mu2.Unlock()
		t.Fatalf("expected 4 attempts, got %d", len(timestamps))
	}
	d1 := timestamps[1].Sub(timestamps[0])
	d2 := timestamps[2].Sub(timestamps[1])
	d3 := timestamps[3].Sub(timestamps[2])
	mu2.Unlock()

	if d1 < 4*time.Millisecond || d1 > 50*time.Millisecond {
		t.Fatalf("unexpected d1 %v", d1)
	}
	if d2 < 8*time.Millisecond || d2 > 100*time.Millisecond {
		t.Fatalf("unexpected d2 %v", d2)
	}
	if d3 < 16*time.Millisecond || d3 > 200*time.Millisecond {
		t.Fatalf("unexpected d3 %v", d3)
	}
}

func TestStop_WhileRetriesScheduled(t *testing.T) {
	q := NewQueue(1)
	q.ConfigureRetry(3, 5*time.Millisecond, 2.0, 50*time.Millisecond, false)

	q.SetProcessor(func(ctx context.Context, tw *TaskWork) error {
		return context.DeadlineExceeded
	})

	q.Enqueue(&TaskWork{ID: "t-stop"})
	// give it a moment to schedule retries
	time.Sleep(10 * time.Millisecond)
	q.Stop()
	// If Stop returns, test is successful (no panic/hang)
}

package service

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type TaskWork struct {
	ID       string
	Result   map[string]any
	Attempts int
}

type Processor func(ctx context.Context, t *TaskWork) error

// DLQHandler is called when a task exceeds max attempts
type DLQHandler func(id string)

type Queue struct {
	wg        sync.WaitGroup
	work      chan *TaskWork
	stopOnce  sync.Once
	cancel    context.CancelFunc
	processor Processor

	// retry/scheduling
	retryMu sync.Mutex
	timers  map[string]*time.Timer

	// config
	maxAttempts int
	baseBackoff time.Duration
	factor      float64
	maxBackoff  time.Duration
	jitter      bool
	dlqHandler  DLQHandler

	// stats
	// stats
	inflight  int64
	processed int64
	failed    int64
	dlq       int64
}

func NewQueue(workers int) *Queue {
	q := &Queue{
		work:        make(chan *TaskWork, 1024),
		timers:      make(map[string]*time.Timer),
		maxAttempts: 3,
		baseBackoff: 50 * time.Millisecond,
		factor:      2.0,
		maxBackoff:  5 * time.Second,
		jitter:      true,
	}
	ctx, cancel := context.WithCancel(context.Background())
	q.cancel = cancel

	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go func(i int) {
			defer q.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case w, ok := <-q.work:
					if !ok {
						return
					}
					if w == nil {
						continue
					}
					atomic.AddInt64(&q.inflight, 1)
					if q.processor != nil {
						if err := q.processor(ctx, w); err != nil {
							// handle retry
							q.handleRetry(ctx, w)
							atomic.AddInt64(&q.failed, 1)
						} else {
							atomic.AddInt64(&q.processed, 1)
						}
					}
					atomic.AddInt64(&q.inflight, -1)
				}
			}
		}(i)
	}
	return q
}

func (q *Queue) SetProcessor(p Processor) { q.processor = p }

// ConfigureRetry sets retry/backoff parameters
func (q *Queue) ConfigureRetry(maxAttempts int, base time.Duration, factor float64, max time.Duration, jitter bool) {
	q.maxAttempts = maxAttempts
	q.baseBackoff = base
	q.factor = factor
	q.maxBackoff = max
	q.jitter = jitter
}

func (q *Queue) SetDLQHandler(h DLQHandler) { q.dlqHandler = h }

func (q *Queue) handleRetry(ctx context.Context, t *TaskWork) {
	t.Attempts++
	if t.Attempts >= q.maxAttempts {
		atomic.AddInt64(&q.dlq, 1)
		if q.dlqHandler != nil {
			// call synchronously so caller can observe DLQ handling completion
			q.dlqHandler(t.ID)
		}
		return
	}
	// compute backoff
	backoff := float64(q.baseBackoff) * math.Pow(q.factor, float64(t.Attempts-1))
	if backoff > float64(q.maxBackoff) {
		backoff = float64(q.maxBackoff)
	}
	d := time.Duration(backoff)

	q.retryMu.Lock()
	timer := time.AfterFunc(d, func() {
		// enqueue again unless stopped
		select {
		case <-ctx.Done():
			// canceled
		default:
			q.Enqueue(t)
		}
		q.retryMu.Lock()
		delete(q.timers, t.ID)
		q.retryMu.Unlock()
	})
	q.timers[t.ID] = timer
	q.retryMu.Unlock()
}

func (q *Queue) Enqueue(t *TaskWork) {
	select {
	case q.work <- t:
	default:
		q.work <- t
	}
}

// WaitIdle waits until queue has no queued items, no inflight, and no timers (retries)
func (q *Queue) WaitIdle(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		q.retryMu.Lock()
		timersEmpty := len(q.timers) == 0
		q.retryMu.Unlock()
		queued := int64(len(q.work))
		if queued == 0 && atomic.LoadInt64(&q.inflight) == 0 && timersEmpty {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

func (q *Queue) Stop() {
	q.stopOnce.Do(func() {
		// stop all retry timers
		q.retryMu.Lock()
		for _, t := range q.timers {
			t.Stop()
		}
		q.timers = map[string]*time.Timer{}
		q.retryMu.Unlock()
		// close work channel to let workers drain remaining items
		close(q.work)
		q.wg.Wait()
		// finally cancel any context to free resources
		q.cancel()
	})
}

// Stats returns basic counters
func (q *Queue) Stats() (queued, inflight, processed, failed, dlq int64) {
	q.retryMu.Lock()
	timersCount := int64(len(q.timers))
	q.retryMu.Unlock()
	queued = int64(len(q.work)) + timersCount
	return queued, atomic.LoadInt64(&q.inflight), atomic.LoadInt64(&q.processed), atomic.LoadInt64(&q.failed), atomic.LoadInt64(&q.dlq)
}

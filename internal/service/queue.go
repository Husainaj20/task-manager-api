package service

import (
	"context"
	"log"
	"sync"
)

type TaskWork struct {
	ID     string
	Result map[string]any
}

type Processor func(ctx context.Context, t *TaskWork) error

type Queue struct {
	wg        sync.WaitGroup
	work      chan *TaskWork
	stopOnce  sync.Once
	cancel    context.CancelFunc
	processor Processor
}

func NewQueue(workers int) *Queue {
	q := &Queue{
		work: make(chan *TaskWork, 1024),
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
				case w := <-q.work:
					if w == nil { continue }
					if q.processor != nil {
						if err := q.processor(ctx, w); err != nil {
							log.Printf("worker %d error: %v", i, err)
						}
					}
				}
			}
		}(i)
	}
	return q
}

func (q *Queue) SetProcessor(p Processor) { q.processor = p }

func (q *Queue) Enqueue(t *TaskWork) {
	select {
	case q.work <- t:
	default:
		q.work <- t
	}
}

func (q *Queue) Stop() {
	q.stopOnce.Do(func() {
		q.cancel()
		close(q.work)
		q.wg.Wait()
	})
}

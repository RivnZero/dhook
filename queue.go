package dhook

import (
	"context"
	"sync"
)

type Queue struct {
	client      *Client
	jobs        chan func()
	workerCount int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	running     bool
	mu          sync.Mutex
}

func NewQueue(client *Client, workerCount int) *Queue {
	if workerCount <= 0 {
		workerCount = 5
	}
	return &Queue{
		client:      client,
		jobs:        make(chan func(), 10000),
		workerCount: workerCount,
	}
}

func (q *Queue) Start(ctx context.Context) {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return
	}
	q.ctx, q.cancel = context.WithCancel(ctx)
	q.running = true
	q.mu.Unlock()

	for i := 0; i < q.workerCount; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

func (q *Queue) worker() {
	defer q.wg.Done()
	for {
		select {
		case <-q.ctx.Done():
			for {
				select {
				case <-q.jobs:
				default:
					return
				}
			}
		case fn, ok := <-q.jobs:
			if !ok {
				return
			}
			fn()
		}
	}
}

func (q *Queue) Add(msg *Message) {
	q.jobs <- func() {
		q.client.Send(q.ctx, msg)
	}
}

func (q *Queue) AddFunc(fn func()) {
	q.jobs <- fn
}

func (q *Queue) Stop() {
	q.mu.Lock()
	if !q.running {
		q.mu.Unlock()
		return
	}
	q.cancel()
	q.running = false
	q.mu.Unlock()
	q.wg.Wait()
}

func (q *Queue) Len() int {
	return len(q.jobs)
}

func (q *Queue) Cap() int {
	return cap(q.jobs)
}

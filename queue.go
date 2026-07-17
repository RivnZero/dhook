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
	submitters  sync.WaitGroup
	running     bool
	stopped     bool
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
	if q.running || q.stopped {
		q.mu.Unlock()
		return
	}
	q.ctx, q.cancel = context.WithCancel(ctx)
	q.running = true
	q.wg.Add(q.workerCount)
	q.mu.Unlock()

	for i := 0; i < q.workerCount; i++ {
		go q.worker()
	}
}

func (q *Queue) worker() {
	defer q.wg.Done()
	for fn := range q.jobs {
		fn()
	}
}

func (q *Queue) Add(msg *Message) {
	q.mu.Lock()
	if q.stopped {
		q.mu.Unlock()
		return
	}
	q.submitters.Add(1)
	q.mu.Unlock()
	defer q.submitters.Done()

	q.jobs <- func() {
		q.client.Send(q.ctx, msg)
	}
}

func (q *Queue) AddFunc(fn func()) {
	q.mu.Lock()
	if q.stopped {
		q.mu.Unlock()
		return
	}
	q.submitters.Add(1)
	q.mu.Unlock()
	defer q.submitters.Done()

	q.jobs <- fn
}

func (q *Queue) Stop() {
	q.mu.Lock()
	if !q.running {
		q.mu.Unlock()
		return
	}
	q.running = false
	q.stopped = true
	q.mu.Unlock()
	q.submitters.Wait()
	close(q.jobs)
	q.wg.Wait()
	q.cancel()
}

func (q *Queue) Len() int {
	return len(q.jobs)
}

func (q *Queue) Cap() int {
	return cap(q.jobs)
}

package dhook

import (
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueueStopDrainsAcceptedJobs(t *testing.T) {
	queue := NewQueue(nil, 1)
	queue.Start(context.Background())

	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	queue.AddFunc(func() {
		close(firstStarted)
		<-releaseFirst
	})
	<-firstStarted

	const acceptedJobs = 1_000
	var ran atomic.Int32
	for i := 0; i < acceptedJobs; i++ {
		queue.AddFunc(func() { ran.Add(1) })
	}

	stopReturned := make(chan struct{})
	stopStarted := make(chan struct{})
	go func() {
		close(stopStarted)
		queue.Stop()
		close(stopReturned)
	}()
	<-stopStarted
	close(releaseFirst)
	select {
	case <-stopReturned:
	case <-time.After(time.Second):
		t.Fatal("Stop did not return after draining accepted jobs")
	}
	if got := ran.Load(); got != acceptedJobs {
		t.Fatalf("Stop returned after running %d of %d accepted jobs", got, acceptedJobs)
	}
}

func TestQueueIgnoresAddsAfterStop(t *testing.T) {
	queue := NewQueue(nil, 1)
	queue.Start(context.Background())
	queue.Stop()

	var ran atomic.Int32
	queue.AddFunc(func() { ran.Add(1) })
	queue.Add(&Message{})

	if got := ran.Load(); got != 0 {
		t.Fatalf("post-stop function ran %d times", got)
	}
}

func TestQueueStartUnblocksSubmissionWhenFull(t *testing.T) {
	queue := NewQueue(nil, 1)
	for i := 0; i < queue.Cap(); i++ {
		queue.AddFunc(func() {})
	}

	submitted := make(chan struct{})
	go func() {
		queue.AddFunc(func() { close(submitted) })
	}()
	runtime.Gosched()

	queue.Start(context.Background())
	select {
	case <-submitted:
	case <-time.After(time.Second):
		t.Fatal("Start did not unblock a submission waiting on a full queue")
	}
	queue.Stop()
}

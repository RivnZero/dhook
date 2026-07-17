package dhook

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestRateLimiterRetryAfterFraction(t *testing.T) {
	const url = "https://discord.com/api/webhooks/test"

	limiter := NewRateLimiter()
	limiter.HandleResponse(url, http.StatusTooManyRequests, http.Header{"Retry-After": {"0.01"}})

	if retryAfter := limiter.getLastRetryAfter(url); retryAfter < 5*time.Millisecond || retryAfter > 100*time.Millisecond {
		t.Fatalf("Retry-After 0.01 produced %s; want a short fractional delay", retryAfter)
	}

	done := make(chan struct{})
	go func() {
		limiter.Wait(context.Background(), url)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("rate limiter did not unblock after the fractional Retry-After delay")
	}
}

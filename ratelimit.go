package dhook

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimiter struct {
	mu     sync.Mutex
	limits map[string]*limitState
}

type limitState struct {
	mu           sync.Mutex
	remaining    int
	resetAt      time.Time
	blockedUntil time.Time
	lastRetry    time.Duration
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*limitState),
	}
}

func (r *RateLimiter) getState(url string) *limitState {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.limits[url]
	if !ok {
		s = &limitState{remaining: -1}
		r.limits[url] = s
	}
	return s
}

func (r *RateLimiter) Wait(ctx context.Context, url string) {
	s := r.getState(url)
	s.mu.Lock()
	defer s.mu.Unlock()

	for {
		if ctx.Err() != nil {
			return
		}

		now := time.Now()
		var waitUntil time.Time

		if !s.blockedUntil.IsZero() && s.blockedUntil.After(now) {
			waitUntil = s.blockedUntil
		} else if s.remaining == 0 && !s.resetAt.IsZero() && s.resetAt.After(now) {
			waitUntil = s.resetAt
		}

		if waitUntil.IsZero() || !waitUntil.After(now) {
			break
		}

		duration := time.Until(waitUntil)
		s.mu.Unlock()

		timer := time.NewTimer(duration)
		select {
		case <-ctx.Done():
			timer.Stop()
			s.mu.Lock()
			return
		case <-timer.C:
		}
		s.mu.Lock()
	}

	if s.remaining > 0 {
		s.remaining--
	}
}

func (r *RateLimiter) HandleResponse(url string, statusCode int, headers http.Header) {
	s := r.getState(url)
	s.mu.Lock()
	defer s.mu.Unlock()

	if statusCode == 429 {
		d := parseRetryAfter(headers)
		s.blockedUntil = time.Now().Add(d)
		s.lastRetry = d
		s.remaining = 0
		return
	}

	if statusCode >= 500 {
		if s.blockedUntil.IsZero() || time.Now().After(s.blockedUntil) {
			s.blockedUntil = time.Now().Add(time.Second)
		} else {
			newDur := time.Until(s.blockedUntil) * 2
			if newDur > 30*time.Second {
				newDur = 30 * time.Second
			}
			s.blockedUntil = time.Now().Add(newDur)
		}
		return
	}

	s.blockedUntil = time.Time{}

	if v := headers.Get("X-Rate-Limit-Remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.remaining = n
		}
	}

	if v := headers.Get("X-Rate-Limit-Reset-After"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			s.resetAt = time.Now().Add(time.Duration(f * float64(time.Second)))
		}
	}
}

func (r *RateLimiter) getLastRetryAfter(url string) time.Duration {
	s := r.getState(url)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastRetry
}

func (r *RateLimiter) Reset(url string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.limits, url)
}

func (r *RateLimiter) ResetAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.limits = make(map[string]*limitState)
}

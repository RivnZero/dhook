package dhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type EventType int

const (
	EventSuccess EventType = iota
	EventRateLimit
	EventError
)

type SuccessFunc func(response *Response)
type RateLimitFunc func(retryAfter time.Duration)
type ErrorFunc func(err error)

type Message struct {
	Content         string           `json:"content,omitempty"`
	Username        string           `json:"username,omitempty"`
	AvatarURL       string           `json:"avatar_url,omitempty"`
	TTS             bool             `json:"tts,omitempty"`
	Embeds          []*Embed         `json:"embeds,omitempty"`
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`
}

type AllowedMentions struct {
	Parse []string `json:"parse,omitempty"`
	Roles []string `json:"roles,omitempty"`
	Users []string `json:"users,omitempty"`
}

type Response struct {
	ID             string   `json:"id"`
	ChannelID      string   `json:"channel_id"`
	GuildID        string   `json:"guild_id,omitempty"`
	WebhookID      string   `json:"webhook_id,omitempty"`
	Content        string   `json:"content"`
	Embeds         []*Embed `json:"embeds,omitempty"`
	Timestamp      string   `json:"timestamp,omitempty"`
}

type Client struct {
	urls          []string
	httpClient    *http.Client
	rateLimiter   *RateLimiter
	successHooks  []SuccessFunc
	rateLimitHooks []RateLimitFunc
	errorHooks    []ErrorFunc
	hooksMu       sync.RWMutex
}

func New(urls ...string) *Client {
	return &Client{
		urls:        urls,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		rateLimiter: NewRateLimiter(),
	}
}

func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

func (c *Client) AddHook(event EventType, fn interface{}) {
	c.hooksMu.Lock()
	defer c.hooksMu.Unlock()
	switch event {
	case EventSuccess:
		if f, ok := fn.(SuccessFunc); ok {
			c.successHooks = append(c.successHooks, f)
		}
	case EventRateLimit:
		if f, ok := fn.(RateLimitFunc); ok {
			c.rateLimitHooks = append(c.rateLimitHooks, f)
		}
	case EventError:
		if f, ok := fn.(ErrorFunc); ok {
			c.errorHooks = append(c.errorHooks, f)
		}
	}
}

func (c *Client) fireSuccess(resp *Response) {
	c.hooksMu.RLock()
	defer c.hooksMu.RUnlock()
	for _, h := range c.successHooks {
		h(resp)
	}
}

func (c *Client) fireRateLimit(d time.Duration) {
	c.hooksMu.RLock()
	defer c.hooksMu.RUnlock()
	for _, h := range c.rateLimitHooks {
		h(d)
	}
}

func (c *Client) fireError(err error) {
	c.hooksMu.RLock()
	defer c.hooksMu.RUnlock()
	for _, h := range c.errorHooks {
		h(err)
	}
}

func (c *Client) Send(ctx context.Context, msg *Message) ([]*Response, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	var (
		responses []*Response
		mu        sync.Mutex
		errs      []error
		wg        sync.WaitGroup
	)

	for _, url := range c.urls {
		wg.Add(1)
		go func(webhookURL string) {
			defer wg.Done()
			resp, err := c.doPost(ctx, webhookURL, "application/json", data)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				c.fireError(err)
				return
			}
			mu.Lock()
			responses = append(responses, resp)
			mu.Unlock()
			c.fireSuccess(resp)
		}(url)
	}

	wg.Wait()

	if len(errs) > 0 {
		return responses, errors.Join(errs...)
	}
	return responses, nil
}

func (c *Client) doPost(ctx context.Context, url, contentType string, data []byte) (*Response, error) {
	maxRetries := 5

	for attempt := 0; attempt <= maxRetries; attempt++ {
		c.rateLimiter.Wait(ctx, url)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("User-Agent", "dhook/1.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		c.rateLimiter.HandleResponse(url, resp.StatusCode, resp.Header)

		if resp.StatusCode == 429 {
			c.fireRateLimit(c.rateLimiter.getLastRetryAfter(url))
			continue
		}

		if resp.StatusCode >= 500 && attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("dhook: webhook returned status %d: %s", resp.StatusCode, string(respBody))
		}

		var r Response
		if err := json.Unmarshal(respBody, &r); err != nil {
			return nil, err
		}
		return &r, nil
	}

	return nil, errors.New("dhook: max retries exceeded")
}

func (c *Client) Edit(ctx context.Context, messageID string, msg *Message) (*Response, error) {
	if len(c.urls) == 0 {
		return nil, errors.New("dhook: no webhook URLs configured")
	}
	return c.editMessage(ctx, c.urls[0], messageID, msg)
}

func (c *Client) editMessage(ctx context.Context, url, messageID string, msg *Message) (*Response, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	editURL := url + "/messages/" + messageID
	maxRetries := 5

	for attempt := 0; attempt <= maxRetries; attempt++ {
		c.rateLimiter.Wait(ctx, editURL)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPatch, editURL, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "dhook/1.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		c.rateLimiter.HandleResponse(editURL, resp.StatusCode, resp.Header)

		if resp.StatusCode == 429 {
			c.fireRateLimit(c.rateLimiter.getLastRetryAfter(editURL))
			continue
		}

		if resp.StatusCode >= 500 && attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("dhook: edit returned status %d: %s", resp.StatusCode, string(respBody))
		}

		var r Response
		if err := json.Unmarshal(respBody, &r); err != nil {
			return nil, err
		}
		return &r, nil
	}

	return nil, errors.New("dhook: max retries exceeded on edit")
}

func (c *Client) Delete(ctx context.Context, messageID string) error {
	if len(c.urls) == 0 {
		return errors.New("dhook: no webhook URLs configured")
	}
	return c.deleteMessage(ctx, c.urls[0], messageID)
}

func (c *Client) deleteMessage(ctx context.Context, url, messageID string) error {
	deleteURL := url + "/messages/" + messageID
	maxRetries := 5

	for attempt := 0; attempt <= maxRetries; attempt++ {
		c.rateLimiter.Wait(ctx, deleteURL)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "dhook/1.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		c.rateLimiter.HandleResponse(deleteURL, resp.StatusCode, resp.Header)

		if resp.StatusCode == 429 {
			c.fireRateLimit(c.rateLimiter.getLastRetryAfter(deleteURL))
			continue
		}

		if resp.StatusCode >= 500 && attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			continue
		}

		if resp.StatusCode >= 400 {
			return fmt.Errorf("dhook: delete returned status %d: %s", resp.StatusCode, string(respBody))
		}

		return nil
	}

	return errors.New("dhook: max retries exceeded on delete")
}

func parseRetryAfter(headers http.Header) time.Duration {
	v := headers.Get("Retry-After")
	if v == "" {
		return time.Second
	}

	if secs, err := strconv.Atoi(v); err == nil {
		d := time.Duration(secs) * time.Second
		if d > 0 {
			return d
		}
	}

	for _, layout := range []string{http.TimeFormat, time.RFC1123Z, time.RFC3339} {
		if t, err := time.Parse(layout, v); err == nil {
			d := time.Until(t)
			if d > 0 {
				return d
			}
		}
	}

	return time.Second
}

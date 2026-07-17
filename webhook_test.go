package dhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendSucceedsForNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	responses, err := New(server.URL).Send(context.Background(), &Message{Content: "hello"})
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}
	if len(responses) != 1 || responses[0] == nil {
		t.Fatalf("Send() responses = %#v, want one response", responses)
	}
}

func TestSendDecodesJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(Response{ID: "message-id", Content: "hello"}); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	responses, err := New(server.URL).Send(context.Background(), &Message{Content: "hello"})
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}
	if len(responses) != 1 || responses[0].ID != "message-id" || responses[0].Content != "hello" {
		t.Fatalf("Send() responses = %#v, want decoded response", responses)
	}
}

func TestSendAddsWaitAndPreservesQuery(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.URL.Query().Get("thread_id"); got != "123" {
			t.Errorf("thread_id = %q, want %q", got, "123")
		}
		if got := r.URL.Query().Get("wait"); got != "true" {
			t.Errorf("wait = %q, want %q", got, "true")
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(Response{ID: "message-id", Content: "hello"}); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	// When
	responses, err := New(server.URL+"?thread_id=123").Send(context.Background(), &Message{Content: "hello"})

	// Then
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}
	if len(responses) != 1 || responses[0].ID != "message-id" || responses[0].Content != "hello" {
		t.Fatalf("Send() responses = %#v, want decoded response", responses)
	}
}

func TestSuccessHookCanRegisterHook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(server.URL)
	client.AddHook(EventSuccess, SuccessFunc(func(*Response) {
		client.AddHook(EventSuccess, SuccessFunc(func(*Response) {}))
	}))

	done := make(chan error, 1)
	go func() {
		_, err := client.Send(context.Background(), &Message{Content: "hello"})
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Send() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Send() did not return after a success hook registered another hook")
	}
}

func TestSetHTTPClientNilKeepsExistingClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(server.URL)
	client.SetHTTPClient(nil)
	if _, err := client.Send(context.Background(), &Message{Content: "hello"}); err != nil {
		t.Fatalf("Send() after SetHTTPClient(nil) error = %v, want nil", err)
	}
}

func TestEditAndDeletePreserveWebhookQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/messages/message-id" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/messages/message-id")
		}
		if got := r.URL.Query().Get("thread_id"); got != "123" {
			t.Errorf("thread_id = %q, want %q", got, "123")
		}
		switch r.Method {
		case http.MethodPatch:
			if err := json.NewEncoder(w).Encode(Response{ID: "message-id"}); err != nil {
				t.Fatal(err)
			}
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("method = %s, want PATCH or DELETE", r.Method)
		}
	}))
	defer server.Close()

	client := New(server.URL + "?thread_id=123")
	if _, err := client.Edit(context.Background(), "message-id", &Message{Content: "edited"}); err != nil {
		t.Fatalf("Edit() error = %v, want nil", err)
	}
	if err := client.Delete(context.Background(), "message-id"); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
}

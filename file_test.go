package dhook

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestSendRequiresWebhookURL(t *testing.T) {
	// Given
	client := New()

	// When
	responses, err := client.Send(context.Background(), &Message{Content: "hello"})

	// Then
	if err == nil || !strings.Contains(err.Error(), "no webhook URLs configured") {
		t.Fatalf("Send() error = %v, want no webhook URLs configured", err)
	}
	if len(responses) != 0 {
		t.Fatalf("Send() responses = %#v, want none", responses)
	}
}

func TestSendFilesRequiresWebhookURL(t *testing.T) {
	// Given
	client := New()
	file := NewFile("message.txt", bytes.NewReader([]byte("hello")))

	// When
	responses, err := client.SendFiles(context.Background(), &Message{Content: "hello"}, file)

	// Then
	if err == nil || !strings.Contains(err.Error(), "no webhook URLs configured") {
		t.Fatalf("SendFiles() error = %v, want no webhook URLs configured", err)
	}
	if len(responses) != 0 {
		t.Fatalf("SendFiles() responses = %#v, want none", responses)
	}
}

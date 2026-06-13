package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RivnZero/dhook"
)

func main() {
	client := dhook.New(
		"https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN",
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	msg := &dhook.Message{
		Content: "Hello from dhook!",
	}

	responses, err := client.Send(ctx, msg)
	if err != nil {
		log.Fatal(err)
	}

	for _, resp := range responses {
		fmt.Printf("Message sent! ID: %s\n", resp.ID)
	}
}

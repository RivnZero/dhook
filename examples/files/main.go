package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RivnZero/dhook"
)

func main() {
	client := dhook.New(
		"https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN",
	)

	file, err := os.Open("screenshot.png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	msg := &dhook.Message{
		Content: "Check out this screenshot!",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	responses, err := client.SendFile(ctx, "screenshot.png", file, msg)
	if err != nil {
		log.Fatal(err)
	}

	for _, resp := range responses {
		fmt.Printf("File sent! Message ID: %s\n", resp.ID)
	}
}

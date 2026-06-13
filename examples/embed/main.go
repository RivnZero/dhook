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

	embed := dhook.NewEmbed().
		SetTitle("Project Update").
		SetDescription("Here is the latest status report for the project.").
		SetColor(0x5865F2).
		AddField("Status", "In Progress", true).
		AddField("Priority", "High", true).
		AddField("Assigned To", "Team Alpha", false).
		SetFooter("dhook SDK", "").
		SetAuthor("CI Bot", "", "").
		SetTimestamp(time.Now())

	msg := &dhook.Message{
		Content: "**New Project Update**",
		Embeds:  []*dhook.Embed{embed},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	responses, err := client.Send(ctx, msg)
	if err != nil {
		log.Fatal(err)
	}

	for _, resp := range responses {
		fmt.Printf("Embed sent! Message ID: %s\n", resp.ID)
	}
}

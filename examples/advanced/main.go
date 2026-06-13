package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/RivnZero/dhook"
)

func main() {
	webhookURLs := flag.String("urls", "", "comma-separated webhook URLs (required)")
	content := flag.String("content", "", "message content")
	username := flag.String("username", "", "override display username")
	avatarURL := flag.String("avatar", "", "override avatar URL")
	filePath := flag.String("file", "", "file path to attach")
	fileName := flag.String("filename", "", "attachment filename override")
	embedTitle := flag.String("embed-title", "", "embed title")
	embedDesc := flag.String("embed-desc", "", "embed description")
	embedColor := flag.Int("embed-color", 0, "embed color as hex int (e.g. 5865F2)")
	queueMode := flag.Bool("queue", false, "use background queue workers")
	workerCount := flag.Int("workers", 5, "number of queue workers")
	timeout := flag.Duration("timeout", 30*time.Second, "request timeout")

	flag.Parse()

	if *webhookURLs == "" {
		fmt.Fprintln(os.Stderr, "error: --urls is required")
		flag.Usage()
		os.Exit(1)
	}

	if *content == "" && *embedTitle == "" && *filePath == "" {
		fmt.Fprintln(os.Stderr, "error: provide --content, --embed-title, or --file")
		flag.Usage()
		os.Exit(1)
	}

	var urls []string
	for _, u := range strings.Split(*webhookURLs, ",") {
		u = strings.TrimSpace(u)
		if u != "" {
			urls = append(urls, u)
		}
	}

	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "error: no valid webhook URLs provided")
		os.Exit(1)
	}

	client := dhook.New(urls...)

	client.AddHook(dhook.EventSuccess, func(resp *dhook.Response) {
		fmt.Printf("[SUCCESS] delivered: %s\n", resp.ID)
	})

	client.AddHook(dhook.EventRateLimit, func(retryAfter time.Duration) {
		fmt.Printf("[RATE LIMITED] retrying after %v\n", retryAfter)
	})

	client.AddHook(dhook.EventError, func(err error) {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, *timeout)
	defer cancel()

	msg := &dhook.Message{
		Content:   *content,
		Username:  *username,
		AvatarURL: *avatarURL,
	}

	if *embedTitle != "" || *embedDesc != "" {
		embed := dhook.NewEmbed().SetTimestamp(time.Now())
		if *embedTitle != "" {
			embed.SetTitle(*embedTitle)
		}
		if *embedDesc != "" {
			embed.SetDescription(*embedDesc)
		}
		if *embedColor != 0 {
			embed.SetColor(*embedColor)
		}
		msg.Embeds = []*dhook.Embed{embed}
	}

	if *queueMode {
		runQueue(ctx, client, msg, *filePath, *fileName, *workerCount)
		return
	}

	if *filePath != "" {
		name := *filePath
		if *fileName != "" {
			name = *fileName
		} else {
			parts := strings.Split(*filePath, "/")
			name = parts[len(parts)-1]
		}
		file, err := os.Open(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		responses, err := client.SendFile(ctx, name, file, msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		for _, r := range responses {
			fmt.Printf("[SENT] %s\n", r.ID)
		}
		return
	}

	responses, err := client.Send(ctx, msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	for _, r := range responses {
		fmt.Printf("[SENT] %s\n", r.ID)
	}
}

func runQueue(ctx context.Context, client *dhook.Client, msg *dhook.Message, filePath, fileName string, workers int) {
	queue := dhook.NewQueue(client, workers)
	queue.Start(ctx)

	if filePath != "" {
		queue.AddFunc(func() {
			name := filePath
			if fileName != "" {
				name = fileName
			} else {
				parts := strings.Split(filePath, "/")
				name = parts[len(parts)-1]
			}
			file, err := os.Open(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] open file: %v\n", err)
				return
			}
			defer file.Close()
			if _, err := client.SendFile(ctx, name, file, msg); err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			}
		})
	} else {
		queue.Add(msg)
	}

	queue.Stop()
	fmt.Println("[QUEUE] all jobs completed")
}

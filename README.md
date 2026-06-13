<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT%20with%20Attribution-blue?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/github/actions/workflow/status/RivnZero/dhook/ci.yml?style=for-the-badge&label=CI" alt="CI Status">
  <img src="https://img.shields.io/badge/Zero%20Dependencies-lightgrey?style=for-the-badge" alt="Zero Dependencies">
</p>

<h1 align="center">dhook</h1>

<p align="center">
  A production-ready, enterprise-grade Discord Webhook SDK for Go.<br>
  Multi-URL routing, automatic rate limiting, resilient queue, and fluent embed builder.
</p>

---

## Features

- **Multi-URL Routing** — Broadcast messages to multiple webhooks simultaneously
- **Automatic Rate Limiting** — Respects Discord's rate limits with seamless retry
- **Smart Retry Backoff** — Exponential backoff for 5xx server errors
- **Fluent Embed Builder** — Chainable API for constructing rich embeds
- **File Uploads** — Send files and attachments via multipart form data
- **Background Queue** — Concurrent worker pool for high-throughput delivery
- **Event Hooks** — Register callbacks for success, rate limit, and error events
- **Context Support** — All methods accept `context.Context` for timeout and cancellation
- **Zero Dependencies** — Built entirely on the Go standard library

## Installation

```bash
go get github.com/RivnZero/dhook
```

## Quick Start

### Send a Simple Message

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/RivnZero/dhook"
)

func main() {
    client := dhook.New(
        "https://discord.com/api/webhooks/ID/TOKEN",
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

    log.Printf("Sent to %d webhook(s)", len(responses))
}
```

### Send a Rich Embed

```go
embed := dhook.NewEmbed().
    SetTitle("Build Complete").
    SetDescription("All tests passed.").
    SetColor(0x57F287).
    AddField("Status", "Success", true).
    AddField("Duration", "12s", true).
    SetFooter("CI Pipeline", "").
    SetTimestamp(time.Now())

msg := &dhook.Message{
    Content: "**Build Report**",
    Embeds:  []*dhook.Embed{embed},
}

client.Send(ctx, msg)
```

### Send Files

```go
file, _ := os.Open("report.pdf")
defer file.Close()

msg := &dhook.Message{
    Content: "Here is the report.",
}

client.SendFile(ctx, "report.pdf", file, msg)
```

### Multi-URL Routing

```go
client := dhook.New(
    "https://discord.com/api/webhooks/ID_1/TOKEN_1",
    "https://discord.com/api/webhooks/ID_2/TOKEN_2",
    "https://discord.com/api/webhooks/ID_3/TOKEN_3",
)

msg := &dhook.Message{Content: "Broadcasting to all channels!"}
responses, err := client.Send(ctx, msg)
```

### Event Hooks

```go
client.AddHook(dhook.EventSuccess, func(resp *dhook.Response) {
    log.Printf("Message delivered: %s", resp.ID)
})

client.AddHook(dhook.EventRateLimit, func(retryAfter time.Duration) {
    log.Printf("Rate limited, retrying after %v", retryAfter)
})

client.AddHook(dhook.EventError, func(err error) {
    log.Printf("Error: %v", err)
})
```

### Background Queue

```go
queue := dhook.NewQueue(client, 10)
queue.Start(ctx)
defer queue.Stop()

for i := 0; i < 1000; i++ {
    queue.Add(&dhook.Message{
        Content: fmt.Sprintf("Message #%d", i),
    })
}
```

## API Reference

### Client

| Method | Description |
|--------|-------------|
| `New(urls ...string) *Client` | Create a new client with one or more webhook URLs |
| `SetHTTPClient(client *http.Client)` | Override the default HTTP client |
| `Send(ctx, msg) ([]*Response, error)` | Send a message to all configured webhooks |
| `SendFile(ctx, name, reader, msg) ([]*Response, error)` | Send a file with an optional message |
| `SendFiles(ctx, msg, files...) ([]*Response, error)` | Send multiple files with a message |
| `Edit(ctx, messageID, msg) (*Response, error)` | Edit a previously sent message |
| `Delete(ctx, messageID) error` | Delete a message |
| `AddHook(event, callback)` | Register an event callback |

### Embed Builder

| Method | Description |
|--------|-------------|
| `NewEmbed() *Embed` | Create a new embed |
| `SetTitle(string)` | Set the embed title |
| `SetDescription(string)` | Set the embed description |
| `SetColor(int)` | Set the embed color (hex integer) |
| `SetURL(string)` | Set the embed URL |
| `SetTimestamp(time.Time)` | Set the embed timestamp |
| `SetFooter(text, iconURL)` | Set the embed footer |
| `SetImage(url)` | Set the embed image |
| `SetThumbnail(url)` | Set the embed thumbnail |
| `SetAuthor(name, url, iconURL)` | Set the embed author |
| `AddField(name, value, inline)` | Add a field to the embed |

### Queue

| Method | Description |
|--------|-------------|
| `NewQueue(client, workerCount)` | Create a new queue with N workers |
| `Start(ctx)` | Start processing jobs |
| `Stop()` | Stop processing and drain remaining jobs |
| `Add(msg)` | Queue a message for sending |
| `AddFunc(fn)` | Queue a custom function for execution |
| `Len()` | Number of pending jobs |
| `Cap()` | Maximum queue capacity |

## CLI Tool

dhook ships with a standalone CLI binary built from the advanced example. Use it to send messages directly from the terminal:

```bash
dhook \
  --urls "https://discord.com/api/webhooks/ID/TOKEN" \
  --content "Hello from the CLI!" \
  --embed-title "Deploy Complete" \
  --embed-desc "All services are live." \
  --embed-color 0x57F287
```

### CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--urls` | Comma-separated webhook URLs | (required) |
| `--content` | Message content | |
| `--username` | Override display username | |
| `--avatar` | Override avatar URL | |
| `--file` | File path to attach | |
| `--filename` | Attachment filename override | (uses basename) |
| `--embed-title` | Embed title | |
| `--embed-desc` | Embed description | |
| `--embed-color` | Embed color as hex int | `0` |
| `--queue` | Use background queue workers | `false` |
| `--workers` | Number of queue workers | `5` |
| `--timeout` | Request timeout | `30s` |

## Deployment & Release Pipeline

### Local Pre-Release (deploy.bat)

The `deploy.bat` script is the local gatekeeper before any release. Run it to verify everything is production-ready:

```bat
deploy.bat
```

It performs the following steps:

1. **`go vet ./...`** — Static analysis for common issues
2. **`go test -v ./...`** — Runs all tests (stops on first failure)
3. **Cross-compilation dry-run** — Builds the CLI binary for all 9 targets:
   - `windows/386`, `windows/amd64`, `windows/arm64`
   - `linux/386`, `linux/amd64`, `linux/arm64`
   - `darwin/amd64`, `darwin/arm64`
4. **Verdict** — If all builds succeed, prints the safe-to-tag instructions:

```
git tag vX.Y.Z
git push origin vX.Y.Z
```

### Cloud Release (GoReleaser + GitHub Actions)

Once you push a semantic version tag, the `release.yml` workflow automatically:

1. Triggers on any `v*` tag push
2. Sets up the latest Go version
3. Runs GoReleaser which:
   - Cross-compiles for all OS/arch targets (excluding `darwin/386`)
   - Packages `.zip` archives for Windows and `.tar.gz` for Linux/macOS
   - Generates a changelog grouped by features and fixes
   - Creates a GitHub Release with all artifacts and checksums

### Full Release Workflow

```bash
# 1. Run local gatekeeper
deploy.bat

# 2. Tag and push
git tag v1.0.0
git push origin v1.0.0

# 3. GitHub Actions handles the rest automatically
```

## How It Works

### Rate Limiting

dhook reads Discord's `X-Rate-Limit-Remaining` and `Retry-After` headers on every response. When limits are hit, the rate limiter automatically blocks subsequent requests until the cooldown expires. No dropped messages, no manual retry logic.

### Exponential Backoff

On 5xx server errors, dhook retries with exponential backoff starting at 1 second, doubling up to 30 seconds, with a maximum of 5 retries per request.

### Concurrent Broadcasting

When multiple webhook URLs are configured, `Send` dispatches requests to all URLs concurrently using goroutines. Each URL is rate-limited independently.

## License

MIT License with Attribution Requirement. See [LICENSE](LICENSE) for details.

**Author: Mohammad Kian Ostadamahmadi**

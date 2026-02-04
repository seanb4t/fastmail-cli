# Go Library API

The FastMail CLI includes a Go library (`pkg/fastmail`) for programmatic access to FastMail services.

## Installation

```bash
go get github.com/seanb4t/fastmail-cli/pkg/fastmail
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func main() {
    ctx := context.Background()

    // Create client with JMAP endpoint and access token
    client := fastmail.NewClient(
        "https://api.fastmail.com/jmap/session",
        "your-api-token",
    )

    // Establish session
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // List emails from inbox
    emails, err := client.Mail().List(ctx, "Inbox", 10)
    if err != nil {
        log.Fatal(err)
    }

    for _, email := range emails {
        fmt.Printf("%s: %s\n", email.ReceivedAt.Format("Jan 2"), email.Subject)
    }
}
```

## Package Structure

| Type | Description |
|------|-------------|
| [`Client`](fastmail.md) | Main client for FastMail operations |
| [`MailService`](mail.md) | Email operations (list, search, send, reply) |
| [`ContactsClient`](contacts.md) | Contact management via CardDAV |
| [`CalendarService`](calendar.md) | Calendar and event operations via CalDAV |
| [`MaskedEmailService`](masked-email.md) | Masked email address management |

## Authentication

The library requires a FastMail API token. Generate one from:

1. Log into FastMail web interface
2. Go to **Settings** → **Privacy & Security** → **API tokens**
3. Create a new token with appropriate scopes

## Error Handling

All methods return errors wrapped with context using [oops](https://github.com/samber/oops). Errors include stack traces and contextual information for debugging.

```go
emails, err := client.Mail().List(ctx, "Inbox", 10)
if err != nil {
    // Error includes context: "listing emails: executing JMAP request: ..."
    log.Printf("Failed: %v", err)
    return
}
```

## Context Support

All operations accept a `context.Context` for:

- Cancellation
- Timeouts
- Request tracing

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

emails, err := client.Mail().List(ctx, "Inbox", 10)
```

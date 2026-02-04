# Client

The `Client` type is the main entry point for FastMail operations.

## Creating a Client

```go
import "github.com/seanb4t/fastmail-cli/pkg/fastmail"

client := fastmail.NewClient(endpoint, accessToken)
```

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `endpoint` | `string` | JMAP session URL (e.g., `https://api.fastmail.com/jmap/session`) |
| `accessToken` | `string` | FastMail API token |

## Methods

### Connect

Establishes a session and retrieves the account ID. Must be called before using other operations.

```go
func (c *Client) Connect(ctx context.Context) error
```

#### Example

```go
ctx := context.Background()
client := fastmail.NewClient(endpoint, token)

if err := client.Connect(ctx); err != nil {
    log.Fatalf("Connection failed: %v", err)
}
```

#### Errors

| Condition | Error |
|-----------|-------|
| Invalid token | `authenticating: ...` |
| No mail account | `no mail account found` |

### Mail

Returns the mail service for email operations.

```go
func (c *Client) Mail() *MailService
```

#### Example

```go
mailService := client.Mail()
emails, err := mailService.List(ctx, "Inbox", 10)
```

See [MailService](mail.md) for available operations.

### MaskedEmail

Returns the masked email service for managing masked email addresses.

```go
func (c *Client) MaskedEmail() *MaskedEmailService
```

#### Example

```go
maskedService := client.MaskedEmail()
addresses, err := maskedService.List(ctx)
```

See [MaskedEmailService](masked-email.md) for available operations.

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func main() {
    ctx := context.Background()

    client := fastmail.NewClient(
        "https://api.fastmail.com/jmap/session",
        os.Getenv("FASTMAIL_TOKEN"),
    )

    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // Use mail service
    emails, _ := client.Mail().List(ctx, "Inbox", 5)
    for _, e := range emails {
        fmt.Println(e.Subject)
    }

    // Use masked email service
    masked, _ := client.MaskedEmail().List(ctx)
    for _, m := range masked {
        fmt.Println(m.Email)
    }
}
```

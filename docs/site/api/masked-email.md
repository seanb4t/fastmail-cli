# MaskedEmailService

Manage FastMail masked email addresses.

Masked emails are unique, disposable addresses that forward to your real inbox. They help protect your primary email from spam and tracking.

## Access

```go
maskedService := client.MaskedEmail()
```

## Methods

### List

Returns all masked email addresses.

```go
func (s *MaskedEmailService) List(ctx context.Context) ([]MaskedEmail, error)
```

#### Example

```go
addresses, err := client.MaskedEmail().List(ctx)
for _, m := range addresses {
    fmt.Printf("%s (%s) - %s\n", m.Email, m.State, m.ForDomain)
}
```

### Create

Creates a new masked email address.

```go
func (s *MaskedEmailService) Create(ctx context.Context, opts CreateMaskedEmailOptions) (*MaskedEmail, error)
```

#### CreateMaskedEmailOptions

| Field | Type | Description |
|-------|------|-------------|
| `ForDomain` | `string` | Associated domain (e.g., `"example.com"`) |
| `Description` | `string` | Description for your reference |

#### Example

```go
masked, err := client.MaskedEmail().Create(ctx, fastmail.CreateMaskedEmailOptions{
    ForDomain:   "shopping-site.com",
    Description: "Shopping account",
})

fmt.Printf("Created: %s\n", masked.Email)
// Output: Created: abc123xyz@fastmail.com
```

### Enable

Enables a disabled masked email address.

```go
func (s *MaskedEmailService) Enable(ctx context.Context, id string) error
```

#### Example

```go
err := client.MaskedEmail().Enable(ctx, maskedID)
```

### Disable

Disables a masked email address. Emails sent to disabled addresses are rejected.

```go
func (s *MaskedEmailService) Disable(ctx context.Context, id string) error
```

#### Example

```go
err := client.MaskedEmail().Disable(ctx, maskedID)
```

### Delete

Permanently deletes a masked email address.

```go
func (s *MaskedEmailService) Delete(ctx context.Context, id string) error
```

#### Example

```go
err := client.MaskedEmail().Delete(ctx, maskedID)
```

## Types

### MaskedEmail

Represents a masked email address.

```go
type MaskedEmail struct {
    ID            string
    Email         string
    State         MaskedEmailState
    ForDomain     string
    Description   string
    URL           string
    CreatedBy     string
    CreatedAt     string
    LastMessageAt string
}
```

| Field | Description |
|-------|-------------|
| `ID` | Unique identifier |
| `Email` | The masked email address |
| `State` | Current state (see below) |
| `ForDomain` | Associated domain |
| `Description` | User description |
| `URL` | Associated URL |
| `CreatedBy` | Creator identifier |
| `CreatedAt` | Creation timestamp |
| `LastMessageAt` | Last received message timestamp |

### MaskedEmailState

State constants for masked email lifecycle:

| Constant | Value | Description |
|----------|-------|-------------|
| `MaskedEmailStateEnabled` | `enabled` | Active, receiving email |
| `MaskedEmailStateDisabled` | `disabled` | Inactive, rejecting email |
| `MaskedEmailStatePending` | `pending` | Awaiting first use |
| `MaskedEmailStateDeleted` | `deleted` | Permanently removed |

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

    maskedService := client.MaskedEmail()

    // List existing masked emails
    addresses, err := maskedService.List(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d masked emails\n", len(addresses))
    for _, m := range addresses {
        status := "✓"
        if m.State == fastmail.MaskedEmailStateDisabled {
            status = "✗"
        }
        fmt.Printf("  %s %s (%s)\n", status, m.Email, m.ForDomain)
    }

    // Create a new masked email
    newMasked, err := maskedService.Create(ctx, fastmail.CreateMaskedEmailOptions{
        ForDomain:   "newsletter.example.com",
        Description: "Newsletter subscription",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\nCreated: %s\n", newMasked.Email)

    // Later, disable it if spam starts arriving
    // err = maskedService.Disable(ctx, newMasked.ID)
}
```

## Use Cases

### Sign up for services

Create a unique masked email for each service to track who shares your address:

```go
masked, _ := client.MaskedEmail().Create(ctx, fastmail.CreateMaskedEmailOptions{
    ForDomain:   "shady-service.com",
    Description: "Free trial signup",
})
```

### Manage spam

When a masked email starts receiving spam, disable it:

```go
// Find the problematic address
addresses, _ := client.MaskedEmail().List(ctx)
for _, m := range addresses {
    if m.ForDomain == "spammy-site.com" {
        client.MaskedEmail().Disable(ctx, m.ID)
        break
    }
}
```

### Clean up unused addresses

Delete masked emails you no longer need:

```go
addresses, _ := client.MaskedEmail().List(ctx)
for _, m := range addresses {
    if m.State == fastmail.MaskedEmailStateDisabled && m.LastMessageAt == "" {
        client.MaskedEmail().Delete(ctx, m.ID)
    }
}
```

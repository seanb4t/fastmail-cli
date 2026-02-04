# Contacts

Contact management via CardDAV.

## ContactsClient

Standalone client for contact operations.

### Creating a Client

```go
import "github.com/seanb4t/fastmail-cli/pkg/fastmail"

client := fastmail.NewContactsClient(endpoint, username, password)
```

#### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `endpoint` | `string` | CardDAV URL (e.g., `https://carddav.fastmail.com`) |
| `username` | `string` | FastMail username |
| `password` | `string` | App-specific password |

## Methods

### List

Returns all contacts from the default address book.

```go
func (c *ContactsClient) List(ctx context.Context) ([]Contact, error)
```

#### Example

```go
contacts, err := client.List(ctx)
for _, c := range contacts {
    fmt.Printf("%s <%s>\n", c.Name, c.Email)
}
```

### Get

Returns a single contact by ID.

```go
func (c *ContactsClient) Get(ctx context.Context, id string) (*Contact, error)
```

#### Example

```go
contact, err := client.Get(ctx, "abc123")
if err != nil {
    log.Fatal(err)
}
fmt.Println(contact.Name)
```

### Search

Returns contacts matching a query string. Searches across name and email fields.

```go
func (c *ContactsClient) Search(ctx context.Context, query string) ([]Contact, error)
```

#### Example

```go
// Search by name
matches, err := client.Search(ctx, "Smith")

// Search by email domain
matches, err := client.Search(ctx, "@company.com")
```

### Create

Adds a new contact to the address book.

```go
func (c *ContactsClient) Create(ctx context.Context, contact *Contact) error
```

#### Example

```go
contact := &fastmail.Contact{
    Name:  "Jane Doe",
    Email: "jane@example.com",
    Phone: "+1-555-0123",
}

err := client.Create(ctx, contact)
// contact.ID is populated after creation
```

### Update

Modifies an existing contact.

```go
func (c *ContactsClient) Update(ctx context.Context, contact *Contact) error
```

#### Example

```go
contact, _ := client.Get(ctx, contactID)
contact.Phone = "+1-555-9999"
err := client.Update(ctx, contact)
```

### Delete

Removes a contact by ID.

```go
func (c *ContactsClient) Delete(ctx context.Context, id string) error
```

#### Example

```go
err := client.Delete(ctx, contactID)
```

## Types

### Contact

Represents a contact.

```go
type Contact struct {
    ID      string
    Name    string
    Email   string
    Phone   string
    Address string
}
```

| Field | Description |
|-------|-------------|
| `ID` | Unique identifier (UUID) |
| `Name` | Full display name |
| `Email` | Primary email address |
| `Phone` | Primary phone number |
| `Address` | Formatted postal address |

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

    client := fastmail.NewContactsClient(
        "https://carddav.fastmail.com",
        os.Getenv("FASTMAIL_USER"),
        os.Getenv("FASTMAIL_PASSWORD"),
    )

    // List all contacts
    contacts, err := client.List(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d contacts\n", len(contacts))

    // Create a new contact
    newContact := &fastmail.Contact{
        Name:  "New Person",
        Email: "new@example.com",
    }

    if err := client.Create(ctx, newContact); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created contact: %s\n", newContact.ID)

    // Search contacts
    matches, _ := client.Search(ctx, "Person")
    for _, c := range matches {
        fmt.Printf("Match: %s\n", c.Name)
    }
}
```

## Authentication

CardDAV requires an app-specific password, not your main FastMail password:

1. Log into FastMail web interface
2. Go to **Settings** → **Privacy & Security** → **App passwords**
3. Create a new app password for CardDAV access

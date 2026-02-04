# pkg/fastmail

High-level Fastmail client library for Go.

## Purpose

Provide a clean Go API for Fastmail operations:
- Email (list, read, send, archive, flag)
- Contacts (list, search, get)
- Calendar (list events, get event)
- Masked email (list, create, disable)
- Data export

## Key Types

| Type | File | Description |
|------|------|-------------|
| `Client` | client.go | High-level Fastmail client |
| `MailService` | mail.go | Email operations |
| `ContactsService` | contacts.go | Contact operations |
| `MaskedEmailService` | masked_email.go | Masked email operations |
| `Email` | email.go | Email message type |
| `Mailbox` | mailbox.go | Mailbox/folder type |
| `Contact` | contact.go | Contact type |
| `Event` | event.go | Calendar event type |

## Usage

```go
client := fastmail.NewClient(endpoint, accessToken)
if err := client.Connect(ctx); err != nil {
    return err
}

// List emails
emails, err := client.Mail().List(ctx, fastmail.ListOptions{
    Mailbox: "inbox",
    Limit:   10,
})

// List contacts
contacts, err := client.Contacts().List(ctx)
```

## Conventions

- Hide JMAP/DAV complexity from consumers
- Return domain types, not protocol types
- Context for cancellation and timeout
- Service pattern: `client.ServiceName().Method()`

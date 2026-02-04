# MailService

The `MailService` provides email operations via JMAP.

## Access

```go
mailService := client.Mail()
```

## Methods

### List

Returns emails from a folder.

```go
func (s *MailService) List(ctx context.Context, folder string, limit uint64) ([]Email, error)
```

#### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `folder` | `string` | Folder name (`"Inbox"`, `"Sent"`) or mailbox ID |
| `limit` | `uint64` | Maximum emails to return; `0` for server default |

#### Example

```go
// List latest 10 inbox emails
emails, err := client.Mail().List(ctx, "Inbox", 10)

// List from Sent folder
sent, err := client.Mail().List(ctx, "Sent", 20)
```

### Get

Returns a single email by ID.

```go
func (s *MailService) Get(ctx context.Context, id string) (*Email, error)
```

#### Example

```go
email, err := client.Mail().Get(ctx, "M12345")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Subject: %s\n", email.Subject)
```

### Search

Returns emails matching a query string.

```go
func (s *MailService) Search(ctx context.Context, query string, limit uint64) ([]Email, error)
```

#### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `query` | `string` | JMAP filter text (e.g., `"from:alice subject:meeting"`) |
| `limit` | `uint64` | Maximum results; `0` for server default |

#### Example

```go
// Search by sender
results, err := client.Mail().Search(ctx, "from:alice@example.com", 50)

// Search by subject
results, err := client.Mail().Search(ctx, "subject:quarterly report", 10)
```

### Move

Moves an email to a different folder.

```go
func (s *MailService) Move(ctx context.Context, id string, folder string) error
```

#### Example

```go
// Move to Archive
err := client.Mail().Move(ctx, emailID, "Archive")

// Move to custom folder
err := client.Mail().Move(ctx, emailID, "Projects")
```

### Delete

Moves an email to Trash, or permanently destroys it if already in Trash.

```go
func (s *MailService) Delete(ctx context.Context, id string) error
```

#### Example

```go
// First delete moves to Trash
err := client.Mail().Delete(ctx, emailID)

// Calling again on same email permanently deletes it
```

### Send

Creates and sends a new email.

```go
func (s *MailService) Send(ctx context.Context, opts SendOptions) (string, error)
```

#### SendOptions

| Field | Type | Description |
|-------|------|-------------|
| `To` | `[]EmailAddress` | Primary recipients |
| `Cc` | `[]EmailAddress` | Carbon copy recipients |
| `Bcc` | `[]EmailAddress` | Blind carbon copy recipients |
| `Subject` | `string` | Email subject line |
| `Body` | `string` | Plain text body |

#### Example

```go
emailID, err := client.Mail().Send(ctx, fastmail.SendOptions{
    To: []fastmail.EmailAddress{
        {Name: "Alice", Email: "alice@example.com"},
    },
    Subject: "Meeting Tomorrow",
    Body:    "Hi Alice,\n\nCan we meet at 2pm?\n\nBest",
})
```

### Reply

Creates and sends a reply to an existing email.

```go
func (s *MailService) Reply(ctx context.Context, opts ReplyOptions) (string, error)
```

#### ReplyOptions

| Field | Type | Description |
|-------|------|-------------|
| `EmailID` | `string` | ID of email to reply to |
| `Body` | `string` | Reply body text |
| `ReplyAll` | `bool` | Include all original recipients |

#### Example

```go
// Reply to sender only
replyID, err := client.Mail().Reply(ctx, fastmail.ReplyOptions{
    EmailID: originalID,
    Body:    "Thanks for the update!",
})

// Reply all
replyID, err := client.Mail().Reply(ctx, fastmail.ReplyOptions{
    EmailID:  originalID,
    Body:     "Adding my thoughts...",
    ReplyAll: true,
})
```

## Types

### Email

Represents an email message.

```go
type Email struct {
    ID         string
    ThreadID   string
    Subject    string
    From       EmailAddress
    To         []EmailAddress
    Cc         []EmailAddress
    Bcc        []EmailAddress
    ReceivedAt time.Time
    Preview    string
    Body       string
    Keywords   []string
    MailboxIDs []string
    Size       uint64
}
```

#### Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `HasKeyword(keyword)` | `bool` | Check for specific keyword |
| `IsRead()` | `bool` | Check if `$seen` keyword present |
| `IsFlagged()` | `bool` | Check if `$flagged` keyword present |
| `IsDraft()` | `bool` | Check if `$draft` keyword present |

### EmailAddress

Represents an email address with optional display name.

```go
type EmailAddress struct {
    Name  string  // Display name (e.g., "John Doe")
    Email string  // Email address (e.g., "john@example.com")
}
```

#### Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `String()` | `string` | Formatted as `"Name <email>"` or just email |

### Email Keywords

Common JMAP email keywords:

| Constant | Value | Description |
|----------|-------|-------------|
| `KeywordSeen` | `$seen` | Email has been read |
| `KeywordFlagged` | `$flagged` | Email is starred/flagged |
| `KeywordDraft` | `$draft` | Email is a draft |
| `KeywordAnswered` | `$answered` | Email has been replied to |
| `KeywordForwarded` | `$forwarded` | Email has been forwarded |

### Mailbox

Represents an email folder.

```go
type Mailbox struct {
    ID            string
    Name          string
    Role          MailboxRole
    ParentID      string
    TotalEmails   uint64
    UnreadEmails  uint64
    TotalThreads  uint64
    UnreadThreads uint64
}
```

### MailboxRole

Standard mailbox roles:

| Constant | Description |
|----------|-------------|
| `RoleInbox` | Primary inbox |
| `RoleDrafts` | Draft messages |
| `RoleSent` | Sent messages |
| `RoleTrash` | Deleted messages |
| `RoleJunk` | Spam/junk messages |
| `RoleArchive` | Archived messages |
| `RoleAll` | All emails (virtual) |
| `RoleImportant` | Important messages |
| `RoleFlagged` | Flagged messages (virtual) |

# MCP Tools Reference

Tools allow Claude to perform actions on your FastMail account.

## Mail Tools

### mail_list

List emails from a mailbox folder.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `folder` | string | No | Mailbox folder name (default: "Inbox") |
| `limit` | integer | No | Maximum emails to return (default: 10) |

**Example:**

```
List the last 5 emails from my Sent folder
```

**Response:**

```json
[
  {
    "id": "M123",
    "thread_id": "T456",
    "subject": "Meeting notes",
    "preview": "Here are the notes from...",
    "received_at": "2024-01-15T10:30:00Z",
    "is_read": true,
    "is_flagged": false
  }
]
```

---

### mail_get

Get a single email by ID.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The email ID |

**Example:**

```
Show me the full content of email M123
```

---

### mail_search

Search emails by text query.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `query` | string | Yes | Search query (e.g., "from:alice subject:meeting") |
| `limit` | integer | No | Maximum results (default: 10) |

**Query Syntax:**

- `from:email@example.com` - Sender
- `to:email@example.com` - Recipient
- `subject:keyword` - Subject contains
- `keyword` - Any field contains

**Example:**

```
Search for emails from alice@example.com about the project
```

---

### mail_send

Compose and send a new email.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `to` | array | Yes | Recipient email addresses |
| `cc` | array | No | CC recipient addresses |
| `bcc` | array | No | BCC recipient addresses |
| `subject` | string | Yes | Email subject |
| `body` | string | Yes | Email body text |

**Example:**

```
Send an email to bob@example.com with subject "Quick question" and body "Are you free for a call tomorrow?"
```

**Response:**

```json
{
  "id": "M789",
  "status": "sent"
}
```

---

### mail_reply

Send a reply to an existing email.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `email_id` | string | Yes | ID of the email to reply to |
| `body` | string | Yes | Reply body text |
| `reply_all` | boolean | No | Reply to all recipients (default: false) |

**Example:**

```
Reply to email M123 saying "Thanks, I'll review it today"
```

---

### mail_move

Move an email to a different folder.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The email ID |
| `folder` | string | Yes | Target folder name |

**Example:**

```
Move email M123 to the Archive folder
```

---

### mail_delete

Delete an email (moves to Trash, or permanently deletes if already in Trash).

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The email ID |

**Example:**

```
Delete email M123
```

---

## Masked Email Tools

### masked_email_list

List all masked email addresses.

**Input Schema:** None

**Example:**

```
Show my masked email addresses
```

**Response:**

```json
[
  {
    "id": "ME123",
    "email": "abc123@fastmail.com",
    "state": "enabled",
    "for_domain": "example.com",
    "description": "Newsletter signup"
  }
]
```

---

### masked_email_create

Create a new masked email address.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `domain` | string | No | Domain to associate with the masked email |
| `description` | string | No | Description for the masked email |

**Example:**

```
Create a masked email for signing up to newsletter.example.com
```

---

## Contact Tools

### contacts_list

List all contacts from the address book.

**Input Schema:** None

**Example:**

```
Show my contacts
```

---

### contacts_get

Get a single contact by ID.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The contact ID |

**Example:**

```
Show details for contact C456
```

---

### contacts_create

Create a new contact.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Full name |
| `email` | string | No | Email address |
| `phone` | string | No | Phone number |

**Example:**

```
Add a contact named "Jane Smith" with email jane@example.com
```

---

## Calendar Tools

### calendar_events

List calendar events within a date range.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `calendar_id` | string | No | Calendar ID (uses default if not specified) |
| `start` | string | Yes | Start date/time in RFC3339 format |
| `end` | string | Yes | End date/time in RFC3339 format |

**Example:**

```
Show my calendar events for tomorrow
```

---

### calendar_create

Create a new calendar event.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `calendar_id` | string | Yes | Calendar ID |
| `summary` | string | Yes | Event title/summary |
| `description` | string | No | Event description |
| `location` | string | No | Event location |
| `start` | string | Yes | Start date/time in RFC3339 format |
| `end` | string | Yes | End date/time in RFC3339 format |
| `all_day` | boolean | No | Whether this is an all-day event (default: false) |

**Example:**

```
Create a meeting called "Team standup" tomorrow at 10am for 30 minutes
```

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

### mail_thread

Get all emails in a conversation thread.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `thread_id` | string | Yes | The thread ID |

**Example:**

```
Show me the full thread for thread T456
```

---

### mail_flag

Set or remove flags/keywords on an email.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The email ID |
| `keywords` | object | Yes | Keys are keyword names (e.g., `$seen`, `$flagged`), values are `true`/`false` |

**Example:**

```
Flag email M123 as important and mark it as read
```

**Response:**

```json
{
  "id": "M123",
  "keywords": {
    "$seen": true,
    "$flagged": true
  }
}
```

---

## Mailbox Tools

### mailbox_list

List all mailbox folders with unread and total message counts.

**Input Schema:** None

**Example:**

```
Show my mailbox folders
```

**Response:**

```json
[
  {
    "id": "MB001",
    "name": "Inbox",
    "unread_count": 12,
    "total_count": 458
  }
]
```

---

### mailbox_create

Create a new mailbox folder.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Name for the new mailbox |
| `parent_id` | string | No | Parent mailbox ID for nesting |

**Example:**

```
Create a new mailbox folder called "Receipts"
```

---

### mailbox_rename

Rename an existing mailbox folder.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The mailbox ID |
| `name` | string | Yes | New name for the mailbox |

**Example:**

```
Rename mailbox MB001 to "Old Receipts"
```

---

### mailbox_delete

Delete a mailbox folder by ID.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The mailbox ID |

**Example:**

```
Delete mailbox MB001
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

### masked_email_enable

Enable a previously disabled masked email address.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The masked email ID |

**Example:**

```
Enable masked email ME123
```

---

### masked_email_disable

Disable a masked email address (stops receiving mail).

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The masked email ID |

**Example:**

```
Disable masked email ME123
```

---

### masked_email_delete

Delete a masked email address permanently.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The masked email ID |

**Example:**

```
Delete masked email ME123
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

### contacts_update

Update an existing contact.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The contact ID |
| `name` | string | No | Updated full name |
| `email` | string | No | Updated email address |
| `phone` | string | No | Updated phone number |

**Example:**

```
Update contact C456's email to jane.smith@newdomain.com
```

---

### contacts_delete

Delete a contact by ID.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The contact ID |

**Example:**

```
Delete contact C456
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

---

## Vacation Tools

### vacation_status

Get the current vacation/out-of-office auto-reply status.

**Input Schema:** None

**Example:**

```
Check if my vacation auto-reply is enabled
```

**Response:**

```json
{
  "enabled": true,
  "subject": "Out of Office",
  "body": "I'm on vacation until Feb 14. I'll respond when I return.",
  "from_date": "2026-02-07T00:00:00Z",
  "to_date": "2026-02-14T23:59:59Z"
}
```

---

### vacation_set

Enable or disable the vacation auto-reply.

**Input Schema:**

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `enabled` | boolean | Yes | Whether to enable or disable the auto-reply |
| `subject` | string | Yes (if enabled) | Auto-reply subject line |
| `body` | string | Yes (if enabled) | Auto-reply body text |
| `from_date` | string | No | Start date in RFC3339 format |
| `to_date` | string | No | End date in RFC3339 format |

**Example:**

```
Set my vacation reply from Feb 10 to Feb 14 with subject "Out of Office" and body "I'll be back on the 15th"
```

---

## Account Tools

### quota_get

Get storage quota usage for the account.

**Input Schema:** None

**Example:**

```
How much storage am I using?
```

**Response:**

```json
{
  "used": 2147483648,
  "total": 53687091200,
  "used_human": "2.0 GB",
  "total_human": "50.0 GB",
  "percent_used": 4.0
}

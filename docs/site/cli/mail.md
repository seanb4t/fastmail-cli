# mail

Commands for reading, sending, and managing email.

## Commands

- [mail list](#mail-list) - List emails
- [mail read](#mail-read) - Read an email
- [mail send](#mail-send) - Send an email
- [mail reply](#mail-reply) - Reply to an email

---

## mail list

List emails from a mailbox folder.

```bash
fastmail-cli mail list [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--limit` | `-n` | `10` | Maximum emails to return |
| `--folder` | `-f` | `Inbox` | Mailbox folder name |

### Output

Text output shows email ID, subject, and status:
```
M12345  Meeting tomorrow [read]
M12346  Invoice #1234 [flagged]
M12347  Welcome to FastMail
```

JSON output includes full email metadata.

### Examples

```bash
# List recent inbox emails
fastmail-cli mail list

# List 20 emails from Archive
fastmail-cli mail list --folder Archive --limit 20

# List sent emails as JSON
fastmail-cli mail list --folder Sent --json

# List only 5 emails
fastmail-cli mail list -n 5
```

---

## mail read

Display the full content of an email by its ID.

```bash
fastmail-cli mail read EMAIL_ID
```

### Arguments

| Argument | Description |
|----------|-------------|
| `EMAIL_ID` | ID of the email to read |

### Output

Text output shows headers and body:
```
From:    Alice Smith <alice@example.com>
To:      bob@example.com
Date:    2026-02-04 10:30
Subject: Meeting tomorrow

Hi Bob, can we meet at 2pm tomorrow?
```

JSON output includes the full Email struct.

### Examples

```bash
# Read an email by ID
fastmail-cli mail read M12345

# Read as JSON
fastmail-cli mail read M12345 --json

# List then read the first email
EMAIL_ID=$(fastmail-cli mail list -n 1 --json | jq -r '.[0].id')
fastmail-cli mail read "$EMAIL_ID"
```

---

## mail send

Compose and send a new email.

```bash
fastmail-cli mail send [flags]
```

### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--to` | Yes | Recipient email address (can be repeated) |
| `--cc` | No | CC recipient (can be repeated) |
| `--bcc` | No | BCC recipient (can be repeated) |
| `--subject` | Yes | Email subject |
| `--body` | Yes | Email body text |

### Address Format

Addresses can be plain emails or include names:
- `alice@example.com`
- `Alice Smith <alice@example.com>`

### Output

Text output:
```
Email sent: M12345
```

JSON output:
```json
{
  "id": "M12345",
  "status": "sent"
}
```

### Examples

```bash
# Send a simple email
fastmail-cli mail send \
  --to alice@example.com \
  --subject "Hello" \
  --body "Hi Alice!"

# Send to multiple recipients with CC
fastmail-cli mail send \
  --to alice@example.com \
  --to bob@example.com \
  --cc manager@example.com \
  --subject "Team Update" \
  --body "Here's the weekly update..."

# Send with named recipient
fastmail-cli mail send \
  --to "Alice Smith <alice@example.com>" \
  --subject "Meeting" \
  --body "Can we meet tomorrow?"

# Send and get JSON response
fastmail-cli mail send \
  --to alice@example.com \
  --subject "Test" \
  --body "Testing" \
  --json
```

---

## mail reply

Send a reply to an existing email.

```bash
fastmail-cli mail reply EMAIL_ID [flags]
```

### Arguments

| Argument | Description |
|----------|-------------|
| `EMAIL_ID` | ID of the email to reply to |

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--body` | (required) | Reply body text |
| `--all` | `false` | Reply to all recipients |

### Output

Text output:
```
Email sent: M12346
```

JSON output:
```json
{
  "id": "M12346",
  "status": "sent"
}
```

### Examples

```bash
# Reply to an email
fastmail-cli mail reply M12345 --body "Thanks for the update!"

# Reply all
fastmail-cli mail reply M12345 --body "Adding my thoughts..." --all

# Get reply ID as JSON
fastmail-cli mail reply M12345 --body "Got it." --json
```

## Common Workflows

### Process Inbox

```bash
# List unread emails
fastmail-cli mail list --limit 50 --json | jq '.[] | select(.keywords["$seen"] != true)'

# Quick reply to an email
EMAIL_ID=$(fastmail-cli mail list -n 1 --json | jq -r '.[0].id')
fastmail-cli mail reply "$EMAIL_ID" --body "Thanks, I'll look into this."
```

### Send Notification Script

```bash
#!/bin/bash
fastmail-cli mail send \
  --to alerts@example.com \
  --subject "Build Complete" \
  --body "Build #$BUILD_NUMBER finished successfully."
```

## See Also

- [CLI Reference](index.md)
- [export](export.md) - export emails to files
- [auth](auth.md) - authentication required

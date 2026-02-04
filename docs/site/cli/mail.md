# mail

Read, send, and manage email.

## Synopsis

```bash
fastmail-cli mail <command> [options]
```

## Commands

### list

List emails from a mailbox folder.

```bash
fastmail-cli mail list [--folder <name>] [--limit <n>]
```

**Options:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--folder <name>` | `-f` | `Inbox` | Mailbox folder name |
| `--limit <n>` | `-n` | `10` | Maximum emails to return |

**Examples:**

```bash
# List recent inbox emails
fastmail-cli mail list

# List 20 emails from Sent folder
fastmail-cli mail list --folder Sent --limit 20

# List as JSON
fastmail-cli mail list --json
```

**Output:**

```
M12345  Meeting tomorrow [read]
M12346  Project update [flagged]
M12347  Weekly report
```

Status indicators:

- `[read]` - Email has been read
- `[flagged]` - Email is flagged/starred

---

### send

Send a new email.

```bash
fastmail-cli mail send --to <address> --subject <text> --body <text> [options]
```

**Options:**

| Flag | Required | Description |
|------|----------|-------------|
| `--to <address>` | Yes | Recipient email (can be repeated) |
| `--subject <text>` | Yes | Email subject line |
| `--body <text>` | Yes | Email body text |
| `--cc <address>` | No | CC recipient (can be repeated) |
| `--bcc <address>` | No | BCC recipient (can be repeated) |

**Examples:**

```bash
# Simple email
fastmail-cli mail send \
  --to user@example.com \
  --subject "Hello" \
  --body "This is a test email."

# Multiple recipients
fastmail-cli mail send \
  --to alice@example.com \
  --to bob@example.com \
  --cc manager@example.com \
  --subject "Team Update" \
  --body "Here's the weekly update..."

# With name in address
fastmail-cli mail send \
  --to "Alice Smith <alice@example.com>" \
  --subject "Meeting" \
  --body "Let's meet tomorrow."
```

**Output:**

```
Email sent: M12348
```

---

### reply

Reply to an existing email.

```bash
fastmail-cli mail reply <email-id> --body <text> [--all]
```

**Arguments:**

| Argument | Description |
|----------|-------------|
| `<email-id>` | ID of the email to reply to |

**Options:**

| Flag | Description |
|------|-------------|
| `--body <text>` | Reply body text (required) |
| `--all` | Reply to all recipients |

**Examples:**

```bash
# Reply to sender only
fastmail-cli mail reply M12345 --body "Thanks for the update!"

# Reply to all recipients
fastmail-cli mail reply M12345 --body "I agree with this plan." --all
```

**Output:**

```
Email sent: M12349
```

!!! tip "Finding Email IDs"
    Use `fastmail-cli mail list --json` to get email IDs for replying:
    ```bash
    fastmail-cli mail list --json | jq '.[0].id'
    ```

## Common Folders

FastMail uses these standard folder names:

| Folder | Description |
|--------|-------------|
| `Inbox` | Incoming mail |
| `Drafts` | Draft messages |
| `Sent` | Sent messages |
| `Trash` | Deleted messages |
| `Archive` | Archived messages |
| `Spam` | Spam/junk mail |

Custom folders use their display name.

## JSON Output

With `--json`, list output includes full email metadata:

```json
[
  {
    "id": "M12345",
    "subject": "Meeting tomorrow",
    "from": [{"name": "Alice", "email": "alice@example.com"}],
    "to": [{"name": "You", "email": "you@fastmail.com"}],
    "receivedAt": "2024-01-15T10:30:00Z",
    "keywords": {"$seen": true}
  }
]
```

## See Also

- [export](export.md) - Export emails to backup formats
- [MCP Integration](../reference/mcp.md) - Email tools for AI agents

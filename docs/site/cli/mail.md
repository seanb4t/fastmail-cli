# mail

Commands for reading, sending, searching, and managing email.

## Subcommands

- [mail list](#mail-list) -- List emails
- [mail show](#mail-show) -- Show a single email
- [mail search](#mail-search) -- Search emails
- [mail send](#mail-send) -- Send an email
- [mail reply](#mail-reply) -- Reply to an email
- [mail move](#mail-move) -- Move an email to a folder
- [mail delete](#mail-delete) -- Delete an email
- [mail flag](#mail-flag) -- Set or remove flags on an email
- [mail thread](#mail-thread) -- Show all emails in a thread
- [mail attachments](#mail-attachments) -- List attachments
- [mail download](#mail-download) -- Download an attachment
- [mail upload](#mail-upload) -- Upload a file as a blob
- [mail import](#mail-import) -- Import an RFC 5322 email
- [mail scheduled list](#mail-scheduled-list) -- List pending scheduled sends
- [mail scheduled cancel](#mail-scheduled-cancel) -- Cancel a scheduled send

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

### Examples

```bash
fastmail-cli mail list
fastmail-cli mail list --folder Sent --limit 20
fastmail-cli mail list --folder Archive --json
```

---

## mail show

Display the details of a single email by its ID.

```bash
fastmail-cli mail show EMAIL_ID
```

### Examples

```bash
fastmail-cli mail show M12345
fastmail-cli mail show M12345 --json
```

---

## mail search

Search emails using a text query and/or structured filters.

```bash
fastmail-cli mail search [QUERY] [flags]
```

Query syntax supports field:value filters combined with free text:

```
from:alice subject:"meeting notes" has:attachment before:2024-06-01
is:unread is:flagged in:Inbox
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--limit` | `-n` | `10` | Maximum results |
| `--snippets` | `-s` | `false` | Include highlighted search snippets |
| `--from` | | | Filter by sender |
| `--to` | | | Filter by recipient |
| `--subject` | | | Filter by subject text |
| `--before` | | | Emails before date (YYYY-MM-DD) |
| `--after` | | | Emails after date (YYYY-MM-DD) |
| `--has-attachment` | | | Filter for emails with attachments |
| `--folder` | `-f` | | Filter by mailbox folder name |
| `--unread` | | | Filter for unread emails |
| `--flagged` | | | Filter for flagged/starred emails |

Flags are ANDed with the query. At least one query argument or filter flag is required.

### Examples

```bash
fastmail-cli mail search "quarterly report"
fastmail-cli mail search --from alice@example.com --after 2025-01-01
fastmail-cli mail search --unread --folder Inbox --snippets
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
| `--schedule` | No | Schedule delivery time (RFC3339, e.g. `2024-06-15T14:00:00Z`) |

Addresses can be plain emails or include names: `"Alice Smith <alice@example.com>"`.

### Examples

```bash
fastmail-cli mail send \
  --to alice@example.com \
  --subject "Hello" \
  --body "Hi Alice!"

# Schedule for later delivery
fastmail-cli mail send \
  --to alice@example.com \
  --subject "Reminder" \
  --body "Don't forget!" \
  --schedule 2026-02-10T09:00:00Z
```

---

## mail reply

Send a reply to an existing email.

```bash
fastmail-cli mail reply EMAIL_ID [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--body` | (required) | Reply body text |
| `--all` | `false` | Reply to all recipients |

### Examples

```bash
fastmail-cli mail reply M12345 --body "Thanks for the update!"
fastmail-cli mail reply M12345 --body "Adding my thoughts..." --all
```

---

## mail move

Move an email to a different mailbox folder.

```bash
fastmail-cli mail move EMAIL_ID [flags]
```

### Flags

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--folder` | `-f` | Yes | Destination mailbox folder |

### Examples

```bash
fastmail-cli mail move M12345 --folder Archive
fastmail-cli mail move M12345 -f Projects
```

---

## mail delete

Delete an email by moving it to Trash, or permanently if already in Trash.

```bash
fastmail-cli mail delete EMAIL_ID
```

---

## mail flag

Set or remove keyword flags (read, starred, custom) on an email.

```bash
fastmail-cli mail flag EMAIL_ID [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--read` | Mark as read (`$seen`) |
| `--unread` | Mark as unread (remove `$seen`) |
| `--star` | Star the email (`$flagged`) |
| `--unstar` | Unstar the email (remove `$flagged`) |
| `--flag KEY` | Add a custom keyword (can be repeated) |
| `--unflag KEY` | Remove a custom keyword (can be repeated) |

At least one flag option is required. `--read` and `--unread` are mutually exclusive, as are `--star` and `--unstar`.

### Examples

```bash
fastmail-cli mail flag M12345 --read --star
fastmail-cli mail flag M12345 --unread
fastmail-cli mail flag M12345 --flag "$forwarded"
```

---

## mail thread

Display all emails in a conversation thread, ordered chronologically.

```bash
fastmail-cli mail thread THREAD_ID
```

---

## mail attachments

List attachments on an email message.

```bash
fastmail-cli mail attachments EMAIL_ID
```

Output shows filename, MIME type, size, and blob ID.

---

## mail download

Download an email attachment by name or blob ID.

```bash
fastmail-cli mail download EMAIL_ID [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--attachment` | | Attachment filename to download |
| `--blob-id` | | Blob ID to download |
| `--output` | `-o` | Output file path (default: stdout) |

Either `--attachment` or `--blob-id` is required.

### Examples

```bash
fastmail-cli mail download M12345 --attachment "report.pdf" -o report.pdf
fastmail-cli mail download M12345 --blob-id Bxyz789 -o file.dat
```

---

## mail upload

Upload a file to the server for use in email drafts.

```bash
fastmail-cli mail upload FILE
```

Returns the blob ID and size.

---

## mail import

Import an RFC 5322 email message (.eml file) into a mailbox. Use `-` to read from stdin.

```bash
fastmail-cli mail import FILE [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--folder` | `-f` | `Inbox` | Target mailbox folder |
| `--seen` | | `false` | Mark as read |
| `--flagged` | | `false` | Mark as flagged |

### Examples

```bash
fastmail-cli mail import message.eml --folder Archive --seen
cat message.eml | fastmail-cli mail import -
```

---

## mail scheduled list

List emails that are scheduled for future delivery.

```bash
fastmail-cli mail scheduled list [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--limit` | `-n` | `10` | Maximum results |

---

## mail scheduled cancel

Cancel a pending scheduled email delivery.

```bash
fastmail-cli mail scheduled cancel SUBMISSION_ID
```

## See Also

- [CLI Reference](index.md)
- [export](export.md) -- export emails to files
- [mailbox](mailbox.md) -- manage mailbox folders

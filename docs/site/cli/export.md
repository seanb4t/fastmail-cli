# export

Export emails from a folder to various formats.

## Synopsis

```bash
fastmail-cli export [--folder <name>] [--format <fmt>] [--output <path>]
```

## Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--folder <name>` | `-f` | `Inbox` | Mailbox folder name |
| `--format <fmt>` | | `jsonl` | Export format: `jsonl`, `maildir`, `mbox` |
| `--output <path>` | `-o` | stdout | Output file or directory |

## Formats

### jsonl (JSON Lines)

One JSON object per line, suitable for processing with tools like `jq`.

```bash
# Export to stdout
fastmail-cli export --folder Inbox --format jsonl

# Export to file
fastmail-cli export --folder Inbox --format jsonl --output emails.jsonl
```

**Output format:**

```json
{"id":"M001","subject":"Hello","from":[{"email":"a@example.com"}],...}
{"id":"M002","subject":"Meeting","from":[{"email":"b@example.com"}],...}
```

### maildir

Standard Maildir directory structure, compatible with most mail clients.

```bash
fastmail-cli export --folder Inbox --format maildir --output ~/backup/inbox
```

Creates structure:

```
~/backup/inbox/
├── cur/        # Read messages
├── new/        # Unread messages
└── tmp/        # Temporary files
```

!!! note "Output Required"
    The `--output` flag is required for maildir format.

### mbox

Standard Unix mbox format, a single file with all messages.

```bash
# Export to stdout
fastmail-cli export --folder Archive --format mbox

# Export to file
fastmail-cli export --folder Archive --format mbox --output archive.mbox
```

## Examples

### Backup Inbox

```bash
# Full inbox backup as JSON Lines
fastmail-cli export --folder Inbox --format jsonl --output inbox-backup.jsonl

# As Maildir for import into another client
fastmail-cli export --folder Inbox --format maildir --output ~/mail-backup/inbox
```

### Backup Multiple Folders

```bash
#!/bin/bash
# Backup all standard folders
for folder in Inbox Sent Archive Drafts; do
  fastmail-cli export \
    --folder "$folder" \
    --format mbox \
    --output "backup-$(date +%Y%m%d)-${folder}.mbox"
done
```

### Process with jq

```bash
# Export and filter to emails from specific domain
fastmail-cli export --folder Inbox --format jsonl | \
  jq 'select(.from[0].email | endswith("@company.com"))'

# Count emails by sender
fastmail-cli export --folder Inbox --format jsonl | \
  jq -r '.from[0].email' | sort | uniq -c | sort -rn
```

### Import to Other Clients

```bash
# Export as mbox for import into Thunderbird/Apple Mail
fastmail-cli export --folder Archive --format mbox --output archive.mbox

# Then import the .mbox file using your mail client's import function
```

## Progress Output

Export shows progress on stderr:

```
Fetching emails from Inbox...
Exporting 1523 emails...
Exported to inbox-backup.jsonl
```

Use `--quiet` to suppress progress messages:

```bash
fastmail-cli export --folder Inbox --format jsonl --quiet > backup.jsonl
```

## Performance Notes

- Export fetches all emails in the folder (no limit)
- Large folders may take significant time
- Consider exporting to local file rather than piping for large exports
- Progress is shown on stderr, so piping stdout still works

## See Also

- [mail](mail.md) - List and manage emails
- [Configuration](index.md#configuration) - CLI configuration

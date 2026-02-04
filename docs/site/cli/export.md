# export

Export emails from a mailbox folder to various formats.

## Synopsis

```bash
fastmail-cli export [flags]
```

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--folder` | `-f` | `Inbox` | Mailbox folder name |
| `--format` | | `jsonl` | Export format: `jsonl`, `maildir`, `mbox` |
| `--output` | `-o` | stdout | Output file/directory |

## Formats

### jsonl

JSON Lines format with one email per line. Best for programmatic processing.

```bash
fastmail-cli export --format jsonl > emails.jsonl
```

Output can be piped to stdout (default) or written to a file with `--output`.

### maildir

Maildir directory structure. Standard format compatible with many email clients.

```bash
fastmail-cli export --format maildir --output ~/backup/inbox
```

Creates the standard Maildir structure:
```
~/backup/inbox/
├── cur/
├── new/
└── tmp/
```

**Note:** `--output` is required for maildir format.

### mbox

Standard Unix mbox format. Single file containing all messages.

```bash
fastmail-cli export --format mbox --output archive.mbox
```

Output can be piped to stdout or written to a file with `--output`.

## Examples

### Export Inbox to JSON Lines

```bash
# Export to stdout
fastmail-cli export --folder Inbox --format jsonl

# Export to file
fastmail-cli export --folder Inbox --format jsonl --output inbox.jsonl

# Or using redirection
fastmail-cli export --folder Inbox > inbox.jsonl
```

### Backup Archive to Maildir

```bash
fastmail-cli export --folder Archive --format maildir --output ~/backup/archive
```

### Export Sent Mail to Mbox

```bash
fastmail-cli export --folder Sent --format mbox --output sent.mbox
```

### Export All Folders

```bash
#!/bin/bash
# Backup all folders
for folder in Inbox Sent Archive Drafts; do
  fastmail-cli export \
    --folder "$folder" \
    --format maildir \
    --output ~/backup/"$folder"
done
```

### Process Exported Emails

```bash
# Export and count emails by sender domain
fastmail-cli export --format jsonl | \
  jq -r '.from[0].email // empty' | \
  cut -d@ -f2 | \
  sort | uniq -c | sort -rn | head -10
```

### Incremental Backup with Timestamps

```bash
# Export with timestamp
DATE=$(date +%Y-%m-%d)
fastmail-cli export \
  --folder Inbox \
  --format jsonl \
  --output "inbox-backup-$DATE.jsonl"
```

## Progress Output

The export command shows progress on stderr:
```
Fetching emails from Inbox...
Exporting 1234 emails...
Exported to inbox.jsonl
```

Use `--quiet` to suppress progress messages:
```bash
fastmail-cli export --quiet --format jsonl > emails.jsonl
```

## Large Exports

For large mailboxes:

1. **Use maildir format** for incremental backups - each email is a separate file
2. **Use jsonl for streaming** - process emails one line at a time
3. **Consider folder filtering** - export specific folders rather than everything

## See Also

- [CLI Reference](index.md)
- [mail](mail.md) - email operations
- [auth](auth.md) - authentication required

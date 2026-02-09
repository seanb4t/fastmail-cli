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

### maildir

Maildir directory structure. Standard format compatible with many email clients. `--output` is required for this format.

```bash
fastmail-cli export --format maildir --output ~/backup/inbox
```

### mbox

Standard Unix mbox format. Single file containing all messages.

```bash
fastmail-cli export --format mbox --output archive.mbox
```

## Examples

```bash
# Export inbox to JSON Lines
fastmail-cli export --folder Inbox --format jsonl --output inbox.jsonl

# Backup archive to Maildir
fastmail-cli export --folder Archive --format maildir --output ~/backup/archive

# Export sent mail to mbox
fastmail-cli export --folder Sent --format mbox --output sent.mbox

# Export all folders
for folder in Inbox Sent Archive Drafts; do
  fastmail-cli export --folder "$folder" --format maildir --output ~/backup/"$folder"
done

# Process exported emails
fastmail-cli export --format jsonl | \
  jq -r '.from[0].email // empty' | \
  cut -d@ -f2 | sort | uniq -c | sort -rn | head -10
```

Use `--quiet` to suppress progress messages:

```bash
fastmail-cli export --quiet --format jsonl > emails.jsonl
```

## See Also

- [CLI Reference](index.md)
- [mail](mail.md) -- email operations

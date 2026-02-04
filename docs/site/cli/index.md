# CLI Reference

Command-line reference for FastMail CLI.

## Installation

```bash
go install github.com/seanb4t/fastmail-cli/cmd/fastmail-cli@latest
```

## Synopsis

```bash
fastmail-cli [command] [options]
```

## Global Flags

These flags are available on all commands:

| Flag | Description |
|------|-------------|
| `--config <path>` | Path to configuration file |
| `--json` | Output in JSON format |
| `--quiet` | Suppress non-essential output |
| `--help` | Show help for any command |
| `--version` | Show version information |

## Commands

| Command | Description |
|---------|-------------|
| [`auth`](auth.md) | Manage authentication credentials |
| [`mail`](mail.md) | Read, send, and manage email |
| [`contacts`](contacts.md) | Manage contacts via CardDAV |
| [`masked-email`](masked-email.md) | Manage masked email addresses |
| [`export`](export.md) | Export emails to various formats |
| [`mcp`](mcp.md) | Run the MCP server |

## Configuration

FastMail CLI uses a configuration file located at `~/.config/fastmail-cli/config.yaml` by default. Override with `--config`.

```yaml
# ~/.config/fastmail-cli/config.yaml
endpoint: https://api.fastmail.com/jmap/session
carddav_endpoint: https://carddav.fastmail.com/dav/addressbooks/user
carddav_username: your-username@fastmail.com
```

Environment variables can also configure the CLI:

| Variable | Description |
|----------|-------------|
| `FASTMAIL_TOKEN` | API token (alternative to `auth login`) |
| `FASTMAIL_ENDPOINT` | JMAP API endpoint |
| `FASTMAIL_CARDDAV_USERNAME` | CardDAV username |
| `FASTMAIL_CARDDAV_ENDPOINT` | CardDAV endpoint |

## Authentication

Before using most commands, authenticate with your FastMail API token:

```bash
# Interactive login
fastmail-cli auth login

# Non-interactive login
fastmail-cli auth login --token "fmu1-xxxxxxxx"
```

Get your API token from [FastMail Settings > Password & Security > App Passwords](https://www.fastmail.com/settings/security/devicekeys).

## Output Formats

Most commands support JSON output with the `--json` flag:

```bash
# Text output (default)
fastmail-cli mail list
# M12345  Meeting tomorrow [read]
# M12346  Project update

# JSON output
fastmail-cli mail list --json
# [
#   {"id": "M12345", "subject": "Meeting tomorrow", ...},
#   {"id": "M12346", "subject": "Project update", ...}
# ]
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (authentication failure, API error, etc.) |

## Examples

### Quick Start

```bash
# Login
fastmail-cli auth login

# List recent emails
fastmail-cli mail list

# Send an email
fastmail-cli mail send --to user@example.com --subject "Hello" --body "Message"

# Create a masked email
fastmail-cli masked-email create --for-domain example.com

# Export inbox to backup
fastmail-cli export --folder Inbox --format mbox --output backup.mbox
```

### Scripting

```bash
# List emails as JSON for processing
fastmail-cli mail list --json | jq '.[] | select(.subject | contains("urgent"))'

# Check if logged in
if fastmail-cli auth status --json | jq -e '.logged_in' > /dev/null; then
  echo "Authenticated"
fi
```

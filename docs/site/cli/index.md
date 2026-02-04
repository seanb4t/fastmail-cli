# CLI Reference

`fastmail-cli` is a command-line interface for interacting with FastMail via JMAP.

## Installation

```bash
go install github.com/seanb4t/fastmail-cli/cmd/fastmail-cli@latest
```

## Quick Start

```bash
# Authenticate
fastmail-cli auth login

# List recent emails
fastmail-cli mail list

# Send an email
fastmail-cli mail send --to alice@example.com --subject "Hello" --body "Hi there!"

# List contacts
fastmail-cli contacts list

# Create a masked email
fastmail-cli masked-email create --for-domain example.com
```

## Global Options

These flags are available for all commands:

| Flag | Description |
|------|-------------|
| `--config PATH` | Config file path (default: `~/.config/fastmail-cli/config.yaml`) |
| `--json` | Output as JSON |
| `--quiet` | Suppress output |
| `--version` | Show version |

## Commands

| Command | Description |
|---------|-------------|
| [auth](auth.md) | Manage authentication credentials |
| [mail](mail.md) | Read, send, and manage email |
| [contacts](contacts.md) | Manage contacts via CardDAV |
| [masked-email](masked-email.md) | Manage masked email addresses |
| [export](export.md) | Export emails to various formats |

## Configuration

The CLI reads configuration from `~/.config/fastmail-cli/config.yaml` by default. Override with `--config`.

```yaml
# Example config.yaml
endpoint: https://api.fastmail.com/jmap/session
carddav_endpoint: https://carddav.fastmail.com/dav/
carddav_username: username@fastmail.com
```

Environment variables:

| Variable | Description |
|----------|-------------|
| `FASTMAIL_API_TOKEN` | API token (alternative to `auth login`) |
| `FASTMAIL_ENDPOINT` | JMAP endpoint URL |
| `FASTMAIL_CARDDAV_ENDPOINT` | CardDAV endpoint URL |
| `FASTMAIL_CARDDAV_USERNAME` | CardDAV username |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (see stderr for details) |

## See Also

- [Getting Started](../getting-started.md)
- [FastMail API Documentation](https://www.fastmail.com/dev/)

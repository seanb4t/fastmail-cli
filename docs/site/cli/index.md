# CLI Reference

`fastmail-cli` is a command-line interface for interacting with FastMail via JMAP, CardDAV, and CalDAV.

## Installation

```bash
go install github.com/seanb4t/fastmail-cli/cmd/fastmail-cli@latest
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
| [mail](mail.md) | Read, send, search, and manage email |
| [mailbox](mailbox.md) | Manage mailbox folders |
| [contacts](contacts.md) | Manage contacts via CardDAV |
| [calendar](calendar.md) | Manage calendars and events via CalDAV |
| [masked-email](masked-email.md) | Manage masked email addresses |
| [vacation](vacation.md) | Manage vacation/out-of-office auto-reply |
| [identity](identity.md) | Manage sender identities |
| [filter](filter.md) | Manage Sieve filter scripts |
| [export](export.md) | Export emails to various formats |
| [account](account.md) | View account information (quota) |

## Configuration

The CLI reads configuration from `~/.config/fastmail-cli/config.yaml` by default. Override with `--config`.

```yaml
# Example config.yaml
endpoint: https://api.fastmail.com/jmap/session
carddav_endpoint: https://carddav.fastmail.com/dav/
caldav_endpoint: https://caldav.fastmail.com/dav/
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
| 0 | Success / Authenticated |
| 1 | Error or No token stored |
| 2 | Token expired or revoked |
| 3 | Cannot reach FastMail API |
| 4 | Authentication failed |

## See Also

- [Getting Started](../getting-started.md)
- [FastMail API Documentation](https://www.fastmail.com/dev/)

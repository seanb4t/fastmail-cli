# fastmail-cli

A full-featured CLI, interactive TUI, and MCP server for FastMail — supporting email, contacts, calendars, masked emails, filters, and more via JMAP, CardDAV, and CalDAV.

[![CI](https://github.com/seanb4t/fastmail-cli/actions/workflows/ci.yaml/badge.svg)](https://github.com/seanb4t/fastmail-cli/actions/workflows/ci.yaml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/seanb4t/fastmail-cli)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Features

| | | |
|---|---|---|
| **Interactive TUI** — Browse mailboxes, read/reply/compose, manage attachments | **Full CLI** — Scriptable commands for mail, contacts, calendars, filters, and more | **MCP Server** — 49 tools + resource URIs for AI agent integration |
| **Three Protocols** — JMAP, CardDAV, CalDAV under one interface | **Masked Email** — Create, enable, disable throwaway addresses | **Secure Auth** — Keychain-first credential storage with file fallback |

## Quick Start

```bash
# Authenticate (stores token in system keychain)
fastmail-cli auth login

# Launch interactive TUI (also the default with no args)
fastmail-cli

# Or use CLI commands directly
fastmail-cli mail list
fastmail-cli contacts list
fastmail-cli calendar list-events
```

## Installation

### Using Go Install

```bash
go install github.com/seanb4t/fastmail-cli/cmd/fastmail-cli@latest
```

### Binary Releases

Download pre-built binaries from [GitHub Releases](https://github.com/seanb4t/fastmail-cli/releases). Builds are available for Linux and macOS on amd64 and arm64.

### From Source

```bash
git clone https://github.com/seanb4t/fastmail-cli.git
cd fastmail-cli
task build    # requires https://taskfile.dev
```

## Architecture

FastMail exposes three distinct protocols for different data types. This CLI unifies them behind a single interface so you don't have to think about which protocol handles what.

```
                    ┌─────────────────────────┐
                    │   CLI / TUI / MCP Server │
                    └────────────┬────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │     pkg/fastmail        │
                    │   (unified public API)   │
                    └──┬─────────┬─────────┬──┘
                       │         │         │
              ┌────────▼──┐ ┌───▼────┐ ┌──▼────────┐
              │   JMAP    │ │ CardDAV│ │  CalDAV   │
              │ internal/ │ │internal│ │ internal/ │
              │  jmap/    │ │ /dav/  │ │   dav/    │
              └───────────┘ └────────┘ └───────────┘
               Email,        Contacts    Calendars,
               Mailboxes,                Events
               Masked Email,
               Filters,
               Identities
```

- **JMAP** — JSON Meta Application Protocol for email, mailboxes, masked emails, filters, identities, and vacation responses. Supports session caching, method call batching with back-references, and capability discovery.
- **CardDAV** — WebDAV-based protocol for contacts. Parses vCard format internally and exposes clean domain types.
- **CalDAV** — WebDAV-based protocol for calendars and events. Parses iCalendar format internally.

The public API in `pkg/fastmail` hides all protocol details behind a service pattern:

```go
client := fastmail.NewClient(endpoint, accessToken)
client.Connect(ctx)
emails, _ := client.Mail().List(ctx, ...)
contacts, _ := client.Contacts().List(ctx, ...)
events, _ := client.Calendar().ListEvents(ctx, ...)
```

## Interactive TUI

Running `fastmail-cli` with no arguments launches the interactive terminal UI (requires authentication).

| View | Key | Action |
|------|-----|--------|
| **Mailbox List** | `enter` | Open mailbox |
| | `/` | Filter mailboxes |
| **Email List** | `enter` | Open email |
| | `a` | Archive |
| | `x` | Delete (press twice) |
| | `.` | Toggle read/unread |
| | `f` | Toggle flag |
| | `m` | Move to mailbox |
| | `c` | Compose new email |
| | `/` | Search emails |
| **Email Reader** | `j/k` | Scroll line |
| | `d/u` | Half page down/up |
| | `r` / `R` | Reply / Reply All |
| | `A` | View attachments |
| | `t` | View thread |
| **Compose** | `tab` | Next field |
| | `ctrl+s` | Send email |
| | `esc` | Cancel |
| **Global** | `?` | Help overlay |
| | `q` | Quit / Back |

## CLI Commands

| Command | Description |
|---------|-------------|
| `mail` | List, read, send, reply, forward, archive, flag, move, delete, import |
| `mailbox` | List, create, rename, delete mailboxes |
| `contacts` | List, show, create, update, delete contacts |
| `calendar` | List calendars, list/show/create/update/delete events |
| `masked-email` | List, create, enable, disable, delete masked email addresses |
| `vacation` | Show, enable, disable vacation/out-of-office response |
| `identity` | List and update sender identities |
| `filter` | List, show, create, activate, deactivate Sieve filter scripts |
| `export` | Export emails in JSONL, Maildir, or mbox format |
| `account` | Show storage quota |
| `auth` | Login, logout, check authentication status |
| `mcp` | Run MCP server for AI assistant integration |
| `tui` | Launch interactive terminal UI |
| `completion` | Generate shell completions (bash, zsh, fish, powershell) |

### Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Config file path |
| `--json` | Output as JSON |
| `--quiet` | Suppress output |
| `-v, --version` | Show version |

## Configuration

### API Token

1. Log in to [FastMail](https://www.fastmail.com) web interface
2. Navigate to **Settings** > **Privacy & Security** > **API Tokens**
3. Create a new token with the required scopes

### Authentication

```bash
# Store token securely (uses system keychain on macOS)
fastmail-cli auth login

# Or set via environment variable
export FASTMAIL_TOKEN=your-api-token
```

### Config File

Located at `~/.config/fastmail-cli/config.yaml`:

```yaml
endpoint: https://api.fastmail.com/jmap/session
carddav_endpoint: https://carddav.fastmail.com/dav/
carddav_username: username@fastmail.com
```

## MCP Integration

The MCP server exposes 49 tools and 7 resource URIs for AI agent integration over stdio JSON-RPC 2.0.

```bash
fastmail-cli mcp
```

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "fastmail": {
      "command": "fastmail-cli",
      "args": ["mcp"]
    }
  }
}
```

### Claude Code

Add to your project or global `.claude/settings.json`:

```json
{
  "mcpServers": {
    "fastmail": {
      "command": "fastmail-cli",
      "args": ["mcp"]
    }
  }
}
```

### Resource URIs

| URI | Description |
|-----|-------------|
| `fastmail://inbox` | Recent inbox messages |
| `fastmail://mail/{id}` | Single email content |
| `fastmail://contacts` | Contact list |
| `fastmail://contact/{id}` | Single contact details |
| `fastmail://calendar/today` | Today's calendar events |
| `fastmail://masked-emails` | Masked email addresses |
| `fastmail://masked-email/{id}` | Single masked email details |

## Shell Completion

```bash
# Bash
fastmail-cli completion bash > /etc/bash_completion.d/fastmail-cli

# Zsh
fastmail-cli completion zsh > "${fpath[1]}/_fastmail-cli"

# Fish
fastmail-cli completion fish > ~/.config/fish/completions/fastmail-cli.fish

# PowerShell
fastmail-cli completion powershell > fastmail-cli.ps1
```

## Development

This project uses [Task](https://taskfile.dev) as the build tool (not Make).

```bash
task all              # lint → test → build
task test             # run tests with -race -cover
task lint             # golangci-lint run
task mocks            # regenerate mocks via mockery
task build            # build binary to ./bin/fastmail-cli
task docs:serve       # serve documentation site locally
```

Run a single test:

```bash
go test -run TestName ./internal/jmap/
```

## Documentation

Full documentation is available at [fastmail-cli.pages.dev](https://fastmail-cli.pages.dev).

## License

[MIT](LICENSE)

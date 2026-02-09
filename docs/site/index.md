# FastMail CLI

A full-featured CLI and MCP server for FastMail -- supporting email, contacts, calendars, masked emails, filters, and more via JMAP, CardDAV, and CalDAV.

## Overview

FastMail CLI provides a powerful terminal interface for managing your FastMail account:

- **Full CLI** -- Scriptable commands for mail, contacts, calendars, filters, and more
- **MCP Server** -- Tools and resource URIs for AI agent integration (Claude Desktop, Claude Code)
- **Three Protocols** -- JMAP, CardDAV, CalDAV unified under one interface
- **Masked Email** -- Create, enable, disable throwaway addresses
- **Secure Auth** -- Keychain-first credential storage with file fallback
- **Go Library** -- Public API in `pkg/fastmail` for programmatic access

## Quick Start

```bash
# Authenticate (stores token in system keychain)
fastmail-cli auth login

# List recent emails
fastmail-cli mail list

# Send an email
fastmail-cli mail send --to alice@example.com --subject "Hello" --body "Hi there!"

# List contacts
fastmail-cli contacts list

# List calendar events for today
fastmail-cli calendar list

# Create a masked email
fastmail-cli masked-email create --for-domain example.com
```

## Architecture

FastMail exposes three distinct protocols for different data types. This CLI unifies them behind a single interface.

```
                    +-----------------------------+
                    |   CLI / MCP Server          |
                    +-------------+---------------+
                                  |
                    +-------------v---------------+
                    |     pkg/fastmail            |
                    |   (unified public API)      |
                    +--+----------+----------+----+
                       |          |          |
              +--------v--+ +----v----+ +---v---------+
              |   JMAP    | | CardDAV | |  CalDAV     |
              | internal/ | |internal/| | internal/   |
              |  jmap/    | |  dav/   | |   dav/      |
              +-----------+ +---------+ +-------------+
               Email,        Contacts    Calendars,
               Mailboxes,                Events
               Masked Email,
               Filters,
               Identities
```

- **JMAP** -- Email, mailboxes, masked emails, filters, identities, vacation responses, and storage quota
- **CardDAV** -- Contacts management with vCard parsing
- **CalDAV** -- Calendar and event management with iCalendar parsing

## CLI Commands

| Command | Description |
|---------|-------------|
| `auth` | Login, logout, check authentication status |
| `mail` | List, show, search, send, reply, move, delete, flag, thread, attachments, download, upload, import, scheduled sends |
| `mailbox` | List, create, rename, delete mailboxes |
| `contacts` | List, show, create, update, delete contacts |
| `calendar` | List calendars, list/show/create/update/delete events |
| `masked-email` | List, create, enable, disable, delete masked email addresses |
| `vacation` | Show, enable, disable vacation/out-of-office response |
| `identity` | List and update sender identities |
| `filter` | List, show, create, activate, deactivate, validate, delete Sieve filter scripts |
| `export` | Export emails in JSONL, Maildir, or mbox format |
| `account` | Show storage quota |
| `mcp` | Run MCP server for AI assistant integration |
| `completion` | Generate shell completions (bash, zsh, fish, powershell) |

### Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Config file path |
| `--json` | Output as JSON |
| `--quiet` | Suppress output |
| `-v, --version` | Show version |

## MCP Server

The MCP server exposes tools and resource URIs for AI agent integration over stdio JSON-RPC 2.0.

```bash
fastmail-cli mcp
```

See [MCP Integration](mcp/index.md) for configuration and usage.

## Next Steps

- [Getting Started](getting-started.md) -- Installation and configuration guide
- [CLI Reference](cli/index.md) -- Complete command documentation
- [MCP Integration](mcp/index.md) -- Using FastMail with AI assistants
- [Go Library](api/index.md) -- Programmatic API documentation

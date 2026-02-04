# FastMail CLI

A command-line interface for interacting with FastMail.

## Overview

FastMail CLI provides a powerful terminal interface for managing your FastMail account, including:

- **Mail Operations** - Read, search, and manage emails
- **Contacts** - Browse and manage your address book
- **Calendars** - View and manage calendar events
- **MCP Server** - Model Context Protocol integration for AI assistants

## Quick Start

```bash
# Install
go install github.com/user/fastmail-cli/cmd/fastmail-cli@latest

# Configure credentials
fastmail-cli config set api-token YOUR_API_TOKEN

# List mailboxes
fastmail-cli mail mailboxes

# Read recent emails
fastmail-cli mail list --limit 10
```

## Features

### Command-Line Interface

Full-featured CLI for managing FastMail resources:

- Mail management (list, read, search, move, delete)
- Contact operations (list, show, create, update)
- Calendar access (list events, create, update)

### MCP Server

Integrate FastMail with AI assistants via the Model Context Protocol:

```bash
# Start MCP server
fastmail-cli mcp serve
```

The MCP server exposes FastMail resources and tools to compatible AI clients.

## Next Steps

- [Getting Started](getting-started.md) - Installation and configuration guide
- [CLI Reference](reference/cli.md) - Complete command documentation
- [MCP Integration](reference/mcp.md) - Using FastMail with AI assistants

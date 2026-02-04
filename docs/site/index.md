# FastMail CLI

A command-line interface for interacting with FastMail via JMAP.

## Overview

FastMail CLI provides a powerful terminal interface for managing your FastMail account, including:

- **Mail Operations** - List, send, reply, and export emails
- **Contacts** - Manage your address book via CardDAV
- **Masked Emails** - Create and manage masked email addresses
- **MCP Server** - Model Context Protocol integration for AI assistants

## Quick Start

```bash
# Install
go install github.com/seanb4t/fastmail-cli/cmd/fastmail-cli@latest

# Authenticate
fastmail-cli auth login

# List recent emails
fastmail-cli mail list

# Send an email
fastmail-cli mail send --to alice@example.com --subject "Hello" --body "Hi there!"
```

## Features

### Command-Line Interface

Full-featured CLI for managing FastMail resources:

- Mail management (list, send, reply, export)
- Contact operations (list, show, create, update, delete)
- Masked email management (create, enable, disable, delete)

### MCP Server

Integrate FastMail with AI assistants via the Model Context Protocol:

```bash
# Start MCP server
fastmail-cli mcp
```

The MCP server exposes FastMail resources and tools to compatible AI clients like Claude Desktop.

## Next Steps

- [Getting Started](getting-started.md) - Installation and configuration guide
- [CLI Reference](cli/index.md) - Complete command documentation
- [MCP Integration](mcp/index.md) - Using FastMail with AI assistants

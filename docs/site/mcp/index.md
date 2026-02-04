# MCP Integration

FastMail CLI includes a Model Context Protocol (MCP) server that exposes your FastMail account to AI assistants like Claude.

## Overview

The MCP server provides:

- **Resources** - Read-only access to mailbox, contacts, and calendar data
- **Tools** - Operations for sending email, managing contacts, and more

## Quick Start

1. [Set up the MCP server](setup.md) in Claude Desktop
2. Ask Claude to check your inbox or send emails
3. Use resources for context and tools for actions

## Architecture

```
┌─────────────────────────────────────────────────┐
│              AI Assistant (Claude)              │
└─────────────────────┬───────────────────────────┘
                      │ JSON-RPC 2.0 over stdio
┌─────────────────────▼───────────────────────────┐
│          fastmail-cli mcp serve                 │
│  ┌─────────────┬─────────────────────────┐      │
│  │   Tools     │      Resources          │      │
│  │  (actions)  │   (read-only data)      │      │
│  └─────────────┴─────────────────────────┘      │
└─────────────────────┬───────────────────────────┘
                      │ JMAP API
┌─────────────────────▼───────────────────────────┐
│               FastMail Server                   │
└─────────────────────────────────────────────────┘
```

## When to Use Resources vs Tools

| Use Case | Use |
|----------|-----|
| Give Claude context about your inbox | Resource: `fastmail://inbox` |
| Read a specific email | Resource: `fastmail://mail/{id}` |
| Send a new email | Tool: `mail_send` |
| Reply to an email | Tool: `mail_reply` |
| Search for emails | Tool: `mail_search` |
| Archive or delete email | Tools: `mail_move`, `mail_delete` |

## Documentation

- [Setup Guide](setup.md) - Configure Claude Desktop
- [Tools Reference](tools.md) - Available tools with schemas
- [Resources Reference](resources.md) - Available resources with URIs

## Example Prompts

Once configured, try these prompts with Claude:

- "Check my inbox for unread emails"
- "Search for emails from Alice about the project"
- "Send a reply to the last email from Bob"
- "Create a masked email for signing up to example.com"
- "Show my contacts"

## Security Considerations

- The MCP server runs locally and communicates via stdio
- Your API token is stored in the CLI configuration
- Claude cannot access your account without the MCP server running
- Review tool calls before Claude executes sensitive operations

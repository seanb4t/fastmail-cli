# MCP Integration Guide

This guide explains how to integrate the fastmail-cli MCP server with AI clients like Claude Desktop.

## Overview

The Model Context Protocol (MCP) allows AI assistants to access Fastmail data through a standardized interface. The fastmail-cli implements an MCP server that exposes:

- **Resources**: Read-only access to emails, contacts, calendar events, and masked emails
- **Tools**: Operations like sending email, archiving, and flagging (coming soon)

## Quick Start

### 1. Build the CLI

```bash
go build -o fastmail-cli ./cmd/fastmail-cli
```

### 2. Configure Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "fastmail": {
      "command": "/path/to/fastmail-cli",
      "args": ["mcp"],
      "env": {
        "FASTMAIL_TOKEN": "your-api-token",
        "FASTMAIL_ACCOUNT_ID": "your-account-id"
      }
    }
  }
}
```

### 3. Restart Claude Desktop

The Fastmail resources will be available in your conversations.

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `FASTMAIL_TOKEN` | Yes | Fastmail API token |
| `FASTMAIL_ACCOUNT_ID` | Yes | Your Fastmail account ID |
| `FASTMAIL_API_URL` | No | API endpoint (default: `https://api.fastmail.com`) |

### Getting API Credentials

1. Log into Fastmail web interface
2. Go to Settings → Password & Security → API Tokens
3. Create a new token with required scopes
4. Copy your Account ID from Settings → Account

## Available Resources

### fastmail://inbox

Returns the 10 most recent emails from your inbox.

**Example output:**
```markdown
# Inbox (10 messages)

## 1. Meeting tomorrow
- ID: M123abc
- Received: 2026-02-04T10:30:00Z
- Preview: Hi, just confirming our meeting...
- Status: unread
```

### fastmail://mail/{id}

Returns the full content of a specific email.

**URI Pattern:** `fastmail://mail/M123abc`

**Example output:**
```markdown
# Meeting tomorrow

- ID: M123abc
- Thread: T456def
- Received: 2026-02-04T10:30:00Z
- Size: 2048 bytes
- Status: unread

## Preview

Hi, just confirming our meeting for tomorrow at 2pm...
```

### fastmail://contacts

Returns your address book contacts.

**Example output:**
```markdown
# Contacts (25 total)

## 1. John Doe
- ID: C789ghi
- Email: john@example.com
- Phone: +1-555-0123
```

### fastmail://calendar/today

Returns today's calendar events.

**Note:** Requires CalDAV configuration for full functionality.

### fastmail://masked-emails

Returns your masked email addresses.

**Example output:**
```markdown
# Masked Emails (5 total)

## 1. shopping.abc123@fastmail.com
- ID: ME456jkl
- State: enabled
- Domain: amazon.com
- Description: Amazon shopping
```

## Protocol Details

### Transport

The MCP server uses stdio transport:
- Reads JSON-RPC 2.0 requests from stdin (one per line)
- Writes JSON-RPC 2.0 responses to stdout (one per line)

### Supported Methods

| Method | Description |
|--------|-------------|
| `initialize` | Capability negotiation |
| `tools/list` | List available tools |
| `tools/call` | Execute a tool |
| `resources/list` | List available resources |
| `resources/read` | Read a resource by URI |

### Example Request/Response

**List Resources:**
```json
{"jsonrpc":"2.0","id":1,"method":"resources/list"}
```

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "resources": [
      {
        "uri": "fastmail://inbox",
        "name": "Recent Inbox",
        "description": "Recent emails from the inbox folder",
        "mimeType": "text/plain"
      }
    ]
  }
}
```

**Read Resource:**
```json
{"jsonrpc":"2.0","id":2,"method":"resources/read","params":{"uri":"fastmail://inbox"}}
```

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "contents": [
      {
        "uri": "fastmail://inbox",
        "mimeType": "text/plain",
        "text": "# Inbox (10 messages)\n\n## 1. Meeting tomorrow..."
      }
    ]
  }
}
```

## Error Handling

Errors are returned as JSON-RPC error responses:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Unknown resource: fastmail://invalid"
  }
}
```

| Code | Meaning |
|------|---------|
| -32700 | Parse error (invalid JSON) |
| -32600 | Invalid request |
| -32601 | Method not found |
| -32602 | Invalid params |
| -32603 | Internal error |

## Troubleshooting

### Server not starting

1. Check the CLI path is correct in config
2. Verify environment variables are set
3. Test manually: `echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | fastmail-cli mcp`

### Authentication errors

1. Verify API token has correct scopes
2. Check account ID matches token
3. Ensure token hasn't expired

### No resources appearing

1. Restart Claude Desktop after config changes
2. Check MCP server logs for errors
3. Verify Fastmail API is accessible

## Security Considerations

- API tokens are sensitive; don't commit to version control
- Use environment variables or secure credential storage
- The MCP server only exposes read operations by default
- Tool operations (mutations) require explicit configuration

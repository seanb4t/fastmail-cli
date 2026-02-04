# mcp

Run the Model Context Protocol (MCP) server.

## Synopsis

```bash
fastmail-cli mcp
```

## Description

The `mcp` command starts an MCP server that exposes FastMail functionality to AI agents. The server communicates via stdin/stdout using JSON-RPC 2.0 protocol.

## Usage

The MCP server is designed to be invoked by AI clients (like Claude Desktop) rather than run directly:

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

## Available Tools

The MCP server provides these tools for AI agents:

### Mail Tools

| Tool | Description |
|------|-------------|
| `mail_list` | List emails from a folder |
| `mail_search` | Search emails by query |
| `mail_get` | Get a single email by ID |
| `mail_send` | Send a new email |
| `mail_move` | Move email to a folder |
| `mail_delete` | Delete an email |

### Masked Email Tools

| Tool | Description |
|------|-------------|
| `masked_email_list` | List all masked emails |
| `masked_email_create` | Create a new masked email |
| `masked_email_enable` | Enable a masked email |
| `masked_email_disable` | Disable a masked email |
| `masked_email_delete` | Delete a masked email |

## Tool Details

### mail_list

List emails from a mailbox folder.

**Parameters:**

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `folder` | string | No | `Inbox` | Folder name |
| `limit` | integer | No | `10` | Max emails to return |

### mail_search

Search emails by query text.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `query` | string | Yes | Search query (e.g., `from:alice subject:meeting`) |
| `limit` | integer | No | Max results (default: 10) |

### mail_get

Get a single email by ID.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `id` | string | Yes | Email ID |

### mail_send

Send a new email.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `to` | string | Yes | Recipient email |
| `subject` | string | Yes | Email subject |
| `body` | string | Yes | Email body text |

### mail_move

Move an email to a different folder.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `id` | string | Yes | Email ID |
| `folder` | string | Yes | Target folder name |

### mail_delete

Delete an email.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `id` | string | Yes | Email ID |

### masked_email_list

List all masked email addresses. No parameters.

### masked_email_create

Create a new masked email address.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `domain` | string | No | Associated domain |
| `description` | string | No | Description/note |

### masked_email_enable

Enable a masked email address.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `id` | string | Yes | Masked email ID |

### masked_email_disable

Disable a masked email address.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `id` | string | Yes | Masked email ID |

### masked_email_delete

Delete a masked email address.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `id` | string | Yes | Masked email ID |

## Configuration

The MCP server uses the same authentication as the CLI. Run `fastmail-cli auth login` before starting the server.

## Graceful Shutdown

The server handles SIGTERM and SIGINT for graceful shutdown. Send either signal to stop the server cleanly.

## See Also

- [MCP Integration Guide](../reference/mcp.md) - Detailed MCP documentation
- [auth](auth.md) - Authentication setup

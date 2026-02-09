# MCP Integration

FastMail CLI includes a Model Context Protocol (MCP) server that exposes your FastMail account to AI assistants like Claude.

## Overview

The MCP server provides:

- **Tools** -- Operations for email, contacts, calendar, masked email, filters, identities, vacation, and account quota
- **Resources** -- Read-only access to inbox, individual emails, contacts, calendar, and masked email data

## Quick Start

1. [Set up the MCP server](setup.md) in Claude Desktop or Claude Code
2. Ask Claude to check your inbox or send emails
3. Use resources for context and tools for actions

## Architecture

```
+---------------------------------------------------+
|              AI Assistant (Claude)                 |
+-------------------------+-------------------------+
                          | JSON-RPC 2.0 over stdio
+-------------------------v-------------------------+
|              fastmail-cli mcp                     |
|  +-----------------+-------------------------+    |
|  |     Tools       |      Resources          |    |
|  |   (actions)     |   (read-only data)      |    |
|  +-----------------+-------------------------+    |
+-------------------------+-------------------------+
                          | JMAP / CardDAV / CalDAV
+-------------------------v-------------------------+
|               FastMail Server                     |
+---------------------------------------------------+
```

## Tool Categories

| Category | Tool Count | Description |
|----------|-----------|-------------|
| Mail | 14 | List, get, search, send, reply, move, delete, flag, thread, attachments, download, import, upload, scheduled sends |
| Mailbox | 4 | List, create, rename, delete folders |
| Masked Email | 6 | List, get, create, enable, disable, delete |
| Contacts | 6 | List, get, create, update, delete, search |
| Calendar | 6 | List calendars, list/get/create/update/delete events |
| Vacation | 2 | Get status, set (enable/disable) |
| Identity | 2 | List, update sender identities |
| Filter | 7 | List, get, create, activate, deactivate, validate, delete Sieve scripts |
| Account | 1 | Get storage quota |

**Total: 48 tools**

## Resource URIs

| URI | Description |
|-----|-------------|
| `fastmail://inbox` | Recent inbox messages |
| `fastmail://mail/{id}` | Single email content |
| `fastmail://contacts` | Contact list |
| `fastmail://contact/{id}` | Single contact details |
| `fastmail://calendar/today` | Today's calendar events |
| `fastmail://masked-emails` | Masked email addresses |
| `fastmail://masked-email/{id}` | Single masked email details |

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

- [Setup Guide](setup.md) -- Configure Claude Desktop and Claude Code
- [Tools Reference](tools.md) -- All tools with input schemas
- [Resources Reference](resources.md) -- All resources with URIs

## Example Prompts

Once configured, try these prompts with Claude:

- "Check my inbox for unread emails"
- "Search for emails from Alice about the project"
- "Send a reply to the last email from Bob"
- "Create a masked email for signing up to example.com"
- "Show my contacts"
- "What's on my calendar today?"
- "Check my storage quota"
- "List my Sieve filter scripts"
- "Set my vacation response for next week"

## Security Considerations

- The MCP server runs locally and communicates via stdio
- Your API token is stored in the CLI configuration (keychain or file)
- Claude cannot access your account without the MCP server running
- Review tool calls before Claude executes sensitive operations

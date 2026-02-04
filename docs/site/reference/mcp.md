# MCP Integration

Model Context Protocol integration for AI assistants.

!!! note "Coming Soon"
    Complete MCP integration documentation is in development.

## Overview

FastMail CLI includes an MCP server that exposes FastMail resources and tools to AI assistants.

## Starting the Server

```bash
fastmail-cli mcp serve
```

## Available Resources

| Resource | Description |
|----------|-------------|
| `fastmail://mailboxes` | List of mailboxes |
| `fastmail://contacts` | Contact list |
| `fastmail://calendars` | Calendar list |

## Available Tools

| Tool | Description |
|------|-------------|
| `mail_list` | List emails |
| `mail_read` | Read email content |
| `mail_search` | Search emails |
| `contacts_list` | List contacts |
| `contacts_show` | Show contact details |

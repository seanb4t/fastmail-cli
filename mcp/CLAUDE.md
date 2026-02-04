# mcp/

Model Context Protocol server implementation.

## Purpose

Expose Fastmail functionality to AI agents via MCP:
- Tools for email operations
- Resources for mailbox access
- Prompts for common workflows

## Key Types

| Type | Description |
|------|-------------|
| `Server` | MCP server implementation |
| `Tools` | Available MCP tools |
| `Resources` | Exposed MCP resources |

## Conventions

- Follow MCP specification exactly
- Tools should be atomic operations
- Resources for read-only access
- Validate all inputs from agents

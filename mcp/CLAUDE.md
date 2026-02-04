# mcp/

Model Context Protocol (MCP) server implementation for exposing Fastmail functionality to AI agents.

## Purpose

Expose Fastmail data and operations to AI agents via the [MCP specification](https://spec.modelcontextprotocol.io/):
- **Resources** for read-only mailbox/contact/calendar access
- **Tools** for email operations (send, archive, flag)
- **Prompts** for common workflow templates

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     AI Client                           │
│              (Claude Desktop, etc.)                     │
└─────────────────────┬───────────────────────────────────┘
                      │ JSON-RPC 2.0 over stdio
┌─────────────────────▼───────────────────────────────────┐
│                    Server                               │
│                 (server.go)                             │
│  ┌─────────────┬─────────────┬─────────────┐           │
│  │ initialize  │ tools/*     │ resources/* │           │
│  └─────────────┴──────┬──────┴──────┬──────┘           │
└───────────────────────┼─────────────┼───────────────────┘
                        │             │
              ┌─────────▼─────┐ ┌─────▼─────────────┐
              │ ToolHandler   │ │ ResourceRegistry  │
              │ (user-defined)│ │ (resources.go)    │
              └───────────────┘ └───────────────────┘
```

## Key Types

| Type | File | Description |
|------|------|-------------|
| `Server` | server.go | MCP server with stdio transport |
| `Tool` | protocol.go | Tool definition with JSON Schema input |
| `Resource` | protocol.go | Resource definition (URI, name, mime type) |
| `ResourceRegistry` | resources.go | Manages resources and templates |
| `Request`/`Response` | protocol.go | JSON-RPC 2.0 message types |

## Protocol Methods

| Method | Handler | Description |
|--------|---------|-------------|
| `initialize` | `handleInitialize` | Capability negotiation |
| `tools/list` | `handleToolsList` | List registered tools |
| `tools/call` | `handleToolsCall` | Execute a tool |
| `resources/list` | `handleResourcesList` | List available resources |
| `resources/read` | `handleResourcesRead` | Read resource content |

## Available Resources

| URI | Name | Description |
|-----|------|-------------|
| `fastmail://inbox` | Recent Inbox | Last 10 inbox messages |
| `fastmail://mail/{id}` | Email Message | Single email by ID |
| `fastmail://contacts` | Contacts | Address book entries |
| `fastmail://calendar/today` | Today's Events | Calendar events for today |
| `fastmail://masked-emails` | Masked Emails | Masked email addresses |

## Usage

### Creating a Server

```go
server := mcp.NewServer("fastmail-cli", "1.0.0")

// Register a tool
tool := mcp.NewTool("email/send", "Send an email").
    WithProperty("to", "string", "Recipient address").
    WithProperty("subject", "string", "Email subject").
    WithProperty("body", "string", "Email body").
    WithRequired("to", "subject", "body")

server.RegisterTool(tool, func(ctx context.Context, args map[string]any) (any, error) {
    to := args["to"].(string)
    // ... send email
    return "Email sent", nil
})

// Register resources
registry := mcp.NewResourceRegistry(fastmailClient)
for _, res := range registry.List() {
    server.RegisterResource(&res, registry.Read)
}

// Run with stdio
server.Run(ctx, os.Stdin, os.Stdout)
```

### Tool Input Schema

Tools use JSON Schema for input validation:

```go
tool := mcp.NewTool("email/archive", "Archive emails").
    WithProperty("ids", "array", "Email IDs to archive").
    WithRequired("ids")
```

### Resource Templates

Parameterized resources use URI templates:

```go
registry.registerTemplate(
    mcp.ResourceTemplate{
        URITemplate: "fastmail://mail/{id}",
        Name:        "Email Message",
        Description: "Content of a specific email",
    },
    regexp.MustCompile(`^fastmail://mail/(?P<id>[^/]+)$`),
    handler,
)
```

## Conventions

- Follow MCP specification exactly
- Tools for mutations (send, archive, flag)
- Resources for read-only access (list, get)
- Validate all inputs from agents
- Return structured errors, not panics
- Format output for LLM readability (markdown)

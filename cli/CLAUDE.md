# cli/

Command-line interface implementation using cobra.

## Purpose

Define CLI commands and flags:
- Root command with global flags
- Subcommands for each feature area
- Help and completion generation

## Commands

| File | Command | Description |
|------|---------|-------------|
| root.go | `fastmail` | Root command with global flags |
| auth.go | `auth` | Configure API credentials |
| mail.go | `mail` | Email operations (list, read, send, archive, flag) |
| contacts.go | `contacts` | Contact management (list, search) |
| masked_email.go | `masked-email` | Masked email operations (list, create, disable) |
| export.go | `export` | Data export operations |
| mcp.go | `mcp` | Start MCP server for AI agents |

## Key Types

| Type | Description |
|------|-------------|
| `Execute()` | Run the root command |
| Root command | Global flags: --config, --json, --quiet |

## Conventions

- Use cobra for command definition
- Global flags on root, specific flags on subcommands
- RunE for commands that can error
- Short help under 80 chars, long help in separate files

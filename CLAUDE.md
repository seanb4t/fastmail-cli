# fastmail-cli

A command-line interface and Go library for interacting with Fastmail.

## Project Structure

```
fastmail-cli/
├── cli/          # CLI commands (cobra-based)
├── cmd/          # Binary entrypoints
├── docs/         # Documentation (reference + Zensical site)
├── internal/     # Private implementation packages
│   ├── auth/     # Credential storage (keychain/file)
│   ├── config/   # Configuration loading (viper)
│   ├── dav/      # CardDAV/CalDAV clients
│   ├── jmap/     # JMAP protocol client
│   └── output/   # Output formatters
├── mcp/          # MCP server for AI agents
├── pkg/          # Public library (fastmail/)
└── testdata/     # Test fixtures
```

## Key Commands

| Command | Description |
|---------|-------------|
| `fastmail auth` | Configure API credentials |
| `fastmail mail` | Email operations (list, read, send, archive) |
| `fastmail contacts` | Contact management |
| `fastmail masked-email` | Masked email management |
| `fastmail export` | Data export operations |
| `fastmail mcp` | Start MCP server for AI agents |

## Architecture

- **JMAP** for email/mailbox operations (modern JSON API)
- **CardDAV/CalDAV** for contacts/calendar (WebDAV protocols)
- **MCP** for AI agent integration (stdio JSON-RPC)

## Development

```bash
# Run tests
go test ./...

# Build
go build ./cmd/fastmail-cli

# Lint
golangci-lint run
```

## Conventions

- Internal packages hide protocol complexity
- Public API in `pkg/fastmail` for library consumers
- Context as first parameter for cancellable operations
- Structured errors via `samber/oops`
- Table-driven tests with fixtures

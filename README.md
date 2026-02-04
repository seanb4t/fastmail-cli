# fastmail-cli

A command-line interface for interacting with FastMail via JMAP.

## Features

- **Mail** - List, send, reply, and export emails
- **Contacts** - Manage contacts via CardDAV
- **Masked Email** - Create, enable, disable, and delete masked emails
- **MCP Server** - Model Context Protocol integration for AI assistants

## Quick Start

```bash
# Build from source
go build -o fastmail-cli ./cmd/fastmail-cli

# Authenticate
fastmail-cli auth login

# List recent emails
fastmail-cli mail list

# Send an email
fastmail-cli mail send --to alice@example.com --subject "Hello" --body "Hi there!"

# List contacts
fastmail-cli contacts list

# Create a masked email
fastmail-cli masked-email create --for-domain example.com

# Run MCP server for AI integration
fastmail-cli mcp
```

## Installation

### From Source

```bash
git clone https://github.com/seanb4t/fastmail-cli.git
cd fastmail-cli
go build -o fastmail-cli ./cmd/fastmail-cli
```

### Using Go Install

```bash
go install github.com/seanb4t/fastmail-cli/cmd/fastmail-cli@latest
```

## Configuration

### API Token

Get an API token from FastMail:
1. Log in to FastMail web interface
2. Navigate to **Settings** → **Privacy & Security** → **API Tokens**
3. Create a new token with required scopes

### Authentication Methods

```bash
# Store token securely (uses keychain on macOS)
fastmail-cli auth login

# Or use environment variable
export FASTMAIL_TOKEN=your-api-token
```

### Config File

The CLI reads from `~/.config/fastmail-cli/config.yaml`:

```yaml
endpoint: https://api.fastmail.com/jmap/session
carddav_endpoint: https://carddav.fastmail.com/dav/
carddav_username: username@fastmail.com
```

## Commands

| Command | Description |
|---------|-------------|
| `auth` | Manage authentication (login, logout, status) |
| `mail` | Email operations (list, send, reply) |
| `contacts` | Contact management (list, show, create, update, delete) |
| `masked-email` | Masked email addresses (list, create, enable, disable, delete) |
| `export` | Export emails (jsonl, maildir, mbox formats) |
| `mcp` | Run MCP server for AI assistant integration |

## Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Config file path |
| `--json` | Output as JSON |
| `--quiet` | Suppress output |
| `-v, --version` | Show version |

## MCP Integration

The MCP server exposes FastMail resources and tools to AI assistants:

```bash
fastmail-cli mcp
```

Configure in Claude Desktop `claude_desktop_config.json`:

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

## Documentation

Full documentation available at the [documentation site](docs/site/).

## License

MIT

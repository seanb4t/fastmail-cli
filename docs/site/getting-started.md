# Getting Started

This guide walks you through installing and configuring FastMail CLI.

## Prerequisites

- Go 1.21 or later (for building from source)
- A FastMail account with API access enabled

## Installation

### From Source

```bash
git clone https://github.com/user/fastmail-cli.git
cd fastmail-cli
go build -o fastmail-cli ./cmd/fastmail-cli
```

### Using Go Install

```bash
go install github.com/user/fastmail-cli/cmd/fastmail-cli@latest
```

## Configuration

### Getting an API Token

1. Log in to FastMail web interface
2. Navigate to **Settings** → **Privacy & Security** → **API Tokens**
3. Create a new token with the required scopes:
   - `Mail.ReadItems` - Read emails
   - `Contacts.Read` - Read contacts
   - `Calendars.Read` - Read calendar events

### Setting Up Credentials

Configure your API token:

```bash
fastmail-cli config set api-token YOUR_API_TOKEN
```

Or use environment variables:

```bash
export FASTMAIL_API_TOKEN=YOUR_API_TOKEN
```

## Verify Installation

Test that everything is working:

```bash
# Check version
fastmail-cli version

# List your mailboxes
fastmail-cli mail mailboxes
```

## Basic Usage

### Working with Mail

```bash
# List recent emails
fastmail-cli mail list --limit 10

# Search for emails
fastmail-cli mail search "from:important@example.com"

# Read a specific email
fastmail-cli mail show MESSAGE_ID
```

### Working with Contacts

```bash
# List contacts
fastmail-cli contacts list

# Show contact details
fastmail-cli contacts show CONTACT_ID
```

### Using the MCP Server

Start the MCP server for AI assistant integration:

```bash
fastmail-cli mcp serve
```

## Next Steps

- Explore the [CLI Reference](reference/cli.md) for complete command documentation
- Learn about [MCP Integration](reference/mcp.md) for AI assistant features

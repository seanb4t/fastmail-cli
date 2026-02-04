# Getting Started

This guide walks you through installing and configuring FastMail CLI.

## Prerequisites

- Go 1.21 or later (for building from source)
- A FastMail account with API access enabled

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

### Getting an API Token

1. Log in to FastMail web interface
2. Navigate to **Settings** → **Privacy & Security** → **API Tokens**
3. Create a new token with the required scopes:
   - `urn:ietf:params:jmap:core` - Core JMAP functionality
   - `urn:ietf:params:jmap:mail` - Read and send emails
   - `urn:ietf:params:jmap:submission` - Email submission
   - `https://www.fastmail.com/dev/maskedemail` - Masked emails

### Setting Up Credentials

Store your API token securely (uses system keychain on macOS):

```bash
fastmail-cli auth login
```

Or use an environment variable:

```bash
export FASTMAIL_TOKEN=your-api-token
```

## Verify Installation

Test that everything is working:

```bash
# Check version
fastmail-cli --version

# Check authentication status
fastmail-cli auth status

# List recent emails
fastmail-cli mail list
```

## Basic Usage

### Working with Mail

```bash
# List recent emails
fastmail-cli mail list

# List emails from Inbox
fastmail-cli mail list --folder Inbox

# Send an email
fastmail-cli mail send --to alice@example.com --subject "Hello" --body "Hi there!"

# Reply to an email
fastmail-cli mail reply MESSAGE_ID --body "Thanks for your message!"
```

### Working with Contacts

```bash
# List contacts
fastmail-cli contacts list

# Show contact details
fastmail-cli contacts show CONTACT_ID

# Create a contact
fastmail-cli contacts create --name "Alice Smith" --email alice@example.com
```

### Working with Masked Emails

```bash
# List masked emails
fastmail-cli masked-email list

# Create a masked email for a domain
fastmail-cli masked-email create --for-domain shopping.example.com

# Disable a masked email
fastmail-cli masked-email disable MASKED_EMAIL_ID
```

### Exporting Emails

```bash
# Export inbox to JSON Lines
fastmail-cli export --folder Inbox --format jsonl > inbox.jsonl

# Export to Maildir format
fastmail-cli export --folder Inbox --format maildir --output ~/backup/inbox

# Export to mbox format
fastmail-cli export --folder Archive --format mbox --output archive.mbox
```

### Using the MCP Server

Start the MCP server for AI assistant integration:

```bash
fastmail-cli mcp
```

Configure in Claude Desktop by adding to `claude_desktop_config.json`:

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

## Next Steps

- Explore the [CLI Reference](cli/index.md) for complete command documentation
- Learn about [MCP Integration](mcp/index.md) for AI assistant features

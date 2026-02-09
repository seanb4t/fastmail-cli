# Getting Started

This guide walks you through installing and configuring FastMail CLI.

## Prerequisites

- Go 1.21 or later (for building from source)
- A FastMail account with API access enabled

## Installation

### Binary Releases

Download pre-built binaries from [GitHub Releases](https://github.com/seanb4t/fastmail-cli/releases). Builds are available for Linux and macOS on amd64 and arm64.

### Using Go Install

```bash
go install github.com/seanb4t/fastmail-cli/cmd/fastmail-cli@latest
```

### From Source

```bash
git clone https://github.com/seanb4t/fastmail-cli.git
cd fastmail-cli
task build    # requires https://taskfile.dev
```

## Configuration

### Getting an API Token

1. Log in to [FastMail](https://www.fastmail.com) web interface
2. Navigate to **Settings** > **Privacy & Security** > **API Tokens**
3. Create a new token with the required scopes:
    - `urn:ietf:params:jmap:core` -- Core JMAP functionality
    - `urn:ietf:params:jmap:mail` -- Read and send emails
    - `urn:ietf:params:jmap:submission` -- Email submission
    - `https://www.fastmail.com/dev/maskedemail` -- Masked emails

### Setting Up Credentials

Store your API token securely (uses system keychain on macOS):

```bash
fastmail-cli auth login
```

Or provide the token non-interactively:

```bash
fastmail-cli auth login --token fmu1-xxxxxxxx
```

Or use an environment variable:

```bash
export FASTMAIL_TOKEN=your-api-token
```

### Config File

Located at `~/.config/fastmail-cli/config.yaml`:

```yaml
endpoint: https://api.fastmail.com/jmap/session
carddav_endpoint: https://carddav.fastmail.com/dav/
caldav_endpoint: https://caldav.fastmail.com/dav/
carddav_username: username@fastmail.com
```

Environment variables:

| Variable | Description |
|----------|-------------|
| `FASTMAIL_API_TOKEN` | API token (alternative to `auth login`) |
| `FASTMAIL_ENDPOINT` | JMAP endpoint URL |
| `FASTMAIL_CARDDAV_ENDPOINT` | CardDAV endpoint URL |
| `FASTMAIL_CARDDAV_USERNAME` | CardDAV username |

## Verify Installation

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

# List emails from a specific folder
fastmail-cli mail list --folder Sent --limit 20

# Show a single email
fastmail-cli mail show EMAIL_ID

# Search emails
fastmail-cli mail search "from:alice subject:meeting"

# Send an email
fastmail-cli mail send --to alice@example.com --subject "Hello" --body "Hi there!"

# Reply to an email
fastmail-cli mail reply EMAIL_ID --body "Thanks for your message!"

# Move an email to Archive
fastmail-cli mail move EMAIL_ID --folder Archive

# Flag an email as read and starred
fastmail-cli mail flag EMAIL_ID --read --star
```

### Working with Mailboxes

```bash
# List all mailboxes
fastmail-cli mailbox list

# Create a new mailbox
fastmail-cli mailbox create "Projects"

# Rename a mailbox
fastmail-cli mailbox rename MAILBOX_ID --name "Old Projects"

# Delete a mailbox
fastmail-cli mailbox delete MAILBOX_ID
```

### Working with Contacts

```bash
# List contacts
fastmail-cli contacts list

# Search contacts
fastmail-cli contacts list --search alice

# Show contact details
fastmail-cli contacts show CONTACT_ID

# Create a contact
fastmail-cli contacts create --name "Alice Smith" --email alice@example.com
```

### Working with Calendar

```bash
# List available calendars
fastmail-cli calendar calendars

# List today's events
fastmail-cli calendar list

# List events in a date range
fastmail-cli calendar list --start 2026-02-01 --end 2026-02-28

# Create an event
fastmail-cli calendar create --calendar CAL_ID --summary "Meeting" \
  --start 2026-02-10T14:00:00Z --end 2026-02-10T15:00:00Z
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

### Working with Vacation Responses

```bash
# Check vacation status
fastmail-cli vacation status

# Enable vacation response
fastmail-cli vacation enable --subject "Out of Office" --body "I'll be back Monday."

# Disable vacation response
fastmail-cli vacation disable
```

### Working with Filters

```bash
# List filter scripts
fastmail-cli filter list

# Show a filter script
fastmail-cli filter show SCRIPT_ID

# Create a filter from a file
fastmail-cli filter create --name "My Filter" --file filter.sieve

# Validate a script
fastmail-cli filter validate --file filter.sieve
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

## Shell Completion

```bash
# Bash
fastmail-cli completion bash > /etc/bash_completion.d/fastmail-cli

# Zsh
fastmail-cli completion zsh > "${fpath[1]}/_fastmail-cli"

# Fish
fastmail-cli completion fish > ~/.config/fish/completions/fastmail-cli.fish

# PowerShell
fastmail-cli completion powershell > fastmail-cli.ps1
```

## Next Steps

- [CLI Reference](cli/index.md) for complete command documentation
- [MCP Integration](mcp/index.md) for AI assistant features
- [Go Library](api/index.md) for programmatic access

# MCP Setup Guide

Configure FastMail CLI as an MCP server for Claude Desktop.

## Prerequisites

1. FastMail CLI installed and configured with your API token
2. Claude Desktop installed

## Claude Desktop Configuration

Add the FastMail MCP server to your Claude Desktop configuration.

### macOS

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "fastmail": {
      "command": "fastmail-cli",
      "args": ["mcp", "serve"]
    }
  }
}
```

### Windows

Edit `%APPDATA%\Claude\claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "fastmail": {
      "command": "fastmail-cli.exe",
      "args": ["mcp", "serve"]
    }
  }
}
```

### Linux

Edit `~/.config/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "fastmail": {
      "command": "fastmail-cli",
      "args": ["mcp", "serve"]
    }
  }
}
```

## Using Full Path

If `fastmail-cli` is not in your PATH, use the full path:

```json
{
  "mcpServers": {
    "fastmail": {
      "command": "/path/to/fastmail-cli",
      "args": ["mcp", "serve"]
    }
  }
}
```

## Environment Variables

Pass your API token via environment variable if not using the config file:

```json
{
  "mcpServers": {
    "fastmail": {
      "command": "fastmail-cli",
      "args": ["mcp", "serve"],
      "env": {
        "FASTMAIL_API_TOKEN": "your-api-token-here"
      }
    }
  }
}
```

## Verify Setup

1. Restart Claude Desktop after saving the configuration
2. Look for the MCP server icon in Claude's interface
3. Ask Claude: "What FastMail tools are available?"

Claude should list the available tools and resources.

## Troubleshooting

### Server Not Starting

Check that `fastmail-cli` is accessible:

```bash
fastmail-cli version
```

### Authentication Errors

Verify your API token is configured:

```bash
fastmail-cli config get api-token
```

Or set it:

```bash
fastmail-cli config set api-token YOUR_TOKEN
```

### Connection Issues

Test the MCP server manually:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | fastmail-cli mcp serve
```

You should see a JSON response with server capabilities.

### Logs

Check Claude Desktop logs for MCP-related errors:

- macOS: `~/Library/Logs/Claude/`
- Windows: `%LOCALAPPDATA%\Claude\logs\`
- Linux: `~/.local/share/Claude/logs/`

## Multiple MCP Servers

You can run FastMail alongside other MCP servers:

```json
{
  "mcpServers": {
    "fastmail": {
      "command": "fastmail-cli",
      "args": ["mcp", "serve"]
    },
    "filesystem": {
      "command": "mcp-filesystem",
      "args": ["/path/to/allowed/dir"]
    }
  }
}
```

## Security Best Practices

1. **Token Scope** - Create an API token with only the scopes you need
2. **Review Operations** - Check tool calls before Claude executes them
3. **Local Only** - The MCP server only runs locally via stdio
4. **Restart Required** - Restart Claude Desktop to apply config changes

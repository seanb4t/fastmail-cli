# auth

Manage authentication credentials for FastMail CLI.

## Synopsis

```bash
fastmail-cli auth <command> [options]
```

## Commands

### login

Store your FastMail API token for authentication.

```bash
fastmail-cli auth login [--token <token>]
```

**Options:**

| Flag | Description |
|------|-------------|
| `--token <token>` | API token (for non-interactive use) |

**Examples:**

```bash
# Interactive login (prompts for token)
fastmail-cli auth login

# Non-interactive login
fastmail-cli auth login --token "fmu1-xxxxxxxxxxxxxxxx"

# Login with token from environment
fastmail-cli auth login --token "$FASTMAIL_TOKEN"
```

**Getting an API Token:**

1. Go to [FastMail Settings](https://www.fastmail.com/settings/security/devicekeys)
2. Navigate to **Password & Security > App Passwords**
3. Create a new app password with JMAP access
4. Copy the token (starts with `fmu1-`)

---

### logout

Remove stored authentication credentials.

```bash
fastmail-cli auth logout
```

**Example:**

```bash
fastmail-cli auth logout
# Logged out successfully
```

---

### status

Check current authentication status.

```bash
fastmail-cli auth status [--json]
```

**Options:**

| Flag | Description |
|------|-------------|
| `--json` | Output status as JSON |

**Examples:**

```bash
# Check status
fastmail-cli auth status
# Logged in

# Check status as JSON
fastmail-cli auth status --json
# {
#   "logged_in": true
# }
```

**Exit Codes:**

| Code | Meaning |
|------|---------|
| 0 | Success (status retrieved) |
| 1 | Error reading authentication state |

!!! note "Status vs Validity"
    `auth status` checks if a token is stored locally. It does not verify the token is still valid with FastMail's servers. Use any API command to verify the token works.

## Token Storage

Tokens are stored in the configuration directory:

- **Location:** `~/.config/fastmail-cli/` (or custom `--config` path)
- **Format:** Encrypted file storage

!!! warning "Security"
    Keep your API token secure. Anyone with access to this token can read and send email on your behalf.

## See Also

- [Getting Started](../getting-started.md) - Initial setup guide
- [Configuration](index.md#configuration) - Environment variables and config file

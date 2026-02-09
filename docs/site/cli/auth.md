# auth

Manage FastMail authentication credentials.

## Subcommands

- [auth login](#auth-login) -- Store API token
- [auth logout](#auth-logout) -- Remove stored credentials
- [auth status](#auth-status) -- Show authentication status

---

## auth login

Store your FastMail API token for authentication. Uses system keychain on macOS with file fallback.

```bash
fastmail-cli auth login [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--token TOKEN` | API token (for non-interactive use) |

### Examples

```bash
# Interactive login (prompts for token)
fastmail-cli auth login

# Non-interactive login
fastmail-cli auth login --token fmu1-xxxxxxxx
```

### Getting an API Token

1. Log in to FastMail web interface
2. Go to **Settings** > **Privacy & Security** > **API tokens**
3. Create a new token with appropriate permissions
4. Copy the token and use it with `auth login`

---

## auth logout

Remove stored FastMail API token.

```bash
fastmail-cli auth logout
```

---

## auth status

Validate your authentication against the FastMail API. This checks not just whether a token is stored, but whether it is actually valid by making a request to the FastMail JMAP API.

```bash
fastmail-cli auth status
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Authenticated successfully |
| 1 | No token stored |
| 2 | Token expired or revoked |
| 3 | Cannot reach FastMail API |
| 4 | Authentication failed |

### Output

Text output:

```
Authenticated as user@fastmail.com
```

JSON output (`--json`):

```json
{
  "authenticated": true,
  "username": "user@fastmail.com"
}
```

### Examples

```bash
# Check status in scripts
if fastmail-cli auth status; then
  echo "Ready to use"
else
  echo "Please run: fastmail-cli auth login"
fi

# JSON output for automation
fastmail-cli auth status --json | jq -r '.username'
```

## See Also

- [CLI Reference](index.md)
- [mail](mail.md) -- requires authentication

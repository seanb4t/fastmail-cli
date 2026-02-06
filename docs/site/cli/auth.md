# auth

Manage FastMail authentication credentials.

## Commands

- [auth login](#auth-login) - Store API token
- [auth logout](#auth-logout) - Remove stored credentials
- [auth status](#auth-status) - Show authentication status

---

## auth login

Store your FastMail API token for authentication.

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

# Use with custom config location
fastmail-cli --config ~/.fastmail/config.yaml auth login
```

### Getting an API Token

1. Log in to FastMail web interface
2. Go to Settings → Privacy & Security → API tokens
3. Create a new token with appropriate permissions
4. Copy the token and use it with `auth login`

---

## auth logout

Remove stored FastMail API token.

```bash
fastmail-cli auth logout
```

### Examples

```bash
# Remove stored credentials
fastmail-cli auth logout

# Verify logged out
fastmail-cli auth status
```

---

## auth status

Validate your authentication against the FastMail API.

This command checks not just whether a token is stored, but whether it's actually valid by making a request to the FastMail JMAP API.

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
or
```
Not logged in
```
or
```
Token expired or revoked
```
or
```
Authentication failed
```
or
```
Cannot reach FastMail API: <error details>
```

JSON output (`--json`):
```json
{
  "authenticated": true,
  "username": "user@fastmail.com"
}
```

When not authenticated:
```json
{
  "authenticated": false,
  "reason": "no_token"
}
```

Authentication failures:
```json
{
  "authenticated": false,
  "reason": "auth_error"
}
```

### Examples

```bash
# Check authentication status
fastmail-cli auth status

# Check status in scripts
if fastmail-cli auth status; then
  echo "Ready to use"
else
  echo "Please run: fastmail-cli auth login"
fi

# JSON output for automation
fastmail-cli auth status --json

# Combine with other tools
fastmail-cli auth status --json | jq -r '.username'
```

## See Also

- [CLI Reference](index.md)
- [mail](mail.md) - requires authentication

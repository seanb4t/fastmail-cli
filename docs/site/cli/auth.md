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

Show whether you are currently logged in.

```bash
fastmail-cli auth status
```

### Output

Text output:
```
Logged in
```
or
```
Not logged in
```

JSON output (`--json`):
```json
{
  "logged_in": true
}
```

### Examples

```bash
# Check status
fastmail-cli auth status

# Check status as JSON
fastmail-cli auth status --json
```

## See Also

- [CLI Reference](index.md)
- [mail](mail.md) - requires authentication

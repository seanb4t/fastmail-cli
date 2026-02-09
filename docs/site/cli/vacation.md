# vacation

Commands for managing vacation/out-of-office auto-reply settings.

## Subcommands

- [vacation status](#vacation-status) -- Show vacation response status
- [vacation enable](#vacation-enable) -- Enable vacation response
- [vacation disable](#vacation-disable) -- Disable vacation response

---

## vacation status

Display the current vacation/out-of-office auto-reply settings.

```bash
fastmail-cli vacation status
```

### Output

Text output:

```
Vacation response: ENABLED
Subject: Out of Office
Message: I'll be back on the 15th.
From:    2026-02-10
To:      2026-02-14
```

or:

```
Vacation response: DISABLED
```

JSON output:

```json
{
  "is_enabled": true,
  "subject": "Out of Office",
  "text_body": "I'll be back on the 15th.",
  "from_date": "2026-02-10T00:00:00Z",
  "to_date": "2026-02-14T00:00:00Z"
}
```

---

## vacation enable

Enable the vacation/out-of-office auto-reply with a subject and message body.

```bash
fastmail-cli vacation enable [flags]
```

### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--subject` | Yes | Auto-reply subject line |
| `--body` | Yes | Auto-reply message body |
| `--from` | No | Start date (RFC3339 or YYYY-MM-DD) |
| `--to` | No | End date (RFC3339 or YYYY-MM-DD) |

### Examples

```bash
# Enable with date range
fastmail-cli vacation enable \
  --subject "Out of Office" \
  --body "I'll be back February 15th." \
  --from 2026-02-10 \
  --to 2026-02-14

# Enable without date range (always active)
fastmail-cli vacation enable \
  --subject "Away" \
  --body "I'm currently unavailable."
```

---

## vacation disable

Disable the vacation/out-of-office auto-reply.

```bash
fastmail-cli vacation disable
```

## See Also

- [CLI Reference](index.md)
- [identity](identity.md) -- manage sender identities

# masked-email

Manage masked email addresses.

## Synopsis

```bash
fastmail-cli masked-email <command> [options]
```

## Overview

Masked emails are unique, disposable email addresses that forward to your main inbox. They help protect your primary email address and let you easily disable or delete addresses if they receive spam.

## Commands

### list

List all masked email addresses.

```bash
fastmail-cli masked-email list
```

**Example:**

```bash
fastmail-cli masked-email list
```

**Output:**

```
ME001  abc123@fastmail.com  [enabled]   example.com
ME002  def456@fastmail.com  [disabled]  shopping-site.com
ME003  ghi789@fastmail.com  [pending]   newsletter.com
```

Status states:

- `[enabled]` - Active, receiving mail
- `[disabled]` - Blocked, mail rejected
- `[pending]` - Created but not yet used

---

### create

Create a new masked email address.

```bash
fastmail-cli masked-email create [--for-domain <domain>] [--description <text>]
```

**Options:**

| Flag | Description |
|------|-------------|
| `--for-domain <domain>` | Domain this masked email is for |
| `--description <text>` | Description/note for this address |

**Examples:**

```bash
# Create with no metadata
fastmail-cli masked-email create

# Create for a specific domain
fastmail-cli masked-email create --for-domain "shopping-site.com"

# Create with description
fastmail-cli masked-email create \
  --for-domain "newsletter.com" \
  --description "Tech newsletter signup"
```

**Output:**

```
Created: xyz123@fastmail.com
```

!!! tip "Domain Tracking"
    Using `--for-domain` helps you track which addresses go where and makes cleanup easier if a site starts sending spam.

---

### enable

Enable a disabled masked email address.

```bash
fastmail-cli masked-email enable <id>
```

**Arguments:**

| Argument | Description |
|----------|-------------|
| `<id>` | Masked email ID |

**Example:**

```bash
fastmail-cli masked-email enable ME002
# Enabled: ME002
```

---

### disable

Disable an enabled masked email address.

```bash
fastmail-cli masked-email disable <id>
```

**Arguments:**

| Argument | Description |
|----------|-------------|
| `<id>` | Masked email ID |

**Example:**

```bash
fastmail-cli masked-email disable ME001
# Disabled: ME001
```

When disabled:

- Incoming mail is rejected (bounced)
- Existing conversations are preserved in your inbox
- The address can be re-enabled later

---

### delete

Permanently delete a masked email address.

```bash
fastmail-cli masked-email delete <id>
```

**Arguments:**

| Argument | Description |
|----------|-------------|
| `<id>` | Masked email ID |

**Example:**

```bash
fastmail-cli masked-email delete ME003
# Deleted: ME003
```

!!! warning "Permanent Deletion"
    Once deleted, the masked email address cannot be recovered or recreated. Any mail sent to it will bounce permanently.

## JSON Output

With `--json`, output includes full details:

```json
[
  {
    "id": "ME001",
    "email": "abc123@fastmail.com",
    "state": "enabled",
    "forDomain": "example.com",
    "description": "Example site signup",
    "createdAt": "2024-01-15T10:30:00Z",
    "lastMessageAt": "2024-01-20T14:22:00Z"
  }
]
```

## Common Workflows

### Managing Spam Sources

```bash
# List all masked emails
fastmail-cli masked-email list

# Find the one receiving spam and disable it
fastmail-cli masked-email disable ME001

# Or delete it entirely
fastmail-cli masked-email delete ME001
```

### Bulk Creation Script

```bash
# Create masked emails for multiple sites
for site in site1.com site2.com site3.com; do
  fastmail-cli masked-email create --for-domain "$site"
done
```

### Audit Active Addresses

```bash
# List enabled addresses as JSON and process
fastmail-cli masked-email list --json | \
  jq '.[] | select(.state == "enabled") | .email'
```

## See Also

- [mail](mail.md) - Email operations
- [MCP Integration](../reference/mcp.md) - Masked email tools for AI agents

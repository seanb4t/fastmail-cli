# masked-email

Commands for managing FastMail masked email addresses.

Masked emails are unique, generated email addresses that forward to your main inbox. They help protect your real email address and let you easily disable unwanted mail sources.

## Commands

- [masked-email list](#masked-email-list) - List masked emails
- [masked-email create](#masked-email-create) - Create a masked email
- [masked-email enable](#masked-email-enable) - Enable a masked email
- [masked-email disable](#masked-email-disable) - Disable a masked email
- [masked-email delete](#masked-email-delete) - Delete a masked email

---

## masked-email list

List all masked email addresses.

```bash
fastmail-cli masked-email list
```

### Output

Text output shows ID, address, state, and domain:
```
ME001  shop.abcd1234@fastmail.com  [enabled]   shop.example.com
ME002  news.efgh5678@fastmail.com  [disabled]  newsletter.example.com
ME003  temp.ijkl9012@fastmail.com  [pending]   signup.example.com
```

JSON output includes all fields.

### Examples

```bash
# List all masked emails
fastmail-cli masked-email list

# List as JSON
fastmail-cli masked-email list --json

# Filter enabled emails (with jq)
fastmail-cli masked-email list --json | jq '.[] | select(.state == "enabled")'
```

---

## masked-email create

Create a new masked email address.

```bash
fastmail-cli masked-email create [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--for-domain DOMAIN` | Domain this masked email is for |
| `--description TEXT` | Description of the masked email |

### Output

Text output:
```
Created: shop.abcd1234@fastmail.com
```

JSON output includes full masked email details.

### Examples

```bash
# Create a basic masked email
fastmail-cli masked-email create

# Create for a specific domain
fastmail-cli masked-email create --for-domain shop.example.com

# Create with description
fastmail-cli masked-email create \
  --for-domain newsletter.example.com \
  --description "Weekly newsletter signup"

# Create and get full details
fastmail-cli masked-email create --for-domain test.com --json
```

---

## masked-email enable

Enable a disabled masked email address.

```bash
fastmail-cli masked-email enable ID
```

### Arguments

| Argument | Description |
|----------|-------------|
| `ID` | Masked email ID to enable |

### Output

Text output:
```
Enabled: ME001
```

### Examples

```bash
# Enable a masked email
fastmail-cli masked-email enable ME001

# Enable and verify
fastmail-cli masked-email enable ME001 && fastmail-cli masked-email list --json | jq '.[] | select(.id == "ME001")'
```

---

## masked-email disable

Disable an enabled masked email address.

```bash
fastmail-cli masked-email disable ID
```

### Arguments

| Argument | Description |
|----------|-------------|
| `ID` | Masked email ID to disable |

Disabled masked emails stop receiving new messages. Existing messages are preserved.

### Output

Text output:
```
Disabled: ME001
```

### Examples

```bash
# Disable a masked email
fastmail-cli masked-email disable ME001

# Disable emails from a spammy domain
for id in $(fastmail-cli masked-email list --json | jq -r '.[] | select(.forDomain == "spam.example.com") | .id'); do
  fastmail-cli masked-email disable "$id"
done
```

---

## masked-email delete

Permanently delete a masked email address.

```bash
fastmail-cli masked-email delete ID
```

### Arguments

| Argument | Description |
|----------|-------------|
| `ID` | Masked email ID to delete |

**Warning:** Deleted masked emails cannot be recovered. Any mail sent to the deleted address will bounce.

### Output

Text output:
```
Deleted: ME001
```

### Examples

```bash
# Delete a masked email
fastmail-cli masked-email delete ME001
```

## Common Workflows

### Create Masked Email for Signup

```bash
# Create a new masked email for a service
EMAIL=$(fastmail-cli masked-email create \
  --for-domain service.example.com \
  --description "Service account" \
  --json | jq -r '.email')

echo "Use this email to sign up: $EMAIL"
```

### Disable All Masked Emails for a Domain

```bash
# Disable all emails from a spammy domain
DOMAIN="spam.example.com"
fastmail-cli masked-email list --json | \
  jq -r ".[] | select(.forDomain == \"$DOMAIN\") | .id" | \
  while read -r id; do
    fastmail-cli masked-email disable "$id"
  done
```

### Audit Masked Emails

```bash
# List masked emails by state
echo "=== Enabled ==="
fastmail-cli masked-email list --json | jq -r '.[] | select(.state == "enabled") | "\(.email)\t\(.forDomain)"'

echo "=== Disabled ==="
fastmail-cli masked-email list --json | jq -r '.[] | select(.state == "disabled") | "\(.email)\t\(.forDomain)"'
```

## See Also

- [CLI Reference](index.md)
- [auth](auth.md) - authentication required

# masked-email

Commands for managing FastMail masked email addresses.

Masked emails are unique, generated email addresses that forward to your main inbox. They help protect your real email address and let you easily disable unwanted mail sources.

## Subcommands

- [masked-email list](#masked-email-list) -- List masked emails
- [masked-email create](#masked-email-create) -- Create a masked email
- [masked-email enable](#masked-email-enable) -- Enable a masked email
- [masked-email disable](#masked-email-disable) -- Disable a masked email
- [masked-email delete](#masked-email-delete) -- Delete a masked email

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

### Examples

```bash
fastmail-cli masked-email create
fastmail-cli masked-email create --for-domain shop.example.com
fastmail-cli masked-email create --for-domain newsletter.example.com --description "Weekly newsletter signup"
```

---

## masked-email enable

Enable a disabled masked email address.

```bash
fastmail-cli masked-email enable ID
```

---

## masked-email disable

Disable an enabled masked email address. Disabled masked emails stop receiving new messages. Existing messages are preserved.

```bash
fastmail-cli masked-email disable ID
```

---

## masked-email delete

Permanently delete a masked email address. Deleted masked emails cannot be recovered. Mail sent to the deleted address will bounce.

```bash
fastmail-cli masked-email delete ID
```

## Common Workflows

### Create Masked Email for Signup

```bash
EMAIL=$(fastmail-cli masked-email create \
  --for-domain service.example.com \
  --description "Service account" \
  --json | jq -r '.email')

echo "Use this email to sign up: $EMAIL"
```

### Disable All Masked Emails for a Domain

```bash
DOMAIN="spam.example.com"
fastmail-cli masked-email list --json | \
  jq -r ".[] | select(.forDomain == \"$DOMAIN\") | .id" | \
  while read -r id; do
    fastmail-cli masked-email disable "$id"
  done
```

## See Also

- [CLI Reference](index.md)
- [auth](auth.md) -- authentication required

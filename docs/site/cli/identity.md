# identity

Commands for managing sender identities (name, email, reply-to, signature).

## Subcommands

- [identity list](#identity-list) -- List all sender identities
- [identity set](#identity-set) -- Update a sender identity

---

## identity list

Display all configured sender identities for the account.

```bash
fastmail-cli identity list
```

### Output

Text output shows ID, name, email, and signature snippet:

```
ID001                user@fastmail.com     [sig: Best regards, ...]
ID002                work@company.com
```

JSON output includes all fields: `id`, `name`, `email`, `text_signature`, `html_signature`, `reply_to`, `bcc`, and `may_delete`.

---

## identity set

Update the name, reply-to address, or text signature of a sender identity.

```bash
fastmail-cli identity set ID [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--name` | Display name for the identity |
| `--reply-to` | Reply-to email address |
| `--signature` | Text signature |

At least one of `--name`, `--reply-to`, or `--signature` must be specified.

### Examples

```bash
# Update display name
fastmail-cli identity set ID001 --name "John Doe"

# Set reply-to address
fastmail-cli identity set ID001 --reply-to noreply@example.com

# Update signature
fastmail-cli identity set ID001 --signature "Best regards, John"

# Multiple changes at once
fastmail-cli identity set ID001 --name "J. Doe" --signature "Cheers, J"
```

## See Also

- [CLI Reference](index.md)
- [vacation](vacation.md) -- vacation auto-reply settings

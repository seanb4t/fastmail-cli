# contacts

Commands for managing FastMail contacts via CardDAV.

## Commands

- [contacts list](#contacts-list) - List contacts
- [contacts show](#contacts-show) - Show contact details
- [contacts create](#contacts-create) - Create a contact
- [contacts update](#contacts-update) - Update a contact
- [contacts delete](#contacts-delete) - Delete a contact

---

## contacts list

List all contacts or search by name/email.

```bash
fastmail-cli contacts list [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--search QUERY` | Search contacts by name or email |

### Output

Text output shows ID, name, and email:
```
C001  Alice Smith  alice@example.com
C002  Bob Jones    bob@example.com
```

JSON output includes all contact fields.

### Examples

```bash
# List all contacts
fastmail-cli contacts list

# Search for contacts
fastmail-cli contacts list --search alice

# List contacts as JSON
fastmail-cli contacts list --json
```

---

## contacts show

Display detailed information about a single contact.

```bash
fastmail-cli contacts show ID
```

### Arguments

| Argument | Description |
|----------|-------------|
| `ID` | Contact ID |

### Output

Text output:
```
ID:      C001
Name:    Alice Smith
Email:   alice@example.com
Phone:   +1-555-0100
Address: 123 Main St, City, ST 12345
```

JSON output includes all fields.

### Examples

```bash
# Show contact details
fastmail-cli contacts show C001

# Show as JSON
fastmail-cli contacts show C001 --json
```

---

## contacts create

Create a new contact in your address book.

```bash
fastmail-cli contacts create [flags]
```

### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--name` | Yes | Contact name |
| `--email` | No | Contact email address |
| `--phone` | No | Contact phone number |

### Output

Text output:
```
Created: Alice Smith (C001)
```

JSON output includes full contact details.

### Examples

```bash
# Create contact with name only
fastmail-cli contacts create --name "Alice Smith"

# Create contact with all fields
fastmail-cli contacts create \
  --name "Bob Jones" \
  --email bob@example.com \
  --phone "+1-555-0100"

# Create and get ID
fastmail-cli contacts create --name "Test Contact" --json
```

---

## contacts update

Update an existing contact's information.

```bash
fastmail-cli contacts update ID [flags]
```

### Arguments

| Argument | Description |
|----------|-------------|
| `ID` | Contact ID to update |

### Flags

| Flag | Description |
|------|-------------|
| `--name` | New contact name |
| `--email` | New contact email address |
| `--phone` | New contact phone number |

Only provided fields are updated; others remain unchanged.

### Output

Text output:
```
Updated: C001
```

### Examples

```bash
# Update email only
fastmail-cli contacts update C001 --email newemail@example.com

# Update multiple fields
fastmail-cli contacts update C001 \
  --name "Alice Johnson" \
  --phone "+1-555-0200"
```

---

## contacts delete

Permanently delete a contact from your address book.

```bash
fastmail-cli contacts delete ID [flags]
```

### Arguments

| Argument | Description |
|----------|-------------|
| `ID` | Contact ID to delete |

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation |

### Safety

Without `--force`, the command shows a confirmation message and exits:
```
Are you sure you want to delete contact C001? Use --force to confirm.
```

### Examples

```bash
# Delete with confirmation prompt
fastmail-cli contacts delete C001

# Force delete (no confirmation)
fastmail-cli contacts delete C001 --force
```

## Configuration

Contacts require CardDAV configuration:

```yaml
# config.yaml
carddav_endpoint: https://carddav.fastmail.com/dav/
carddav_username: username@fastmail.com
```

Or via environment:
```bash
export FASTMAIL_CARDDAV_USERNAME=username@fastmail.com
```

## Common Workflows

### Export Contacts

```bash
# Export all contacts to JSON
fastmail-cli contacts list --json > contacts-backup.json
```

### Bulk Search

```bash
# Find contacts from a company
fastmail-cli contacts list --search "@company.com" --json
```

## See Also

- [CLI Reference](index.md)
- [auth](auth.md) - authentication required

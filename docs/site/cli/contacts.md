# contacts

Manage contacts via CardDAV.

## Synopsis

```bash
fastmail-cli contacts <command> [options]
```

## Prerequisites

The contacts command requires CardDAV configuration:

```yaml
# ~/.config/fastmail-cli/config.yaml
carddav_endpoint: https://carddav.fastmail.com/dav/addressbooks/user
carddav_username: your-username@fastmail.com
```

Or via environment variables:

```bash
export FASTMAIL_CARDDAV_ENDPOINT="https://carddav.fastmail.com/dav/addressbooks/user"
export FASTMAIL_CARDDAV_USERNAME="your-username@fastmail.com"
```

## Commands

### list

List all contacts or search by name/email.

```bash
fastmail-cli contacts list [--search <query>]
```

**Options:**

| Flag | Description |
|------|-------------|
| `--search <query>` | Search contacts by name or email |

**Examples:**

```bash
# List all contacts
fastmail-cli contacts list

# Search for contacts
fastmail-cli contacts list --search "alice"

# Output as JSON
fastmail-cli contacts list --json
```

**Output:**

```
C001  Alice Smith  alice@example.com
C002  Bob Jones    bob@example.com
C003  Carol White  carol@example.com
```

---

### show

Display detailed information about a single contact.

```bash
fastmail-cli contacts show <id>
```

**Arguments:**

| Argument | Description |
|----------|-------------|
| `<id>` | Contact ID |

**Example:**

```bash
fastmail-cli contacts show C001
```

**Output:**

```
ID:      C001
Name:    Alice Smith
Email:   alice@example.com
Phone:   +1-555-0100
Address: 123 Main St, City, ST 12345
```

---

### create

Create a new contact in your address book.

```bash
fastmail-cli contacts create --name <name> [--email <email>] [--phone <phone>]
```

**Options:**

| Flag | Required | Description |
|------|----------|-------------|
| `--name <name>` | Yes | Contact name |
| `--email <email>` | No | Email address |
| `--phone <phone>` | No | Phone number |

**Examples:**

```bash
# Create with name only
fastmail-cli contacts create --name "New Contact"

# Create with all fields
fastmail-cli contacts create \
  --name "Alice Smith" \
  --email "alice@example.com" \
  --phone "+1-555-0100"
```

**Output:**

```
Created: Alice Smith (C004)
```

---

### update

Update an existing contact's information.

```bash
fastmail-cli contacts update <id> [--name <name>] [--email <email>] [--phone <phone>]
```

**Arguments:**

| Argument | Description |
|----------|-------------|
| `<id>` | Contact ID to update |

**Options:**

| Flag | Description |
|------|-------------|
| `--name <name>` | New contact name |
| `--email <email>` | New email address |
| `--phone <phone>` | New phone number |

Only provided fields are updated; others remain unchanged.

**Examples:**

```bash
# Update email only
fastmail-cli contacts update C001 --email "new-email@example.com"

# Update multiple fields
fastmail-cli contacts update C001 \
  --name "Alice B. Smith" \
  --phone "+1-555-0200"
```

**Output:**

```
Updated: C001
```

---

### delete

Permanently delete a contact from your address book.

```bash
fastmail-cli contacts delete <id> [--force]
```

**Arguments:**

| Argument | Description |
|----------|-------------|
| `<id>` | Contact ID to delete |

**Options:**

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation prompt |

**Examples:**

```bash
# Delete with confirmation prompt
fastmail-cli contacts delete C001
# Are you sure you want to delete contact C001? Use --force to confirm.

# Delete without confirmation
fastmail-cli contacts delete C001 --force
# Deleted: C001
```

!!! warning "Permanent Deletion"
    Deleted contacts cannot be recovered. Use `--force` carefully.

## JSON Output

With `--json`, contact output includes full details:

```json
[
  {
    "id": "C001",
    "name": "Alice Smith",
    "email": "alice@example.com",
    "phone": "+1-555-0100",
    "address": "123 Main St"
  }
]
```

## CardDAV Notes

- Contacts are stored using the CardDAV protocol (vCard format)
- Changes sync with FastMail web and mobile apps
- Contact IDs are server-assigned and persistent

## See Also

- [auth](auth.md) - Authentication setup
- [Configuration](index.md#configuration) - CardDAV settings

# contacts

Commands for managing FastMail contacts via CardDAV.

## Subcommands

- [contacts list](#contacts-list) -- List contacts
- [contacts show](#contacts-show) -- Show contact details
- [contacts create](#contacts-create) -- Create a contact
- [contacts update](#contacts-update) -- Update a contact
- [contacts delete](#contacts-delete) -- Delete a contact

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

### Examples

```bash
fastmail-cli contacts list
fastmail-cli contacts list --search alice
fastmail-cli contacts list --json
```

---

## contacts show

Display detailed information about a single contact.

```bash
fastmail-cli contacts show ID
```

### Output

```
ID:      C001
Name:    Alice Smith
Email:   alice@example.com
Phone:   +1-555-0100
Address: 123 Main St, City, ST 12345
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

### Examples

```bash
fastmail-cli contacts create --name "Alice Smith"
fastmail-cli contacts create --name "Bob Jones" --email bob@example.com --phone "+1-555-0100"
```

---

## contacts update

Update an existing contact's information. Only provided fields are changed.

```bash
fastmail-cli contacts update ID [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--name` | New contact name |
| `--email` | New contact email address |
| `--phone` | New contact phone number |

### Examples

```bash
fastmail-cli contacts update C001 --email newemail@example.com
fastmail-cli contacts update C001 --name "Alice Johnson" --phone "+1-555-0200"
```

---

## contacts delete

Permanently delete a contact from your address book.

```bash
fastmail-cli contacts delete ID [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation |

Without `--force`, the command shows a confirmation message and exits.

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

## See Also

- [CLI Reference](index.md)
- [calendar](calendar.md) -- also uses DAV protocol

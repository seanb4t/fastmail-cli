# mailbox

Commands for listing, creating, renaming, and deleting mailbox folders.

## Subcommands

- [mailbox list](#mailbox-list) -- List mailboxes
- [mailbox create](#mailbox-create) -- Create a mailbox
- [mailbox rename](#mailbox-rename) -- Rename a mailbox
- [mailbox delete](#mailbox-delete) -- Delete a mailbox

---

## mailbox list

List all mailbox folders with unread and total email counts.

```bash
fastmail-cli mailbox list
```

### Output

Text output shows ID, name, role, and unread/total counts:

```
MB001  Inbox     inbox    12/458
MB002  Sent      sent     0/234
MB003  Archive   archive  0/1500
```

---

## mailbox create

Create a new mailbox folder.

```bash
fastmail-cli mailbox create NAME [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--parent` | Parent mailbox ID for nested folders |

### Examples

```bash
fastmail-cli mailbox create "Projects"
fastmail-cli mailbox create "Subproject" --parent MB001
```

---

## mailbox rename

Rename an existing mailbox folder.

```bash
fastmail-cli mailbox rename ID [flags]
```

### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--name` | Yes | New mailbox name |

### Examples

```bash
fastmail-cli mailbox rename MB003 --name "Old Projects"
```

---

## mailbox delete

Delete a mailbox folder by its ID.

```bash
fastmail-cli mailbox delete ID
```

## See Also

- [CLI Reference](index.md)
- [mail](mail.md) -- email operations

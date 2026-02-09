# filter

Commands for managing server-side Sieve email filter scripts.

## Subcommands

- [filter list](#filter-list) -- List filter scripts
- [filter show](#filter-show) -- Show filter script details
- [filter create](#filter-create) -- Create a filter script
- [filter activate](#filter-activate) -- Activate a filter script
- [filter deactivate](#filter-deactivate) -- Deactivate a filter script
- [filter validate](#filter-validate) -- Validate a filter script
- [filter delete](#filter-delete) -- Delete a filter script

---

## filter list

List all Sieve filter scripts.

```bash
fastmail-cli filter list
```

### Output

Text output shows ID, name, active indicator, and status:

```
FS001  My Filter    [*]  active
FS002  Old Filter   [ ]  inactive
```

---

## filter show

Show a Sieve filter script with its content.

```bash
fastmail-cli filter show ID
```

### Output

```
ID:       FS001
Name:     My Filter
Active:   active
Blob ID:  Bxyz789

--- Script ---
require ["fileinto"];
if address :is "from" "spam@example.com" {
  fileinto "Junk";
}
```

---

## filter create

Create a new Sieve filter script from a file or stdin.

```bash
fastmail-cli filter create [flags]
```

### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--name` | Yes | Filter script name |
| `--file` | No | Path to sieve script file (reads stdin if omitted) |
| `--activate` | No | Activate the script on creation |

### Examples

```bash
# Create from file
fastmail-cli filter create --name "Spam Filter" --file spam.sieve

# Create and activate
fastmail-cli filter create --name "Main Filter" --file rules.sieve --activate

# Create from stdin
cat filter.sieve | fastmail-cli filter create --name "Piped Filter"
```

---

## filter activate

Activate a Sieve filter script. Only one script can be active at a time.

```bash
fastmail-cli filter activate ID
```

---

## filter deactivate

Deactivate an active Sieve filter script.

```bash
fastmail-cli filter deactivate ID
```

---

## filter validate

Validate a Sieve filter script syntax without storing it.

```bash
fastmail-cli filter validate [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--file` | Path to sieve script file (reads stdin if omitted) |

### Output

```
Script is valid
```

or:

```
Script is invalid: line 3: unknown command "fileint"
```

### Examples

```bash
fastmail-cli filter validate --file rules.sieve
cat rules.sieve | fastmail-cli filter validate
```

---

## filter delete

Permanently delete a Sieve filter script.

```bash
fastmail-cli filter delete ID
```

## See Also

- [CLI Reference](index.md)
- [mail](mail.md) -- email operations

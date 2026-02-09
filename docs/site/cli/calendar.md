# calendar

Commands for managing FastMail calendar events via CalDAV.

## Subcommands

- [calendar calendars](#calendar-calendars) -- List all calendars
- [calendar list](#calendar-list) -- List events
- [calendar show](#calendar-show) -- Show event details
- [calendar create](#calendar-create) -- Create an event
- [calendar update](#calendar-update) -- Update an event
- [calendar delete](#calendar-delete) -- Delete an event

---

## calendar calendars

List all available calendars.

```bash
fastmail-cli calendar calendars
```

### Output

Text output shows ID, name, description, and read-only status:

```
CAL001  Personal Calendar
CAL002  Work  (Team calendar)  [read-only]
```

---

## calendar list

List calendar events within a date range. Defaults to today.

```bash
fastmail-cli calendar list [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--start` | Today 00:00 | Start date (RFC3339 or YYYY-MM-DD) |
| `--end` | Today 23:59 | End date (RFC3339 or YYYY-MM-DD) |
| `--calendar` | First calendar | Calendar ID |

### Examples

```bash
# Today's events
fastmail-cli calendar list

# Events this week
fastmail-cli calendar list --start 2026-02-09 --end 2026-02-15

# Events from a specific calendar
fastmail-cli calendar list --calendar CAL002
```

---

## calendar show

Display detailed information about a single calendar event.

```bash
fastmail-cli calendar show ID
```

### Output

```
ID:          EVT001
Summary:     Team Meeting
Calendar:    CAL001
Start:       2026-02-10T14:00:00Z
End:         2026-02-10T15:00:00Z
Location:    Conference Room A
Description: Weekly sync
Status:      CONFIRMED
```

---

## calendar create

Create a new calendar event.

```bash
fastmail-cli calendar create [flags]
```

### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--summary` | Yes | Event title/summary |
| `--start` | Yes | Start time (RFC3339 or YYYY-MM-DD) |
| `--end` | Yes | End time (RFC3339 or YYYY-MM-DD) |
| `--calendar` | Yes | Calendar ID |
| `--description` | No | Event description |
| `--location` | No | Event location |
| `--all-day` | No | Create as all-day event |

### Examples

```bash
fastmail-cli calendar create \
  --calendar CAL001 \
  --summary "Team Meeting" \
  --start 2026-02-10T14:00:00Z \
  --end 2026-02-10T15:00:00Z \
  --location "Zoom"

# All-day event
fastmail-cli calendar create \
  --calendar CAL001 \
  --summary "Company Holiday" \
  --start 2026-03-01 \
  --end 2026-03-02 \
  --all-day
```

---

## calendar update

Update an existing calendar event. Only provided fields are changed.

```bash
fastmail-cli calendar update ID [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--summary` | New event title |
| `--description` | New description |
| `--location` | New location |
| `--start` | New start time (RFC3339 or YYYY-MM-DD) |
| `--end` | New end time (RFC3339 or YYYY-MM-DD) |

### Examples

```bash
fastmail-cli calendar update EVT001 --location "Conference Room B"
fastmail-cli calendar update EVT001 --start 2026-02-10T15:00:00Z --end 2026-02-10T16:00:00Z
```

---

## calendar delete

Permanently delete a calendar event.

```bash
fastmail-cli calendar delete ID [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation |

Without `--force`, the command shows a confirmation message and exits.

## Configuration

Calendar operations require CalDAV configuration:

```yaml
# config.yaml
caldav_endpoint: https://caldav.fastmail.com/dav/
carddav_username: username@fastmail.com
```

## See Also

- [CLI Reference](index.md)
- [contacts](contacts.md) -- also uses DAV protocol

# internal/dav

CalDAV protocol client implementation.

## Purpose

Handle CalDAV/WebDAV communication with Fastmail calendars:
- Calendar collection discovery
- Event CRUD operations
- iCalendar parsing and serialization

## Key Types

| Type | Description |
|------|-------------|
| `Client` | CalDAV client interface |
| `Calendar` | Calendar collection type |
| `Event` | Calendar event with iCalendar properties |

## Functions

| Function | Description |
|----------|-------------|
| `ParseICalendarEvent` | Parse iCalendar data into Event |
| `SerializeEventToICalendar` | Convert Event to iCalendar format |

## Conventions

- Use context for request cancellation
- Return domain types, not protocol types
- Handle iCalendar required properties (DTSTAMP, UID)
- Support all-day and recurring events

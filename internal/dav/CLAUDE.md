# internal/dav

CardDAV and CalDAV protocol client implementations.

## Purpose

Handle WebDAV-based protocols for Fastmail:
- **CardDAV** for contact/address book operations
- **CalDAV** for calendar/event operations

## Key Types

| Type | File | Description |
|------|------|-------------|
| `CardDAVClient` | carddav.go | Contact operations via CardDAV |
| `CalDAVClient` | caldav.go | Calendar operations via CalDAV |
| `Contact` | carddav.go | Contact with vCard properties |
| `Calendar` | caldav.go | Calendar collection |
| `Event` | caldav.go | Calendar event with iCalendar properties |

## CardDAV Operations

| Function | Description |
|----------|-------------|
| `ListAddressBooks` | Discover address book collections |
| `ListContacts` | Get contacts from address book |
| `GetContact` | Retrieve single contact by path |
| `CreateContact` | Create new contact |
| `UpdateContact` | Update existing contact |
| `DeleteContact` | Remove contact |

## CalDAV Operations

| Function | Description |
|----------|-------------|
| `ListCalendars` | Discover calendar collections |
| `ListEvents` | Get events from calendar |
| `GetEvent` | Retrieve single event by path |
| `CreateEvent` | Create new event |
| `UpdateEvent` | Update existing event |
| `DeleteEvent` | Remove event |

## Conventions

- Use context for request cancellation
- Return domain types, not protocol types
- Handle vCard/iCalendar parsing internally
- Support required properties (UID, DTSTAMP)
- HTTPS required for all WebDAV operations

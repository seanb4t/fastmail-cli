# internal/dav

CardDAV and CalDAV client implementation for Fastmail.

## Purpose

Handle WebDAV protocol communication with Fastmail:
- CardDAV for contact management
- CalDAV for calendar management
- Discovery of address books and calendars

## Key Types

| Type | Description |
|------|-------------|
| `CardDAVClient` | CardDAV client for contacts |
| `CalDAVClient` | CalDAV client for calendars |

## Fastmail Endpoints

| Protocol | Endpoint |
|----------|----------|
| CardDAV | `https://carddav.fastmail.com/` |
| CalDAV | `https://caldav.fastmail.com/` |

## Authentication

Uses HTTP Basic Auth with:
- Username: Fastmail email address
- Password: API token with contacts/calendar access

## Conventions

- Use context for request cancellation
- Discover home sets before listing resources
- Access underlying client for advanced operations

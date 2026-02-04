# pkg/fastmail

High-level Fastmail client library.

## Purpose

Provide a clean Go API for Fastmail operations:
- Email operations (list, read, send)
- Contact management
- Calendar access

## Key Types

| Type | Description |
|------|-------------|
| `Client` | High-level Fastmail client |
| `Email` | Email message type |
| `Mailbox` | Mailbox/folder type |

## Conventions

- Hide JMAP complexity from consumers
- Return domain types, not protocol types
- Provide iterator patterns for large results
- Context for cancellation and timeout

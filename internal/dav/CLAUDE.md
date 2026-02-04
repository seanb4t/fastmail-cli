# internal/dav

WebDAV/CardDAV client implementation for contact management.

## Purpose

Provides low-level CardDAV protocol operations for contact synchronization with Fastmail.

## Key Types

| Type | Description |
|------|-------------|
| `CardDAVClient` | HTTP client for CardDAV operations |
| `Contact` | vCard contact data with all fields |
| `Address` | Physical address component |
| `AddressBook` | Address book collection |

## CardDAV Operations

| Method | Description |
|--------|-------------|
| `ListAddressBooks()` | Discover address books via PROPFIND |
| `ListContacts()` | Query all contacts in address book via REPORT |
| `GetContact()` | Retrieve single contact vCard via GET |
| `CreateContact()` | Create new contact via PUT with If-None-Match |
| `UpdateContact()` | Update contact via PUT with ETag precondition |
| `DeleteContact()` | Remove contact via DELETE |

## vCard Handling

| Function | Description |
|----------|-------------|
| `ParseVCard()` | Parse vCard 3.0 string to Contact |
| `Contact.ToVCard()` | Serialize Contact to vCard 3.0 |

## Fastmail Specifics

- CardDAV endpoint: `https://carddav.fastmail.com/`
- Address book path: `/dav/addressbooks/user/{username}/Default/`
- Uses app password authentication via Basic Auth
- vCard 3.0 format for contacts

## Dependencies

- `github.com/emersion/go-vcard` - vCard parsing/serialization

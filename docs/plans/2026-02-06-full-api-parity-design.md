# Full API Parity Design

**Goal:** Bring the CLI and MCP server to full CRUD parity with the Fastmail JMAP, CardDAV, and CalDAV APIs.

**Scope:** Every read and write operation the API supports gets both a CLI command and an MCP tool, developed in lock-step.

**Decisions:**
- Calendar uses a top-level `calendar` command (mirrors `contacts` pattern)
- `mail show` uses configurable display: minimal default, `--full`, `--headers`, `--raw`
- `mail search` uses query string syntax matching Fastmail web UI conventions
- CLI commands and MCP tools are added together (lock-step parity)

---

## Command Surface

```
fastmail-cli
├── auth
│   ├── login                      (exists)
│   ├── logout                     (exists)
│   └── status                     (exists)
├── mail
│   ├── list                       (exists)
│   ├── show <ID>                  NEW
│   ├── search <query>             NEW
│   ├── move <ID>                  NEW
│   ├── delete <ID>                NEW
│   ├── flag <ID>                  NEW
│   ├── send                       (exists)
│   └── reply <ID>                 (exists)
├── mailbox
│   ├── list                       NEW
│   ├── create                     NEW
│   ├── rename <ID>                NEW
│   └── delete <ID>                NEW
├── contacts
│   ├── list                       (exists)
│   ├── show <ID>                  (exists)
│   ├── create                     (exists)
│   ├── update <ID>                (exists)
│   └── delete <ID>                (exists)
├── masked-email
│   ├── list                       (exists)
│   ├── create                     (exists)
│   ├── enable <ID>                (exists)
│   ├── disable <ID>               (exists)
│   └── delete <ID>                (exists)
├── calendar
│   ├── list                       NEW — list events (default: upcoming 7 days)
│   ├── show <ID>                  NEW
│   ├── create                     NEW
│   ├── update <ID>                NEW
│   ├── delete <ID>                NEW
│   └── calendars                  NEW — list calendar containers
├── identity
│   └── list                       NEW — read-only
├── vacation
│   ├── show                       NEW
│   └── set                        NEW
├── thread
│   └── show <ID>                  NEW
├── export                         (exists)
└── mcp                            (exists, tools updated in lock-step)
```

19 new CLI commands across 5 new command groups plus 5 new subcommands under `mail`.

---

## Mail Commands

### `mail show <ID>`

Displays email content with progressive disclosure.

**Flags:**
- `--full` — All headers and metadata
- `--headers` — Raw RFC 5322 headers only
- `--raw` — Raw RFC 5322 source (for piping)

**Default text output:**
```
From:    Alice <alice@example.com>
To:      Bob <bob@example.com>
Date:    2026-02-05 14:30 UTC
Subject: Meeting tomorrow

Hey Bob, are we still on for tomorrow at 3pm?

Attachments:
  1. agenda.pdf (42 KB) [blob:Bf3a9c...]
```

**JSON output:** Full Email object from Email/get (all properties).

**`--raw`:** Fetches blob via `downloadUrl`, writes raw bytes to stdout.

**Service layer:** `MailService.Get()` exists. Add `MailService.GetRaw(ctx, id) (io.Reader, error)` for blob download. JMAP client needs `DownloadBlob(ctx, blobId) (io.Reader, error)`.

### `mail search <query>`

Query string syntax matching Fastmail web UI conventions.

**Flags:**
- `-n, --limit uint` — Maximum results (default: 10)

**Query syntax:**

| Token | JMAP Filter | Example |
|---|---|---|
| `from:<value>` | `from` | `from:alice` |
| `to:<value>` | `to` | `to:bob` |
| `subject:<value>` | `subject` | `subject:meeting` |
| `has:attachment` | `hasAttachment: true` | `has:attachment` |
| `before:<date>` | `before` | `before:2026-02-01` |
| `after:<date>` | `after` | `after:2026-01-01` |
| `is:unread` | `hasKeyword: $seen` (negated) | `is:unread` |
| `is:flagged` | `hasKeyword: $flagged` | `is:flagged` |
| `in:<mailbox>` | `inMailbox` | `in:drafts` |
| free text | `text` | `quarterly report` |

Parser translates tokens into JMAP Email/query FilterCondition objects. Compound queries combine with AND. Uses existing `MailService.Search(ctx, query, limit)`.

### `mail move <ID>`

**Flags:**
- `-f, --folder string` — Target mailbox name (required)

Wraps existing `MailService.Move()`. Outputs `"Moved <ID> to <folder>"`.

### `mail delete <ID>`

**Flags:**
- `-f, --force` — Skip confirmation

Wraps existing `MailService.Delete()`. Without `--force`, prints confirmation prompt.

### `mail flag <ID>`

**Flags:**
- `--read` / `--unread` — Set/clear `$seen` keyword
- `--flagged` / `--unflagged` — Set/clear `$flagged` keyword

At least one flag pair required.

**Service layer:** New `MailService.SetKeywords(ctx, id, keywords map[string]bool)`. Maps to `Email/set` with keyword patch.

---

## Mailbox Commands

### `mailbox list`

**Flags:**
- `--tree` — Show as nested tree using parentId (default: flat list)

**Text output:**
```
ID             Name          Role     Unread  Total
Mab1234        Inbox         inbox    12      847
Mab5678        Drafts        drafts   0       3
Mab9abc        Sent          sent     0       1203
```

**JSON output:** Full Mailbox objects (id, name, role, parentId, totalEmails, unreadEmails, totalThreads, unreadThreads, sortOrder, myRights).

**Service layer:** New `MailboxService.List(ctx) ([]Mailbox, error)`. Maps to `Mailbox/get` with `ids: null`.

### `mailbox create`

**Flags:**
- `--name string` — Folder name (required)
- `--parent string` — Parent mailbox ID (optional)

Maps to `Mailbox/set` create.

### `mailbox rename <ID>`

**Flags:**
- `--name string` — New folder name (required)

Maps to `Mailbox/set` update.

### `mailbox delete <ID>`

**Flags:**
- `-f, --force` — Skip confirmation

Refuses to delete well-known roles (inbox, drafts, sent, trash, junk) even with `--force`. Maps to `Mailbox/set` destroy.

---

## Calendar Commands

### `calendar list`

**Flags:**
- `--from string` — Start date (default: today, format: 2006-01-02)
- `--to string` — End date (default: from + 7 days)
- `--calendar string` — Filter by calendar name/ID

**Text output:**
```
ID       Date        Time        Summary              Calendar
Ev123    2026-02-06  09:00-10:00 Team standup          Work
Ev456    2026-02-06  All day     Mom's birthday        Personal
```

Uses existing `CalendarService.ListEvents()`.

### `calendar show <ID>`

Full event detail: summary, description, location, start/end, calendar, status.

### `calendar create`

**Flags:**
- `--summary string` — Event title (required)
- `--start string` — Start datetime (required, RFC3339 or "2006-01-02 15:04")
- `--end string` — End datetime (required)
- `--location string` — Location (optional)
- `--description string` — Description (optional)
- `--calendar string` — Target calendar name/ID (optional, uses default)
- `--all-day` — Create all-day event

Uses existing `CalendarService.CreateEvent()`.

### `calendar update <ID>`

Same flags as create, all optional. Fetches existing event, applies partial update.

### `calendar delete <ID>`

**Flags:**
- `-f, --force` — Skip confirmation

Uses existing `CalendarService.DeleteEvent()`.

### `calendar calendars`

Lists calendar containers.

```
ID       Name          Color    Default  ReadOnly
Cal1     Personal      #3B82F6  yes      no
Cal2     Work          #EF4444  no       no
```

Uses existing `CalendarService.ListCalendars()`.

---

## Identity, Vacation, Thread Commands

### `identity list`

Read-only. Lists sender identities from JMAP Identity/get.

```
ID       Name           Email                    May Delete
Id1      Sean B         sean@fastmail.com         no
Id2      Work           sean@company.com          yes
```

**JSON output:** Full Identity objects (id, name, email, replyTo, bcc, textSignature, htmlSignature, mayDelete).

No create/update/delete — identity management requires email verification via web UI.

### `vacation show`

Read-only view of auto-reply settings from JMAP VacationResponse/get (singleton).

```
Status:    Enabled
From:      2026-02-10
To:        2026-02-17
Subject:   Out of office
Body:      I'm away until Feb 17...
```

### `vacation set`

**Flags:**
- `--enable` / `--disable` — Toggle auto-reply
- `--subject string` — Response subject
- `--body string` — Response body (plain text)
- `--from-date string` — Start date
- `--to-date string` — End date

At least one flag required. Maps to VacationResponse/set on the singleton.

### `thread show <ID>`

Lists all emails in a conversation thread, ordered oldest-first.

**Flags:**
- `-n, --limit uint` — Max emails to show (default: all)

**Text output:**
```
Thread T12345 (4 emails)

1. From: Alice <alice@ex.com>  2026-02-04 09:00
   Subject: Design review
   Preview: Hey team, let's review the new...

2. From: Bob <bob@ex.com>  2026-02-04 09:30
   Subject: Re: Design review
   Preview: Looks good to me. One question...
```

Two JMAP calls: Thread/get for emailIds, then Email/get for summaries.

---

## MCP Tool Parity

| CLI Command | MCP Tool | MCP Resource | Status |
|---|---|---|---|
| `mail show` | `mail_get` | `fastmail://mail/{id}` | exists |
| `mail search` | `mail_search` | — | exists |
| `mail move` | `mail_move` | — | exists |
| `mail delete` | `mail_delete` | — | exists |
| `mail flag` | `mail_flag` | — | **new** |
| `mailbox list` | `mailbox_list` | `fastmail://mailboxes` | **new** |
| `mailbox create` | `mailbox_create` | — | **new** |
| `mailbox rename` | `mailbox_rename` | — | **new** |
| `mailbox delete` | `mailbox_delete` | — | **new** |
| `calendar list` | `calendar_events` | `fastmail://calendar/today` | exists |
| `calendar show` | `calendar_get` | — | **new** |
| `calendar create` | `calendar_create` | — | exists |
| `calendar update` | `calendar_update` | — | **new** |
| `calendar delete` | `calendar_delete` | — | **new** |
| `calendar calendars` | `calendar_list` | `fastmail://calendars` | **new** |
| `contacts update` | `contacts_update` | — | **new** |
| `contacts delete` | `contacts_delete` | — | **new** |
| `identity list` | `identity_list` | `fastmail://identities` | **new** |
| `vacation show` | `vacation_get` | `fastmail://vacation` | **new** |
| `vacation set` | `vacation_set` | — | **new** |
| `thread show` | `thread_get` | — | **new** |

13 new MCP tools, 4 new MCP resources.

---

## Service Layer Changes

### New Services

| Service | Package | Protocol | Methods |
|---|---|---|---|
| `MailboxService` | `pkg/fastmail/` | JMAP | `List`, `Create`, `Rename`, `Delete` |
| `IdentityService` | `pkg/fastmail/` | JMAP | `List` |
| `VacationService` | `pkg/fastmail/` | JMAP | `Get`, `Set` |
| `ThreadService` | `pkg/fastmail/` | JMAP | `Get` |

### Additions to Existing Services

| Service | New Methods |
|---|---|
| `MailService` | `GetRaw(ctx, id) (io.Reader, error)`, `SetKeywords(ctx, id, keywords)` |
| `Client` | `Mailbox()`, `Identity()`, `Vacation()`, `Thread()` accessors |

### JMAP Client Additions

| Method | JMAP Call |
|---|---|
| `DownloadBlob(ctx, blobId)` | HTTP GET on `downloadUrl` template |
| `MailboxGet(ctx, ids)` | `Mailbox/get` |
| `MailboxSet(ctx, ...)` | `Mailbox/set` |
| `IdentityGet(ctx)` | `Identity/get` |
| `VacationGet(ctx)` | `VacationResponse/get` |
| `VacationSet(ctx, ...)` | `VacationResponse/set` |
| `ThreadGet(ctx, ids)` | `Thread/get` |
| `EmailSet` for keywords | `Email/set` with keyword patches |

---

## Implementation Order

### Phase 1 — Mailbox + Mail Read (highest value)

1. `internal/jmap/mailbox.go` — Mailbox types, get/set builders
2. `pkg/fastmail/mailbox_service.go` — MailboxService
3. `cli/mailbox.go` + `mcp/mailbox_tools.go` — CLI + MCP
4. `mail show` — GetRaw, blob download, display formatting
5. `mail search` — CLI wrapper around existing service method

### Phase 2 — Mail Write Operations

6. `mail move` — CLI wrapper (service exists)
7. `mail delete` — CLI wrapper (service exists)
8. `mail flag` — SetKeywords in service + JMAP, CLI + MCP

### Phase 3 — Calendar CLI

9. `cli/calendar.go` — All 6 calendar subcommands
10. `mcp/calendar_tools.go` — New + updated MCP tools

### Phase 4 — Identity, Vacation, Thread

11. Identity service + CLI + MCP
12. Vacation service + CLI + MCP
13. Thread service + CLI + MCP

### Phase 5 — MCP Catch-up

14. `contacts_update`, `contacts_delete` MCP tools
15. New MCP resources

---

## Testing Strategy

Same patterns as existing codebase:
- Table-driven tests with `testify/assert` + `testify/require`
- `httptest.NewServer` for JMAP mocks (canned JSON responses)
- Each CLI command: success case, error case, JSON output case
- Each MCP tool: handler test with mock fastmail client
- Service methods: unit tests with mock JMAP responses

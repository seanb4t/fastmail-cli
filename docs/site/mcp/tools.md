# MCP Tools Reference

Tools allow Claude to perform actions on your FastMail account.

## Mail Tools

### mail_list

List emails from a mailbox folder.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `folder` | string | No | Mailbox folder name (default: "Inbox") |
| `limit` | integer | No | Maximum emails to return (default: 10) |

---

### mail_get

Get a single email by its ID.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The email ID |

---

### mail_search

Search emails by text query and/or structured filters. Query syntax: `from:alice subject:meeting has:attachment before:2024-06-01 is:unread`. Flags are ANDed with the query.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `query` | string | No | Search query text with field:value filters |
| `from` | string | No | Filter by sender address or name |
| `to` | string | No | Filter by recipient address or name |
| `subject` | string | No | Filter by subject text |
| `before` | string | No | Emails before date (YYYY-MM-DD) |
| `after` | string | No | Emails after date (YYYY-MM-DD) |
| `has_attachment` | boolean | No | Filter for emails with attachments |
| `folder` | string | No | Filter by mailbox folder name |
| `unread` | boolean | No | Filter for unread emails |
| `flagged` | boolean | No | Filter for flagged/starred emails |
| `limit` | integer | No | Maximum results (default: 10) |
| `snippets` | boolean | No | Include highlighted search snippets (default: false) |

At least one query or filter parameter is required.

---

### mail_send

Compose and send a new email. Use `schedule` for delayed delivery.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `to` | array | Yes | Recipient email addresses |
| `cc` | array | No | CC recipient addresses |
| `bcc` | array | No | BCC recipient addresses |
| `subject` | string | Yes | Email subject |
| `body` | string | Yes | Email body text |
| `schedule` | string | No | Schedule delivery time in RFC3339 format |

---

### mail_reply

Send a reply to an existing email.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `email_id` | string | Yes | ID of the email to reply to |
| `body` | string | Yes | Reply body text |
| `reply_all` | boolean | No | Reply to all recipients (default: false) |

---

### mail_move

Move an email to a different folder.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The email ID |
| `folder` | string | Yes | Target folder name |

---

### mail_delete

Delete an email (moves to Trash, or permanently deletes if already in Trash).

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The email ID |

---

### mail_thread

Get all emails in a conversation thread.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `thread_id` | string | Yes | The thread ID |

---

### mail_flag

Set or remove flags/keywords on an email.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The email ID |
| `keywords` | object | Yes | Keys are keyword names (e.g., `$seen`, `$flagged`), values are `true` (set) or `false` (remove) |

---

### mail_attachments

List attachments on an email.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The email ID |

Returns: array of `{ blob_id, name, type, size, disposition }`.

---

### mail_download

Download an email attachment blob.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The email ID |
| `blob_id` | string | Yes | The blob ID of the attachment |
| `name` | string | No | Filename hint for the download |

Returns base64-encoded content.

---

### mail_import

Import an RFC 5322 email message into a mailbox.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `content` | string | Yes | Base64-encoded RFC 5322 message content |
| `folder` | string | No | Target mailbox folder (default: Inbox) |
| `seen` | boolean | No | Mark as read (default: false) |
| `flagged` | boolean | No | Mark as flagged (default: false) |

---

### mail_upload

Upload a blob for use in email drafts.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `content` | string | Yes | Base64-encoded file content |
| `content_type` | string | Yes | MIME content type (e.g., `application/pdf`) |
| `filename` | string | Yes | Filename for the blob |

---

### mail_scheduled

List pending scheduled email deliveries.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `limit` | integer | No | Maximum results (default: 10) |

---

### mail_scheduled_cancel

Cancel a pending scheduled email delivery.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `submission_id` | string | Yes | The submission ID to cancel |

---

## Mailbox Tools

### mailbox_list

List all mailbox folders with unread/total counts.

No input parameters.

---

### mailbox_create

Create a new mailbox folder.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Name for the new mailbox |
| `parent_id` | string | No | Parent mailbox ID for nested folders |

---

### mailbox_rename

Rename an existing mailbox folder.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The mailbox ID |
| `name` | string | Yes | New name for the mailbox |

---

### mailbox_delete

Delete a mailbox folder by ID.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The mailbox ID |

---

## Masked Email Tools

### masked_email_list

List all masked email addresses.

No input parameters.

---

### masked_email_get

Get a single masked email by ID.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The masked email ID |

---

### masked_email_create

Create a new masked email address.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `domain` | string | No | Domain to associate with the masked email |
| `description` | string | No | Description for the masked email |

---

### masked_email_enable

Enable a masked email by ID.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The masked email ID |

---

### masked_email_disable

Disable a masked email by ID.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The masked email ID |

---

### masked_email_delete

Delete a masked email by ID.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The masked email ID |

---

## Contact Tools

### contacts_list

List all contacts from the address book.

No input parameters.

---

### contacts_get

Get a single contact by ID.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The contact ID |

---

### contacts_create

Create a new contact.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Full name |
| `email` | string | No | Email address |
| `phone` | string | No | Phone number |

---

### contacts_update

Update an existing contact.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The contact ID |
| `name` | string | No | Updated full name |
| `email` | string | No | Updated email address |
| `phone` | string | No | Updated phone number |

---

### contacts_delete

Delete a contact by ID.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The contact ID |

---

### contacts_search

Search contacts by name or email.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `query` | string | Yes | Search query text |

---

## Calendar Tools

### calendar_list

List all available calendars.

No input parameters.

---

### calendar_events

List calendar events within a date range.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `calendar_id` | string | No | Calendar ID (uses default if not specified) |
| `start` | string | Yes | Start date/time in RFC3339 format |
| `end` | string | Yes | End date/time in RFC3339 format |

---

### calendar_get

Get a single calendar event by ID.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The event ID |

---

### calendar_create

Create a new calendar event.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `calendar_id` | string | Yes | Calendar ID |
| `summary` | string | Yes | Event title/summary |
| `description` | string | No | Event description |
| `location` | string | No | Event location |
| `start` | string | Yes | Start date/time in RFC3339 format |
| `end` | string | Yes | End date/time in RFC3339 format |
| `all_day` | boolean | No | Whether this is an all-day event (default: false) |

---

### calendar_update

Update an existing calendar event.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The event ID |
| `calendar_id` | string | No | Calendar ID |
| `summary` | string | No | Event title/summary |
| `description` | string | No | Event description |
| `location` | string | No | Event location |
| `start` | string | No | Start date/time in RFC3339 format |
| `end` | string | No | End date/time in RFC3339 format |

---

### calendar_delete

Delete a calendar event by ID.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The event ID |

---

## Vacation Tools

### vacation_status

Get the current vacation/out-of-office auto-reply status.

No input parameters.

Returns: `{ is_enabled, subject, text_body, from_date?, to_date? }`.

---

### vacation_set

Enable or disable the vacation/out-of-office auto-reply. When `enabled=true`, `subject` and `body` are required.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `enabled` | boolean | Yes | Whether the vacation response is enabled |
| `subject` | string | Yes (if enabled) | Auto-reply subject line |
| `body` | string | Yes (if enabled) | Auto-reply body text |
| `from_date` | string | No | Start date in RFC3339 format |
| `to_date` | string | No | End date in RFC3339 format |

---

## Identity Tools

### identity_list

List all sender identities.

No input parameters.

Returns: array of `{ id, name, email, text_signature, html_signature, reply_to?, bcc?, may_delete }`.

---

### identity_set

Update a sender identity.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The identity ID |
| `name` | string | No | Display name |
| `reply_to` | string | No | Reply-to email address |
| `text_signature` | string | No | Text signature |
| `html_signature` | string | No | HTML signature |

---

## Filter Tools

### filter_list

List all Sieve filter scripts.

No input parameters.

Returns: array of `{ id, name, is_active, blob_id }`.

---

### filter_get

Get a Sieve filter script by ID with its content.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The filter script ID |

---

### filter_create

Create a new Sieve filter script.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | Yes | Name for the filter script |
| `script` | string | Yes | Sieve script content |
| `activate` | boolean | No | Activate the script on creation (default: false) |

---

### filter_activate

Activate a Sieve filter script.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The filter script ID |

---

### filter_deactivate

Deactivate a Sieve filter script.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The filter script ID |

---

### filter_validate

Validate Sieve script syntax without storing it.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `script` | string | Yes | Sieve script content to validate |

---

### filter_delete

Delete a Sieve filter script by ID.

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | string | Yes | The filter script ID |

---

## Account Tools

### quota_get

Get storage quota usage for the account.

No input parameters.

Returns: `{ used, limit, used_percent, used_display, limit_display }`.

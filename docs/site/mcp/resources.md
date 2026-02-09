# MCP Resources Reference

Resources provide read-only context about your FastMail account. Claude can access these to understand your email, contacts, and calendar state.

## Available Resources

### fastmail://inbox

Recent emails from your inbox.

| Property | Value |
|----------|-------|
| URI | `fastmail://inbox` |
| Name | Recent Inbox |
| MIME Type | text/plain |

Returns the 10 most recent inbox messages formatted as markdown with subject, ID, received date, preview, and read status.

---

### fastmail://mail/{id}

Content of a specific email message.

| Property | Value |
|----------|-------|
| URI Template | `fastmail://mail/{id}` |
| Name | Email Message |
| MIME Type | text/plain |

| Parameter | Description |
|-----------|-------------|
| `id` | The email ID (from inbox or search results) |

Returns email subject, ID, thread ID, received date, size, read status, and preview text.

---

### fastmail://contacts

Your address book contacts.

| Property | Value |
|----------|-------|
| URI | `fastmail://contacts` |
| Name | Contacts |
| MIME Type | text/plain |

!!! note "CardDAV Required"
    Contacts require CardDAV configuration. If not configured, the resource directs you to use the `contacts_list` tool instead.

---

### fastmail://contact/{id}

Details of a specific contact.

| Property | Value |
|----------|-------|
| URI Template | `fastmail://contact/{id}` |
| Name | Contact |
| MIME Type | text/plain |

| Parameter | Description |
|-----------|-------------|
| `id` | The contact ID |

!!! note "CardDAV Required"
    Contacts require CardDAV configuration. If not configured, the resource directs you to use the `contacts_get` tool instead.

---

### fastmail://calendar/today

Calendar events for today.

| Property | Value |
|----------|-------|
| URI | `fastmail://calendar/today` |
| Name | Today's Events |
| MIME Type | text/plain |

!!! note "CalDAV Required"
    Calendar access requires CalDAV configuration.

---

### fastmail://masked-emails

Your masked email addresses.

| Property | Value |
|----------|-------|
| URI | `fastmail://masked-emails` |
| Name | Masked Emails |
| MIME Type | text/plain |

Returns all masked email addresses with ID, state, domain, description, and last message date.

---

### fastmail://masked-email/{id}

Details of a specific masked email address.

| Property | Value |
|----------|-------|
| URI Template | `fastmail://masked-email/{id}` |
| Name | Masked Email |
| MIME Type | text/plain |

| Parameter | Description |
|-----------|-------------|
| `id` | The masked email ID |

Returns full masked email details including ID, address, state, domain, description, URL, creator, creation date, and last message date.

---

## Using Resources

Resources are read automatically when Claude needs context. You can also explicitly request them:

```
Read fastmail://inbox and summarize any unread emails
```

Resources are useful for:

1. **Context gathering** -- Give Claude background before asking for actions
2. **Verification** -- Check state before and after operations
3. **Exploration** -- Understand your account structure

## Resources vs Tools

| Aspect | Resources | Tools |
|--------|-----------|-------|
| Purpose | Provide context | Perform actions |
| Access | Read-only | Read/write |
| Format | Formatted text (markdown) | Structured JSON |
| Caching | May be cached | Always fresh |

Use resources when Claude needs to understand your data. Use tools when you want Claude to do something.

## Resource Templates

Some resources use URI templates with parameters:

```
fastmail://mail/{id}
fastmail://contact/{id}
fastmail://masked-email/{id}
```

Replace `{id}` with an actual ID from listing tools or other resources:

```
fastmail://mail/M12345abc
fastmail://masked-email/ME001
```

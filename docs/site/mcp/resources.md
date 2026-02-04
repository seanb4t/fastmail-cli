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

**Content:**

Returns the 10 most recent inbox messages formatted as:

```markdown
# Inbox (10 messages)

## 1. Subject Line
- ID: M123
- Received: 2024-01-15T10:30:00Z
- Preview: First 100 characters of the email...
- Status: unread

## 2. Another Subject
...
```

**Example prompt:**

```
Read my inbox resource to see recent emails
```

---

### fastmail://mail/{id}

Content of a specific email message.

| Property | Value |
|----------|-------|
| URI Template | `fastmail://mail/{id}` |
| Name | Email Message |
| MIME Type | text/plain |

**Parameters:**

| Parameter | Description |
|-----------|-------------|
| `id` | The email ID (from inbox or search results) |

**Content:**

```markdown
# Subject Line

- ID: M123
- Thread: T456
- Received: 2024-01-15T10:30:00Z
- Size: 2048 bytes
- Status: read

## Preview

Email preview text here...
```

**Example prompt:**

```
Read the email fastmail://mail/M123
```

---

### fastmail://contacts

Your address book contacts.

| Property | Value |
|----------|-------|
| URI | `fastmail://contacts` |
| Name | Contacts |
| MIME Type | text/plain |

**Content:**

```markdown
# Contacts (25 total)

## 1. Alice Smith
- ID: C123
- Email: alice@example.com
- Phone: +1-555-1234

## 2. Bob Jones
- ID: C456
- Email: bob@example.com
...
```

**Example prompt:**

```
Read my contacts to find Bob's email address
```

---

### fastmail://calendar/today

Calendar events for today.

| Property | Value |
|----------|-------|
| URI | `fastmail://calendar/today` |
| Name | Today's Events |
| MIME Type | text/plain |

**Content:**

Returns today's calendar events. Requires CalDAV configuration.

**Example prompt:**

```
What's on my calendar today?
```

---

### fastmail://masked-emails

Your masked email addresses.

| Property | Value |
|----------|-------|
| URI | `fastmail://masked-emails` |
| Name | Masked Emails |
| MIME Type | text/plain |

**Content:**

```markdown
# Masked Emails (5 total)

## 1. abc123@fastmail.com
- ID: ME123
- State: enabled
- Domain: example.com
- Description: Newsletter signup
- Last Message: 2024-01-10T15:00:00Z

## 2. xyz789@fastmail.com
...
```

**Example prompt:**

```
Show my masked emails to see which services I've signed up for
```

---

## Using Resources

Resources are read automatically when Claude needs context. You can also explicitly request them:

```
Read fastmail://inbox and summarize any unread emails
```

Resources are useful for:

1. **Context gathering** - Give Claude background before asking for actions
2. **Verification** - Check state before and after operations
3. **Exploration** - Understand your account structure

## Resources vs Tools

| Aspect | Resources | Tools |
|--------|-----------|-------|
| Purpose | Provide context | Perform actions |
| Access | Read-only | Read/write |
| Format | Formatted text | Structured JSON |
| Caching | May be cached | Always fresh |

Use resources when Claude needs to understand your data. Use tools when you want Claude to do something.

## Resource Templates

The email resource uses a URI template with a parameter:

```
fastmail://mail/{id}
```

Replace `{id}` with an actual email ID from inbox or search results:

```
fastmail://mail/M12345abc
```

This pattern allows accessing individual emails while keeping the resource list manageable.

// Package fastmail provides a high-level Go client for Fastmail.
//
// This package offers clean domain types and operations for working with
// Fastmail email, contacts, and calendars. It abstracts away the underlying
// JMAP protocol complexity.
package fastmail

import (
	"slices"
	"time"
)

// Email represents an email message.
//
// This is a domain type designed for ease of use, not a direct mapping
// of the JMAP Email object. Protocol-specific details are hidden.
type Email struct {
	// ID is the unique identifier for this email.
	ID string

	// ThreadID groups related emails in a conversation.
	ThreadID string

	// Subject is the email subject line.
	Subject string

	// From contains the sender's email address.
	From EmailAddress

	// To contains the recipient email addresses.
	To []EmailAddress

	// Cc contains carbon copy recipients.
	Cc []EmailAddress

	// Bcc contains blind carbon copy recipients.
	Bcc []EmailAddress

	// ReceivedAt is when the email was received.
	ReceivedAt time.Time

	// Preview is a short text preview of the email body.
	Preview string

	// Body is the email body content (plain text).
	Body string

	// Keywords contains email flags like "$seen", "$flagged", "$draft".
	Keywords []string

	// MailboxIDs lists the mailboxes containing this email.
	MailboxIDs []string

	// Size is the email size in bytes.
	Size uint64

	// Attachments contains any file attachments on this email.
	Attachments []Attachment
}

// Attachment represents a file attachment on an email.
type Attachment struct {
	// BlobID is the server-side blob identifier for downloading.
	BlobID string

	// Name is the filename of the attachment.
	Name string

	// Type is the MIME content type (e.g., "application/pdf").
	Type string

	// Size is the attachment size in bytes.
	Size uint64

	// Disposition is the Content-Disposition (e.g., "attachment", "inline").
	Disposition string
}

// EmailAddress represents an email address with optional display name.
type EmailAddress struct {
	// Name is the display name (e.g., "John Doe").
	Name string

	// Email is the email address (e.g., "john@example.com").
	Email string
}

// String returns the formatted email address.
// If a name is present, returns "Name <email>", otherwise just the email.
func (a EmailAddress) String() string {
	if a.Name == "" {
		return a.Email
	}
	return a.Name + " <" + a.Email + ">"
}

// Thread represents a conversation thread containing related emails.
type Thread struct {
	// ID is the unique identifier for this thread.
	ID string

	// EmailIDs contains the IDs of emails in this thread, in order.
	EmailIDs []string
}

// Common email keywords as defined by JMAP.
const (
	// KeywordSeen indicates the email has been read.
	KeywordSeen = "$seen"

	// KeywordFlagged indicates the email is starred/flagged.
	KeywordFlagged = "$flagged"

	// KeywordDraft indicates the email is a draft.
	KeywordDraft = "$draft"

	// KeywordAnswered indicates the email has been replied to.
	KeywordAnswered = "$answered"

	// KeywordForwarded indicates the email has been forwarded.
	KeywordForwarded = "$forwarded"
)

// HasKeyword reports whether the email has the specified keyword.
func (e *Email) HasKeyword(keyword string) bool {
	return slices.Contains(e.Keywords, keyword)
}

// IsRead reports whether the email has been read.
func (e *Email) IsRead() bool {
	return e.HasKeyword(KeywordSeen)
}

// IsFlagged reports whether the email is flagged/starred.
func (e *Email) IsFlagged() bool {
	return e.HasKeyword(KeywordFlagged)
}

// IsDraft reports whether the email is a draft.
func (e *Email) IsDraft() bool {
	return e.HasKeyword(KeywordDraft)
}

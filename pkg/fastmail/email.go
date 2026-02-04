// Package fastmail provides a high-level Go client for Fastmail operations.
//
// This package abstracts JMAP protocol details and provides clean domain types
// for working with emails, mailboxes, and other Fastmail resources.
package fastmail

import (
	"slices"
	"time"
)

// Email represents an email message.
//
// This is a domain type that provides a clean interface for working with
// emails without exposing JMAP protocol details.
type Email struct {
	// ID is the unique identifier for this email.
	ID string

	// ThreadID groups related emails in a conversation.
	ThreadID string

	// Subject is the email subject line.
	Subject string

	// From contains the sender addresses.
	From []EmailAddress

	// To contains the primary recipient addresses.
	To []EmailAddress

	// Cc contains carbon copy recipient addresses.
	Cc []EmailAddress

	// Bcc contains blind carbon copy recipient addresses.
	Bcc []EmailAddress

	// ReplyTo contains the reply-to addresses.
	ReplyTo []EmailAddress

	// Date is when the email was sent.
	Date time.Time

	// ReceivedAt is when the email was received by the server.
	ReceivedAt time.Time

	// Preview is a short plaintext preview of the email body.
	Preview string

	// Size is the email size in bytes.
	Size uint64

	// Keywords contains email flags like "seen", "flagged", "draft".
	Keywords []string

	// MailboxIDs lists the mailboxes containing this email.
	MailboxIDs []string
}

// EmailAddress represents an email address with optional display name.
type EmailAddress struct {
	// Name is the display name (e.g., "John Doe").
	Name string

	// Email is the email address (e.g., "john@example.com").
	Email string
}

// String returns a formatted email address.
// If Name is set, returns "Name <Email>", otherwise just Email.
func (a EmailAddress) String() string {
	if a.Name == "" {
		return a.Email
	}
	return a.Name + " <" + a.Email + ">"
}

// IsRead returns true if the email has been read.
func (e *Email) IsRead() bool {
	return e.hasKeyword(KeywordSeen)
}

// IsFlagged returns true if the email is flagged/starred.
func (e *Email) IsFlagged() bool {
	return e.hasKeyword(KeywordFlagged)
}

// IsDraft returns true if the email is a draft.
func (e *Email) IsDraft() bool {
	return e.hasKeyword(KeywordDraft)
}

// IsAnswered returns true if the email has been replied to.
func (e *Email) IsAnswered() bool {
	return e.hasKeyword(KeywordAnswered)
}

func (e *Email) hasKeyword(keyword string) bool {
	return slices.Contains(e.Keywords, keyword)
}

// Standard email keywords.
const (
	// KeywordSeen indicates the email has been read.
	KeywordSeen = "$seen"

	// KeywordFlagged indicates the email is flagged/starred.
	KeywordFlagged = "$flagged"

	// KeywordDraft indicates the email is a draft.
	KeywordDraft = "$draft"

	// KeywordAnswered indicates the email has been replied to.
	KeywordAnswered = "$answered"

	// KeywordForwarded indicates the email has been forwarded.
	KeywordForwarded = "$forwarded"

	// KeywordPhishing indicates the email is suspected phishing.
	KeywordPhishing = "$phishing"

	// KeywordJunk indicates the email is junk/spam.
	KeywordJunk = "$junk"

	// KeywordNotJunk indicates the email is not junk/spam.
	KeywordNotJunk = "$notjunk"
)

// Thread represents a conversation thread containing related emails.
type Thread struct {
	// ID is the unique identifier for this thread.
	ID string

	// EmailIDs contains the IDs of emails in this thread, in date order.
	EmailIDs []string
}

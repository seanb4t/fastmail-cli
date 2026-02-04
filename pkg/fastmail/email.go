// Package fastmail provides a high-level Go API for Fastmail operations.
//
// This package exposes clean domain types that hide the underlying JMAP
// protocol complexity. Types here are designed for application use,
// not protocol serialization.
package fastmail

import (
	"slices"
	"time"
)

// Email represents an email message.
//
// This is a domain type that presents email data in a consumer-friendly
// format. It abstracts away JMAP protocol details like blob IDs and
// provides Go-native types (time.Time instead of RFC3339 strings).
type Email struct {
	// ID uniquely identifies this email within the account.
	ID string

	// ThreadID groups related emails in a conversation.
	ThreadID string

	// Subject is the email subject line.
	Subject string

	// From contains the sender addresses.
	From []EmailAddress

	// To contains the primary recipient addresses.
	To []EmailAddress

	// CC contains carbon copy recipient addresses.
	CC []EmailAddress

	// BCC contains blind carbon copy recipient addresses.
	BCC []EmailAddress

	// ReplyTo specifies where replies should be sent.
	ReplyTo []EmailAddress

	// Date is when the email was sent (from the Date header).
	Date time.Time

	// ReceivedAt is when the email was received by the server.
	ReceivedAt time.Time

	// Preview is a short plaintext excerpt of the email body.
	Preview string

	// Size is the email size in bytes.
	Size uint64

	// MailboxIDs lists the mailboxes containing this email.
	MailboxIDs []string

	// Keywords contains flags/labels applied to this email.
	// Common keywords include Seen, Flagged, Answered, etc.
	Keywords []string
}

// EmailAddress represents an email address with optional display name.
type EmailAddress struct {
	// Name is the display name (e.g., "John Doe").
	Name string

	// Email is the email address (e.g., "john@example.com").
	Email string
}

// String returns the email address in standard format.
// If a name is present, returns "Name <email@example.com>",
// otherwise returns just the email address.
func (a EmailAddress) String() string {
	if a.Name == "" {
		return a.Email
	}
	return a.Name + " <" + a.Email + ">"
}

// Thread represents a conversation thread containing related emails.
type Thread struct {
	// ID uniquely identifies this thread.
	ID string

	// EmailIDs contains the IDs of emails in this thread, ordered by date.
	EmailIDs []string
}

// Standard email keywords (flags) as defined by IMAP/JMAP.
// These constants can be used when filtering or updating email keywords.
const (
	// KeywordSeen indicates the email has been read.
	KeywordSeen = "$seen"

	// KeywordFlagged indicates the email is flagged/starred.
	KeywordFlagged = "$flagged"

	// KeywordAnswered indicates the email has been replied to.
	KeywordAnswered = "$answered"

	// KeywordDraft indicates the email is a draft.
	KeywordDraft = "$draft"

	// KeywordForwarded indicates the email has been forwarded.
	KeywordForwarded = "$forwarded"

	// KeywordPhishing indicates the email is suspected phishing.
	KeywordPhishing = "$phishing"

	// KeywordJunk indicates the email is spam/junk.
	KeywordJunk = "$junk"

	// KeywordNotJunk indicates the email is not spam.
	KeywordNotJunk = "$notjunk"
)

// IsSeen returns true if the email has been read.
func (e *Email) IsSeen() bool {
	return e.hasKeyword(KeywordSeen)
}

// IsFlagged returns true if the email is flagged/starred.
func (e *Email) IsFlagged() bool {
	return e.hasKeyword(KeywordFlagged)
}

// IsAnswered returns true if the email has been replied to.
func (e *Email) IsAnswered() bool {
	return e.hasKeyword(KeywordAnswered)
}

// IsDraft returns true if the email is a draft.
func (e *Email) IsDraft() bool {
	return e.hasKeyword(KeywordDraft)
}

// hasKeyword checks if the email has a specific keyword.
func (e *Email) hasKeyword(keyword string) bool {
	return slices.Contains(e.Keywords, keyword)
}

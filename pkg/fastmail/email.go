// Package fastmail provides a high-level Go API for Fastmail operations.
//
// This package exposes clean domain types that hide JMAP protocol complexity.
// For protocol-level access, see the internal/jmap package.
package fastmail

import "time"

// Email represents an email message.
//
// This is a domain type designed for application use, not a direct mapping
// of JMAP protocol structures.
type Email struct {
	// ID is the unique identifier for this email.
	ID string

	// ThreadID identifies the conversation thread this email belongs to.
	ThreadID string

	// MailboxIDs lists the mailbox IDs containing this email.
	// An email may exist in multiple mailboxes simultaneously.
	MailboxIDs []string

	// Subject is the email subject line.
	Subject string

	// From contains the sender address(es).
	From []EmailAddress

	// To contains the primary recipient addresses.
	To []EmailAddress

	// CC contains carbon copy recipient addresses.
	CC []EmailAddress

	// BCC contains blind carbon copy recipient addresses.
	BCC []EmailAddress

	// ReplyTo contains addresses for replies.
	ReplyTo []EmailAddress

	// Date is when the email was sent.
	Date time.Time

	// ReceivedAt is when the email was received by the server.
	ReceivedAt time.Time

	// Preview is a short plaintext preview of the email body.
	Preview string

	// Keywords contains flags like $seen, $flagged, $draft, $answered.
	Keywords []string

	// Size is the email size in bytes.
	Size int64

	// HasAttachment indicates whether the email has attachments.
	HasAttachment bool
}

// EmailAddress represents a single email address with optional display name.
type EmailAddress struct {
	// Name is the display name (e.g., "John Doe").
	// May be empty if only the email address is known.
	Name string

	// Email is the email address (e.g., "john@example.com").
	Email string
}

// String returns a formatted string representation of the address.
// Returns "Name <email>" if name is present, otherwise just the email.
func (a EmailAddress) String() string {
	if a.Name == "" {
		return a.Email
	}
	return a.Name + " <" + a.Email + ">"
}

// Thread represents an email conversation thread.
//
// A thread groups related emails together, typically by subject and references.
type Thread struct {
	// ID is the unique identifier for this thread.
	ID string

	// EmailIDs contains the IDs of emails in this thread, in chronological order.
	EmailIDs []string
}

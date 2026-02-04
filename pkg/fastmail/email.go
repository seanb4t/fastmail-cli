// Package fastmail provides a high-level Go API for Fastmail operations.
package fastmail

import "time"

// Email represents an email message.
// This is a domain type that abstracts JMAP protocol details.
type Email struct {
	// ID is the unique identifier for this email.
	ID string

	// ThreadID groups related emails in a conversation.
	ThreadID string

	// MailboxIDs contains the IDs of mailboxes this email belongs to.
	MailboxIDs []string

	// Subject is the email subject line.
	Subject string

	// From contains the sender addresses.
	From []EmailAddress

	// To contains the primary recipient addresses.
	To []EmailAddress

	// CC contains the carbon copy recipient addresses.
	CC []EmailAddress

	// BCC contains the blind carbon copy recipient addresses.
	BCC []EmailAddress

	// ReplyTo contains the addresses for replies.
	ReplyTo []EmailAddress

	// Date is when the email was sent.
	Date time.Time

	// ReceivedAt is when the email was received by the server.
	ReceivedAt time.Time

	// Preview is a short plaintext snippet of the email body.
	Preview string

	// TextBody is the plaintext body content.
	TextBody string

	// HTMLBody is the HTML body content.
	HTMLBody string

	// Keywords contains email flags like $seen, $flagged, $draft.
	Keywords Keywords

	// Size is the total size of the email in bytes.
	Size int64

	// MessageID is the Message-ID header value.
	MessageID string

	// InReplyTo is the In-Reply-To header value.
	InReplyTo string

	// References contains the References header values.
	References []string

	// HasAttachment indicates whether the email has attachments.
	HasAttachment bool
}

// EmailAddress represents an email address with optional display name.
type EmailAddress struct {
	// Name is the display name (e.g., "John Doe").
	Name string

	// Email is the email address (e.g., "john@example.com").
	Email string
}

// String returns a formatted email address string.
// If Name is set, returns "Name <email>", otherwise just the email.
func (a EmailAddress) String() string {
	if a.Name == "" {
		return a.Email
	}
	return a.Name + " <" + a.Email + ">"
}

// Keywords represents email flags and labels.
type Keywords struct {
	// Seen indicates the email has been read.
	Seen bool

	// Flagged indicates the email is flagged/starred.
	Flagged bool

	// Answered indicates a reply has been sent.
	Answered bool

	// Draft indicates this is a draft email.
	Draft bool

	// Forwarded indicates the email has been forwarded.
	Forwarded bool

	// Custom contains any custom keywords/labels.
	Custom []string
}

// Thread represents a conversation containing multiple emails.
type Thread struct {
	// ID is the unique identifier for this thread.
	ID string

	// EmailIDs contains the IDs of emails in this thread, ordered by date.
	EmailIDs []string
}

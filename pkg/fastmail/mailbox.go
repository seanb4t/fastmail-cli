package fastmail

// MailboxRole represents a standard mailbox role.
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-2
type MailboxRole string

// Standard mailbox roles defined by JMAP.
const (
	RoleInbox     MailboxRole = "inbox"
	RoleDrafts    MailboxRole = "drafts"
	RoleSent      MailboxRole = "sent"
	RoleTrash     MailboxRole = "trash"
	RoleArchive   MailboxRole = "archive"
	RoleJunk      MailboxRole = "junk"
	RoleImportant MailboxRole = "important"
	RoleAll       MailboxRole = "all"
	RoleFlagged   MailboxRole = "flagged"
)

// Mailbox represents an email folder.
type Mailbox struct {
	// ID is the unique identifier for this mailbox.
	ID string

	// Name is the display name of the mailbox.
	Name string

	// ParentID is the ID of the parent mailbox, empty for top-level.
	ParentID string

	// Role is the standard role for this mailbox, if any.
	Role MailboxRole

	// SortOrder is the position hint for display ordering.
	SortOrder int

	// TotalEmails is the total number of emails in this mailbox.
	TotalEmails int64

	// UnreadEmails is the number of unread emails in this mailbox.
	UnreadEmails int64

	// TotalThreads is the total number of threads in this mailbox.
	TotalThreads int64

	// UnreadThreads is the number of threads with unread emails.
	UnreadThreads int64

	// IsSubscribed indicates if the user is subscribed to this mailbox.
	IsSubscribed bool
}

// IsSystemMailbox returns true if this mailbox has a standard role.
func (m *Mailbox) IsSystemMailbox() bool {
	return m.Role != ""
}

// IsTopLevel returns true if this mailbox has no parent.
func (m *Mailbox) IsTopLevel() bool {
	return m.ParentID == ""
}

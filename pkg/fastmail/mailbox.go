package fastmail

// Mailbox represents an email folder or label.
//
// This is a domain type that provides a clean interface for working with
// mailboxes without exposing JMAP protocol details.
type Mailbox struct {
	// ID is the unique identifier for this mailbox.
	ID string

	// Name is the display name of the mailbox.
	Name string

	// ParentID is the ID of the parent mailbox, empty for top-level.
	ParentID string

	// Role indicates a standard mailbox role, empty for user-created.
	Role MailboxRole

	// SortOrder determines display order among siblings.
	SortOrder uint32

	// TotalEmails is the count of all emails in this mailbox.
	TotalEmails uint64

	// UnreadEmails is the count of unread emails in this mailbox.
	UnreadEmails uint64

	// TotalThreads is the count of threads with emails in this mailbox.
	TotalThreads uint64

	// UnreadThreads is the count of threads with unread emails.
	UnreadThreads uint64
}

// IsUnread returns true if the mailbox has unread emails.
func (m *Mailbox) IsUnread() bool {
	return m.UnreadEmails > 0
}

// IsTopLevel returns true if the mailbox has no parent.
func (m *Mailbox) IsTopLevel() bool {
	return m.ParentID == ""
}

// IsSystemMailbox returns true if this is a standard system mailbox.
func (m *Mailbox) IsSystemMailbox() bool {
	return m.Role != ""
}

// MailboxRole represents standard mailbox roles.
type MailboxRole string

// Standard mailbox roles as defined by JMAP.
// These are well-known roles that email clients can use to identify
// special-purpose mailboxes regardless of their display name.
const (
	// RoleInbox is the primary inbox for incoming mail.
	RoleInbox MailboxRole = "inbox"

	// RoleArchive is for archived messages.
	RoleArchive MailboxRole = "archive"

	// RoleDrafts is for draft messages.
	RoleDrafts MailboxRole = "drafts"

	// RoleJunk is for spam/junk messages.
	RoleJunk MailboxRole = "junk"

	// RoleSent is for sent messages.
	RoleSent MailboxRole = "sent"

	// RoleTrash is for deleted messages.
	RoleTrash MailboxRole = "trash"

	// RoleAll is a virtual mailbox containing all messages.
	RoleAll MailboxRole = "all"

	// RoleImportant is for messages marked as important.
	RoleImportant MailboxRole = "important"

	// RoleScheduledSend is for scheduled outgoing messages.
	RoleScheduledSend MailboxRole = "scheduledsend"

	// RoleSubscribed is for subscribed folders (IMAP compatibility).
	RoleSubscribed MailboxRole = "subscribed"
)

// String returns the role as a string.
func (r MailboxRole) String() string {
	return string(r)
}

// IsValid returns true if the role is a known standard role.
func (r MailboxRole) IsValid() bool {
	switch r {
	case RoleInbox, RoleArchive, RoleDrafts, RoleJunk, RoleSent, RoleTrash,
		RoleAll, RoleImportant, RoleScheduledSend, RoleSubscribed:
		return true
	default:
		return false
	}
}

package fastmail

// Mailbox represents an email folder or label.
//
// This is a domain type that provides a clean interface for working
// with mailboxes. Standard mailbox roles (Inbox, Sent, etc.) are
// identified by the Role field.
type Mailbox struct {
	// ID is the unique identifier for this mailbox.
	ID string

	// Name is the display name of the mailbox.
	Name string

	// Role identifies standard mailboxes (inbox, sent, drafts, etc.).
	// Custom mailboxes have an empty role.
	Role MailboxRole

	// ParentID is the ID of the parent mailbox for nested folders.
	// Empty for top-level mailboxes.
	ParentID string

	// TotalEmails is the count of all emails in this mailbox.
	TotalEmails uint64

	// UnreadEmails is the count of unread emails in this mailbox.
	UnreadEmails uint64

	// TotalThreads is the count of all threads in this mailbox.
	TotalThreads uint64

	// UnreadThreads is the count of threads with unread emails.
	UnreadThreads uint64
}

// MailboxRole identifies standard mailbox types.
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-2
type MailboxRole string

// Standard mailbox roles as defined by JMAP.
const (
	// RoleInbox is the primary inbox for incoming mail.
	RoleInbox MailboxRole = "inbox"

	// RoleDrafts holds draft messages being composed.
	RoleDrafts MailboxRole = "drafts"

	// RoleSent holds sent messages.
	RoleSent MailboxRole = "sent"

	// RoleTrash holds deleted messages before permanent removal.
	RoleTrash MailboxRole = "trash"

	// RoleJunk holds spam/junk messages.
	RoleJunk MailboxRole = "junk"

	// RoleArchive holds archived messages.
	RoleArchive MailboxRole = "archive"

	// RoleAll is a virtual mailbox containing all emails.
	RoleAll MailboxRole = "all"

	// RoleImportant holds important/priority messages.
	RoleImportant MailboxRole = "important"

	// RoleFlagged is a virtual mailbox of flagged messages.
	RoleFlagged MailboxRole = "flagged"
)

// IsStandard reports whether this mailbox has a standard role.
func (m *Mailbox) IsStandard() bool {
	return m.Role != ""
}

// IsInbox reports whether this is the primary inbox.
func (m *Mailbox) IsInbox() bool {
	return m.Role == RoleInbox
}

// IsTrash reports whether this is the trash folder.
func (m *Mailbox) IsTrash() bool {
	return m.Role == RoleTrash
}

// IsJunk reports whether this is the spam/junk folder.
func (m *Mailbox) IsJunk() bool {
	return m.Role == RoleJunk
}

// HasUnread reports whether this mailbox has unread emails.
func (m *Mailbox) HasUnread() bool {
	return m.UnreadEmails > 0
}

package fastmail

// Mailbox represents an email folder/mailbox.
//
// Mailboxes organize emails into folders. Each mailbox can have a special
// role (like Inbox, Sent, Trash) or be a user-created folder with no role.
type Mailbox struct {
	// ID uniquely identifies this mailbox within the account.
	ID string

	// Name is the display name of the mailbox.
	Name string

	// Role indicates the mailbox's special function (if any).
	// User-created folders have an empty role.
	Role MailboxRole

	// ParentID is the ID of the parent mailbox for nested folders.
	// Empty for top-level mailboxes.
	ParentID string

	// SortOrder determines display ordering among sibling mailboxes.
	// Lower values appear first.
	SortOrder uint32

	// TotalEmails is the count of emails in this mailbox.
	TotalEmails uint64

	// UnreadEmails is the count of unread emails in this mailbox.
	UnreadEmails uint64

	// TotalThreads is the count of threads with emails in this mailbox.
	TotalThreads uint64

	// UnreadThreads is the count of threads with unread emails.
	UnreadThreads uint64
}

// MailboxRole indicates a mailbox's special function.
// Standard roles are defined by JMAP (RFC 8621).
type MailboxRole string

// Standard mailbox roles as defined by JMAP.
// These identify system mailboxes with special functions.
const (
	// RoleInbox is the primary mailbox for incoming mail.
	RoleInbox MailboxRole = "inbox"

	// RoleArchive is for archived messages.
	RoleArchive MailboxRole = "archive"

	// RoleDrafts holds draft messages being composed.
	RoleDrafts MailboxRole = "drafts"

	// RoleSent holds sent messages.
	RoleSent MailboxRole = "sent"

	// RoleTrash holds deleted messages before permanent deletion.
	RoleTrash MailboxRole = "trash"

	// RoleJunk holds spam/junk messages.
	RoleJunk MailboxRole = "junk"

	// RoleImportant marks messages flagged as important.
	RoleImportant MailboxRole = "important"

	// RoleAll is a virtual mailbox containing all messages.
	RoleAll MailboxRole = "all"

	// RoleScheduledSend holds messages scheduled for future sending.
	RoleScheduledSend MailboxRole = "scheduledsend"

	// RoleNone indicates a user-created folder with no special role.
	RoleNone MailboxRole = ""
)

// IsSystemMailbox returns true if this mailbox has a system role.
func (m *Mailbox) IsSystemMailbox() bool {
	return m.Role != RoleNone
}

// IsInbox returns true if this is the inbox.
func (m *Mailbox) IsInbox() bool {
	return m.Role == RoleInbox
}

// IsTrash returns true if this is the trash mailbox.
func (m *Mailbox) IsTrash() bool {
	return m.Role == RoleTrash
}

// IsSent returns true if this is the sent mail mailbox.
func (m *Mailbox) IsSent() bool {
	return m.Role == RoleSent
}

// IsDrafts returns true if this is the drafts mailbox.
func (m *Mailbox) IsDrafts() bool {
	return m.Role == RoleDrafts
}

// IsJunk returns true if this is the spam/junk mailbox.
func (m *Mailbox) IsJunk() bool {
	return m.Role == RoleJunk
}

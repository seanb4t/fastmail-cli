package fastmail

// Mailbox represents an email folder or label.
//
// Mailboxes organize emails and may have special roles like inbox or sent.
type Mailbox struct {
	// ID is the unique identifier for this mailbox.
	ID string

	// Name is the display name of the mailbox.
	Name string

	// ParentID is the ID of the parent mailbox, if nested.
	// Empty for top-level mailboxes.
	ParentID string

	// Role indicates a special purpose for this mailbox.
	// Empty for user-created mailboxes without special roles.
	Role MailboxRole

	// TotalEmails is the count of emails in this mailbox.
	TotalEmails int64

	// UnreadEmails is the count of unread emails in this mailbox.
	UnreadEmails int64

	// SortOrder determines display order among sibling mailboxes.
	// Lower values appear first.
	SortOrder int
}

// MailboxRole represents a standard mailbox role.
//
// Standard roles are defined by the JMAP specification and are consistent
// across email providers.
type MailboxRole string

// Standard mailbox roles.
const (
	// RoleInbox is the main inbox where new mail arrives.
	RoleInbox MailboxRole = "inbox"

	// RoleDrafts contains draft messages not yet sent.
	RoleDrafts MailboxRole = "drafts"

	// RoleSent contains copies of sent messages.
	RoleSent MailboxRole = "sent"

	// RoleTrash contains deleted messages awaiting permanent deletion.
	RoleTrash MailboxRole = "trash"

	// RoleJunk contains messages identified as spam.
	RoleJunk MailboxRole = "junk"

	// RoleArchive contains archived messages.
	RoleArchive MailboxRole = "archive"

	// RoleAll is a virtual mailbox containing all emails.
	RoleAll MailboxRole = "all"

	// RoleFlagged is a virtual mailbox containing flagged emails.
	RoleFlagged MailboxRole = "flagged"
)

// IsSpecialRole returns true if the mailbox has a special role.
func (m *Mailbox) IsSpecialRole() bool {
	return m.Role != ""
}

// IsTopLevel returns true if the mailbox has no parent.
func (m *Mailbox) IsTopLevel() bool {
	return m.ParentID == ""
}

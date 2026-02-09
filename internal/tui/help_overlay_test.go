package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpForView_MailboxList(t *testing.T) {
	content := helpForView(viewMailboxList)

	assert.Contains(t, content, "Mailbox List")
	assert.Contains(t, content, "Open mailbox")
	assert.Contains(t, content, "Filter mailboxes")
	assert.Contains(t, content, "Quit")
	assert.Contains(t, content, "Show this help")
}

func TestHelpForView_EmailList(t *testing.T) {
	content := helpForView(viewEmailList)

	assert.Contains(t, content, "Email List")
	assert.Contains(t, content, "Open email")
	assert.Contains(t, content, "Archive")
	assert.Contains(t, content, "Delete (press twice)")
	assert.Contains(t, content, "Toggle read/unread")
	assert.Contains(t, content, "Toggle flag")
	assert.Contains(t, content, "Move to mailbox")
	assert.Contains(t, content, "Compose")
	assert.Contains(t, content, "Search emails")
	assert.Contains(t, content, "Back to mailboxes")
	assert.Contains(t, content, "Show this help")
}

func TestHelpForView_EmailReader(t *testing.T) {
	content := helpForView(viewEmailReader)

	assert.Contains(t, content, "Email Reader")
	assert.Contains(t, content, "Scroll line")
	assert.Contains(t, content, "Half page down/up")
	assert.Contains(t, content, "Top/Bottom")
	assert.Contains(t, content, "Archive")
	assert.Contains(t, content, "Delete (press twice)")
	assert.Contains(t, content, "Toggle read/unread")
	assert.Contains(t, content, "Toggle flag")
	assert.Contains(t, content, "Move to mailbox")
	assert.Contains(t, content, "Attachments")
	assert.Contains(t, content, "Back to email list")
	assert.Contains(t, content, "View thread")
	assert.Contains(t, content, "Show this help")
}

func TestHelpForView_ThreadView(t *testing.T) {
	content := helpForView(viewThreadView)

	assert.Contains(t, content, "Thread View")
	assert.Contains(t, content, "View email")
	assert.Contains(t, content, "Back to reader")
	assert.Contains(t, content, "Show this help")
}

func TestHelpForView_AttachmentPicker(t *testing.T) {
	content := helpForView(viewAttachmentPicker)

	assert.Contains(t, content, "Attachments")
	assert.Contains(t, content, "Download attachment")
	assert.Contains(t, content, "Back to reader")
	assert.Contains(t, content, "Show this help")
}

func TestHelpForView_Compose(t *testing.T) {
	content := helpForView(viewCompose)

	assert.Contains(t, content, "Compose")
	assert.Contains(t, content, "Next field")
	assert.Contains(t, content, "Previous field")
	assert.Contains(t, content, "Send email")
	assert.Contains(t, content, "Cancel")
	assert.Contains(t, content, "Show this help")
}

func TestHelpForView_MovePicker(t *testing.T) {
	content := helpForView(viewMovePicker)

	assert.Empty(t, content)
}

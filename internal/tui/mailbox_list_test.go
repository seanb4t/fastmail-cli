package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func TestMailboxItem_Title(t *testing.T) {
	tests := []struct {
		name     string
		mailbox  fastmail.Mailbox
		expected string
	}{
		{
			name:     "no unread",
			mailbox:  fastmail.Mailbox{Name: "Inbox"},
			expected: "Inbox",
		},
		{
			name:     "with unread",
			mailbox:  fastmail.Mailbox{Name: "Inbox", UnreadEmails: 5},
			expected: "Inbox (5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := mailboxItem{mailbox: tt.mailbox}
			assert.Equal(t, tt.expected, item.Title())
		})
	}
}

func TestMailboxItem_Description(t *testing.T) {
	item := mailboxItem{mailbox: fastmail.Mailbox{TotalEmails: 42}}
	assert.Equal(t, "42 emails", item.Description())
}

func TestMailboxItem_FilterValue(t *testing.T) {
	item := mailboxItem{mailbox: fastmail.Mailbox{Name: "Work"}}
	assert.Equal(t, "Work", item.FilterValue())
}

func TestNewMailboxListModel(t *testing.T) {
	m := newMailboxListModel()

	assert.True(t, m.loading)
	assert.Nil(t, m.selected)
	assert.Equal(t, "Mailboxes", m.list.Title)
}

func TestMailboxListModel_Update_MailboxesLoaded(t *testing.T) {
	m := newMailboxListModel()

	mailboxes := []fastmail.Mailbox{
		{ID: "1", Name: "Inbox", Role: fastmail.RoleInbox, TotalEmails: 10},
		{ID: "2", Name: "Sent", Role: fastmail.RoleSent, TotalEmails: 5},
	}

	m, _ = m.update(mailboxesLoadedMsg{mailboxes: mailboxes})

	assert.False(t, m.loading)
	assert.Equal(t, 2, len(m.list.Items()))
}

func TestMailboxListModel_Update_EnterSelectsMailbox(t *testing.T) {
	m := newMailboxListModel()
	m.list.SetSize(80, 24)

	mailboxes := []fastmail.Mailbox{
		{ID: "1", Name: "Inbox", Role: fastmail.RoleInbox, TotalEmails: 10},
	}
	m, _ = m.update(mailboxesLoadedMsg{mailboxes: mailboxes})

	m, _ = m.update(tea.KeyMsg{Type: tea.KeyEnter})

	require.NotNil(t, m.selected)
	assert.Equal(t, "1", m.selected.ID)
}

func TestMailboxListModel_View_Loading(t *testing.T) {
	m := newMailboxListModel()
	assert.Contains(t, m.view(), "Loading")
}

func TestMailboxesToItems_SortOrder(t *testing.T) {
	mailboxes := []fastmail.Mailbox{
		{ID: "3", Name: "Work", TotalEmails: 5},
		{ID: "1", Name: "Inbox", Role: fastmail.RoleInbox, TotalEmails: 10},
		{ID: "4", Name: "Archive", Role: fastmail.RoleArchive, TotalEmails: 100},
		{ID: "2", Name: "Sent", Role: fastmail.RoleSent, TotalEmails: 20},
		{ID: "5", Name: "Personal", TotalEmails: 3},
	}

	items := mailboxesToItems(mailboxes)

	require.Len(t, items, 5)
	// Standard roles first: Inbox, Sent, Archive
	assert.Equal(t, "Inbox", items[0].(mailboxItem).mailbox.Name)
	assert.Equal(t, "Sent", items[1].(mailboxItem).mailbox.Name)
	assert.Equal(t, "Archive", items[2].(mailboxItem).mailbox.Name)
	// Custom folders alphabetically
	assert.Equal(t, "Personal", items[3].(mailboxItem).mailbox.Name)
	assert.Equal(t, "Work", items[4].(mailboxItem).mailbox.Name)
}

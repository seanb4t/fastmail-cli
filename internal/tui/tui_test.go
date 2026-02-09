package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func TestNew(t *testing.T) {
	client := fastmail.NewClient("https://api.fastmail.com", "test-token")
	m := New(client)

	assert.Equal(t, viewMailboxList, m.view)
	assert.NotNil(t, m.client)
	assert.True(t, m.connecting)
	assert.False(t, m.quit)
}

func TestInit_ReturnsConnectCmd(t *testing.T) {
	m := New(nil)
	cmd := m.Init()

	assert.NotNil(t, cmd)
}

func TestUpdate_Quit(t *testing.T) {
	m := New(nil)
	m.connecting = false

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	model := updated.(Model)

	assert.True(t, model.quit)
	assert.NotNil(t, cmd)
}

func TestUpdate_CtrlC(t *testing.T) {
	m := New(nil)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(Model)

	assert.True(t, model.quit)
	assert.NotNil(t, cmd)
}

func TestUpdate_WindowSize(t *testing.T) {
	m := New(nil)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(Model)

	assert.Equal(t, 120, model.width)
	assert.Equal(t, 40, model.height)
}

func TestUpdate_Connected(t *testing.T) {
	m := New(nil)
	m.connecting = true

	updated, cmd := m.Update(connectedMsg{})
	model := updated.(Model)

	assert.False(t, model.connecting)
	assert.NotNil(t, cmd) // should return fetchMailboxesCmd
}

func TestUpdate_Error(t *testing.T) {
	m := New(nil)
	m.connecting = true

	updated, cmd := m.Update(errMsg{err: assert.AnError})
	model := updated.(Model)

	assert.False(t, model.connecting)
	assert.Equal(t, assert.AnError, model.err)
	assert.Nil(t, cmd)
}

func TestUpdate_MailboxesLoaded(t *testing.T) {
	m := New(nil)
	m.connecting = false

	mailboxes := []fastmail.Mailbox{
		{ID: "1", Name: "Inbox", Role: fastmail.RoleInbox, TotalEmails: 42, UnreadEmails: 3},
		{ID: "2", Name: "Sent", Role: fastmail.RoleSent, TotalEmails: 10},
	}

	updated, _ := m.Update(mailboxesLoadedMsg{mailboxes: mailboxes})
	model := updated.(Model)

	assert.False(t, model.mailboxList.loading)
}

func TestView_Connecting(t *testing.T) {
	m := New(nil)
	m.connecting = true

	assert.Contains(t, m.View(), "Connecting")
}

func TestView_Quit(t *testing.T) {
	m := New(nil)
	m.quit = true

	assert.Empty(t, m.View())
}

func TestView_Error(t *testing.T) {
	m := New(nil)
	m.err = assert.AnError

	assert.Contains(t, m.View(), "Error:")
}

func TestView_MailboxList(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.mailboxList.loading = false
	m.mailboxList.setSize(80, 24)

	v := m.View()
	assert.Contains(t, v, "Mailboxes")
}

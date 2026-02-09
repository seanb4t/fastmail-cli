package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// BDD: As a user, when I select an email in the list,
// I see a preview in the bottom pane without leaving the dashboard.
func TestBDD_Preview_SelectShowsPreview(t *testing.T) {
	// Given: email list with loaded emails
	m := setupDashboardWithEmails(t)
	require.NotNil(t, m.emailList)

	// When: I select the first email
	email := fastmail.Email{ID: "e1", Subject: "Hello", Body: "World",
		From: fastmail.EmailAddress{Name: "Alice", Email: "alice@test.com"}}
	m.emailList.selected = &email
	m, _ = applyUpdate(m, tea.Msg(nil))

	// Then: preview pane is populated
	require.NotNil(t, m.emailReader)
	// And: I'm still in the dashboard view, not fullscreen reader
	assert.NotEqual(t, viewEmailReader, m.view)
}

// BDD: As a user, I can adjust the split between list and preview.
func TestBDD_Preview_AdjustSplit(t *testing.T) {
	m := setupDashboardWithEmails(t)
	initial := m.panes.splitPct

	m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	assert.Greater(t, m.panes.splitPct, initial)
}

// BDD: As a user, pressing Enter on preview opens fullscreen reader.
func TestBDD_Preview_EnterOpensFullscreen(t *testing.T) {
	m := setupDashboardWithEmails(t)
	m.panes.focus = PanePreview
	email := fastmail.Email{ID: "e1", Subject: "Test"}
	er := newEmailReaderModel(email)
	er.setSize(80, 20)
	m.emailReader = &er

	m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, viewEmailReader, m.view)
}

func setupDashboardWithEmails(t *testing.T) Model {
	t.Helper()
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40

	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	el.loading = false
	el.setSize(80, 20)
	emails := []fastmail.Email{
		{ID: "e1", Subject: "Hello", From: fastmail.EmailAddress{Email: "a@b.com"}},
		{ID: "e2", Subject: "World", From: fastmail.EmailAddress{Email: "c@d.com"}},
	}
	el, _ = el.update(emailsLoadedMsg{emails: emails})
	m.emailList = &el
	m.view = viewEmailList
	return m
}

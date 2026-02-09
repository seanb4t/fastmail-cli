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

// BDD: As a user, after selecting an email, Tab cycles focus to the preview pane.
func TestBDD_Preview_TabCyclesToPreview(t *testing.T) {
	m := setupDashboardWithEmails(t)

	// Given: I've selected an email so preview is showing
	email := fastmail.Email{ID: "e1", Subject: "Hello", Body: "World",
		From: fastmail.EmailAddress{Name: "Alice", Email: "alice@test.com"}}
	er := newEmailReaderModel(email)
	er.isPreview = true
	er.setSize(80, 15)
	m.emailReader = &er
	m.panes.focus = PaneEmailList

	// When: I press Tab
	m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyTab})

	// Then: focus moves to PanePreview (not past it)
	assert.Equal(t, PanePreview, m.panes.focus)
}

// BDD: As a user, when preview pane is focused, keys route to the email reader
// (not the email list). Pressing 'j' should NOT move the email list cursor.
func TestBDD_Preview_KeysRouteToReader(t *testing.T) {
	m := setupDashboardWithEmails(t)

	// Given: preview is showing with a loaded email and focus is on preview
	email := fastmail.Email{ID: "e1", Subject: "Hello", Body: "Long email body\n\n\n\n\n\n\n\n\n\n\n\nBottom",
		From: fastmail.EmailAddress{Name: "Alice", Email: "alice@test.com"}}
	er := newEmailReaderModel(email)
	er.isPreview = true
	er.setSize(80, 10)
	// Simulate body loaded so viewport is initialized
	er, _ = er.update(emailBodyLoadedMsg{email: email})
	m.emailReader = &er
	m.panes.focus = PanePreview

	// Record email list cursor before keypress
	cursorBefore := m.emailList.list.Index()

	// When: I press 'j' (which scrolls in reader, moves cursor in list)
	m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	// Then: the email list cursor should NOT have moved (key went to reader, not list)
	assert.Equal(t, cursorBefore, m.emailList.list.Index(),
		"j should route to email reader when preview is focused, not move the email list cursor")
}

// BDD: As a user, pressing 'q' with preview focused dismisses the preview, not quits.
func TestBDD_Preview_QDismissesPreview(t *testing.T) {
	m := setupDashboardWithEmails(t)

	// Given: preview is showing and focused
	email := fastmail.Email{ID: "e1", Subject: "Hello"}
	er := newEmailReaderModel(email)
	er.isPreview = true
	er.setSize(80, 10)
	m.emailReader = &er
	m.panes.focus = PanePreview

	// When: I press 'q'
	m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	// Then: the app should NOT quit
	assert.False(t, m.quit, "q should dismiss preview, not quit the app")
	// And: preview should be cleared
	assert.Nil(t, m.emailReader)
	// And: focus should move back to email list
	assert.Equal(t, PaneEmailList, m.panes.focus)
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

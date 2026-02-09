package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// applyUpdate is a test helper that applies a message and returns the typed model.
func applyUpdate(m Model, msg tea.Msg) (Model, tea.Cmd) {
	updated, cmd := m.Update(msg)
	return updated.(Model), cmd
}

// BDD: As a user, when I launch the TUI, I see a pane-based layout
// with mailbox sidebar, email list, and keybinding bar.
func TestBDD_Layout_InitialState(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40

	mailboxes := []fastmail.Mailbox{
		{ID: "1", Name: "Inbox", Role: fastmail.RoleInbox, UnreadEmails: 5},
		{ID: "2", Name: "Sent", Role: fastmail.RoleSent},
	}
	m, _ = applyUpdate(m, mailboxesLoadedMsg{mailboxes: mailboxes})

	v := m.View()

	// Then: I see the keybinding bar
	assert.Contains(t, v, "tab")
	assert.Contains(t, v, "?")
}

// BDD: As a user, I can toggle the sidebar with 'b'.
func TestBDD_Layout_ToggleSidebar(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	require.True(t, m.panes.sidebar)

	m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})

	assert.False(t, m.panes.sidebar)
}

// BDD: As a user, I can cycle focus with Tab.
func TestBDD_Layout_CycleFocus(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.panes.focus = PaneMailbox

	m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyTab})

	assert.Equal(t, PaneEmailList, m.panes.focus)
}

// BDD: As a user, I can adjust the split with +/-.
func TestBDD_Layout_AdjustSplit(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.view = viewEmailList
	el := newEmailListModel(fastmail.Mailbox{Name: "Inbox", ID: "mb1"})
	m.emailList = &el
	initial := m.panes.splitPct

	var cmd tea.Cmd
	m, cmd = applyUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	assert.Greater(t, m.panes.splitPct, initial)
	assert.Nil(t, cmd) // layout keys don't produce commands

	m, _ = applyUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	assert.Equal(t, initial, m.panes.splitPct)
}

// BDD: Theme auto-detection always returns a valid theme.
func TestBDD_Layout_ThemeDetection(t *testing.T) {
	theme := DetectTheme()

	assert.NotEmpty(t, string(theme.FocusBorder))
	assert.NotEmpty(t, string(theme.PaneBorder))
	assert.NotEmpty(t, string(theme.Unread))
}

// BDD: Pane manager computes valid layout dimensions.
func TestBDD_Layout_ComputesDimensions(t *testing.T) {
	pm := newPaneManager()
	layout := pm.computeLayout(120, 40)

	assert.Greater(t, layout.sidebarWidth, 0)
	assert.Greater(t, layout.mainWidth, 0)
	assert.Greater(t, layout.listHeight, 0)
	assert.Greater(t, layout.previewHeight, 0)
	assert.Equal(t, 120, layout.sidebarWidth+layout.mainWidth+1)
}

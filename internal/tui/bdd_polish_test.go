package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// BDD: As a user on a narrow terminal, the sidebar auto-collapses.
func TestBDD_Polish_NarrowTerminalCollapsesSidebar(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.panes.sidebar = true

	m, _ = applyUpdate(m, tea.WindowSizeMsg{Width: 60, Height: 30})

	layout := m.panes.computeLayout(60, 30)
	assert.Equal(t, 0, layout.sidebarWidth)
}

// BDD: As a user, compose renders as a modal overlay.
func TestBDD_Polish_ComposeIsOverlay(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	cm := newComposeModel()
	cm.setSize(100, 30)
	m.composeView = &cm
	m.view = viewCompose

	v := m.View()
	assert.Contains(t, v, "To:")
}

// BDD: As a user, help renders as a modal overlay.
func TestBDD_Polish_HelpIsOverlay(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	m.helpOverlay = true

	v := m.View()
	assert.Contains(t, v, "help")
}

// BDD: As a user, emails with attachments show an indicator.
func TestBDD_Polish_AttachmentIndicator(t *testing.T) {
	theme := DarkTheme()
	item := emailItem{email: fastmail.Email{
		Subject:     "Report",
		From:        fastmail.EmailAddress{Name: "Boss"},
		Attachments: []fastmail.Attachment{{Name: "q4.pdf"}},
	}}

	title := item.richTitle(theme, 80)
	assert.Contains(t, title, "📎")
}

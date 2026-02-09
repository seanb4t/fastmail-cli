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

// BDD: As a user, when I launch the TUI, I land in the Inbox automatically.
func TestBDD_Polish_AutoSelectInbox(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40

	mailboxes := []fastmail.Mailbox{
		{ID: "mb-drafts", Name: "Drafts", Role: fastmail.RoleDrafts},
		{ID: "mb-inbox", Name: "Inbox", Role: fastmail.RoleInbox, UnreadEmails: 3},
		{ID: "mb-sent", Name: "Sent", Role: fastmail.RoleSent},
	}

	// When: mailboxes finish loading
	m, _ = applyUpdate(m, mailboxesLoadedMsg{mailboxes: mailboxes})

	// Then: I immediately see the email list for Inbox (no empty pane flash)
	assert.Equal(t, viewEmailList, m.view)
	assert.NotNil(t, m.emailList)
	assert.Equal(t, "Inbox", m.emailList.list.Title)
}

// BDD: As a user, when I select a mailbox, the email list fits within the dashboard pane.
func TestBDD_Polish_MailboxSelectionFitsDashboard(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40

	// When: I select a mailbox
	mb := fastmail.Mailbox{ID: "mb1", Name: "Inbox"}
	m, _ = applyUpdate(m, mailboxSelectedMsg{mailbox: mb})

	// Then: the email list width matches the layout-constrained main pane
	layout := m.panes.computeLayout(120, 40)
	assert.Equal(t, layout.mainWidth-2, m.emailList.list.Width())

	// And: the rendered dashboard does not exceed terminal height
	v := m.View()
	lines := len(splitLines(v))
	assert.LessOrEqual(t, lines, 40, "dashboard should not exceed terminal height")
}

// splitLines splits a string into lines, matching terminal row count.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
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

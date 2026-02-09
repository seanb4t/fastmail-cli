package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// BDD: As a user, I see unread/flagged counts in the stats bar.
func TestBDD_Stats_ShowsCounts(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40

	mailboxes := []fastmail.Mailbox{
		{ID: "1", Name: "Inbox", UnreadEmails: 42},
		{ID: "2", Name: "Work", UnreadEmails: 8},
	}
	m, _ = applyUpdate(m, mailboxesLoadedMsg{mailboxes: mailboxes})

	v := m.View()
	assert.Contains(t, v, "50") // 42 + 8 unread
}

// BDD: As a user, I see quota percentage with color coding.
func TestBDD_Stats_ShowsQuota(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	m.statsBar.quota = &fastmail.QuotaInfo{UsedPercent: 73.2}

	v := m.View()
	assert.Contains(t, v, "73")
}

// BDD: As a user, I can collapse the sidebar with 'b'.
func TestBDD_Stats_CollapseSidebar(t *testing.T) {
	m := New(nil)
	m.connecting = false
	m.width = 120
	m.height = 40
	require.True(t, m.panes.sidebar)

	m, _ = applyUpdate(m, keyMsg("b"))

	assert.False(t, m.panes.sidebar)
	// Main area should now have full width
	layout := m.panes.computeLayout(120, 40)
	assert.Equal(t, 120, layout.mainWidth)
}

func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

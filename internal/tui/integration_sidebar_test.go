package tui

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ToggleSidebar(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Wait for data to load — sidebar with "Mailboxes" should be visible.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Mailboxes"))
	}, teatest.WithDuration(5*time.Second))

	// Press 'b' to toggle sidebar off.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})

	// Force a fresh render so the output buffer contains the updated view.
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Wait for the re-render without "Mailboxes" (sidebar hidden).
	// We also confirm email content is still visible.
	var out []byte
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		noMailboxes := !bytes.Contains(bts, []byte("Mailboxes"))
		hasEmails := bytes.Contains(bts, []byte("Weekly Team Standup Notes"))
		if noMailboxes && hasEmails {
			out = make([]byte, len(bts))
			copy(out, bts)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	require.NotEmpty(t, out, "expected rendered output after toggling sidebar off")

	output := string(out)
	assert.NotContains(t, output, "Mailboxes", "sidebar header should NOT be visible after toggle")
	assert.Contains(t, output, "Weekly Team Standup Notes", "email list content should still be visible")
	assert.Contains(t, output, "Inbox", "inbox title should still be visible in email list")

	teatest.RequireEqualOutput(t, out)
}

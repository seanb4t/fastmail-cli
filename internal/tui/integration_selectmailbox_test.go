package tui

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/require"
)

func TestIntegration_SelectMailbox_JK(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Wait for data to load — Inbox appears once mailboxes and emails render.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Inbox"))
	}, teatest.WithDuration(5*time.Second))

	// Focus cycles: PaneEmailList → PanePreview → PaneMailbox (2 tabs).
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})

	// Move cursor down twice: Inbox → Drafts → Sent.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	// Select the Sent mailbox.
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for the Sent mailbox to be selected and rendered.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Sent"))
	}, teatest.WithDuration(5*time.Second))

	// Force a fresh render into the now-drained output buffer.
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 40})

	var out []byte
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		if bytes.Contains(bts, []byte("Sent")) {
			out = make([]byte, len(bts))
			copy(out, bts)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	require.NotEmpty(t, out, "expected rendered dashboard with Sent mailbox selected")
	teatest.RequireEqualOutput(t, out)
}

func TestIntegration_SelectMailbox_Arrows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Wait for data to load — Inbox appears once mailboxes and emails render.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Inbox"))
	}, teatest.WithDuration(5*time.Second))

	// Focus cycles: PaneEmailList → PanePreview → PaneMailbox (2 tabs).
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})

	// Move cursor down twice with arrow keys: Inbox → Drafts → Sent.
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// Select the Sent mailbox.
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for the Sent mailbox to be selected and rendered.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Sent"))
	}, teatest.WithDuration(5*time.Second))

	// Force a fresh render into the now-drained output buffer.
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 40})

	var out []byte
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		if bytes.Contains(bts, []byte("Sent")) {
			out = make([]byte, len(bts))
			copy(out, bts)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	require.NotEmpty(t, out, "expected rendered dashboard with Sent mailbox selected")
	teatest.RequireEqualOutput(t, out)
}

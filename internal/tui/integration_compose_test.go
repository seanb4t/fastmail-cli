package tui

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ComposeEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Wait for data to load — "Inbox" appears once mailboxes and emails render.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Inbox"))
	}, teatest.WithDuration(5*time.Second))

	// Press 'c' to open the compose overlay.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	// Wait for the compose overlay to appear — look for "To:" in output.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("To:"))
	}, teatest.WithDuration(5*time.Second))

	// Send a WindowSizeMsg to trigger a fresh, complete render into the
	// (now drained) output buffer.
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 40})

	var out []byte
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		if bytes.Contains(bts, []byte("Subject:")) {
			out = make([]byte, len(bts))
			copy(out, bts)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	require.NotEmpty(t, out, "expected rendered compose overlay output")
	teatest.RequireEqualOutput(t, out)
}

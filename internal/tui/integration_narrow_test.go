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

func TestIntegration_NarrowTerminal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// 60 columns is below the 80-column sidebar auto-collapse threshold.
	tm := newTestFixture(t, 60, 20)

	// Wait for data to load — "Inbox" appears once mailboxes and emails
	// have been fetched and rendered.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Inbox"))
	}, teatest.WithDuration(5*time.Second))

	// WaitFor drained the output buffer. Send a WindowSizeMsg to trigger
	// a fresh, complete render into the (now empty) buffer.
	tm.Send(tea.WindowSizeMsg{Width: 60, Height: 20})

	var out []byte
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		if bytes.Contains(bts, []byte("Inbox")) {
			out = make([]byte, len(bts))
			copy(out, bts)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	require.NotEmpty(t, out, "expected rendered narrow dashboard output")

	output := string(out)

	// Sidebar should NOT be visible at 60 columns (threshold is 80).
	assert.NotContains(t, output, "Mailboxes", "sidebar header should be hidden in narrow terminal")

	// Email list content should be visible in the main area.
	assert.Contains(t, output, "Inbox", "email list title should be visible")

	// Key bar should be present (may be abbreviated at narrow width).
	assert.Contains(t, output, "quit", "key bar quit hint should be present")

	// Golden file comparison for full output stability.
	teatest.RequireEqualOutput(t, out)
}

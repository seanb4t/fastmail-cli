package tui

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ReadEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Step 1: Wait for emails to load in the dashboard.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Weekly Team Standup Notes"))
	}, teatest.WithDuration(5*time.Second))

	// Step 2: Press Enter to select the first email (cursor starts at top).
	// This creates a preview pane and fires fetchEmailBodyCmd.
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Step 3: Wait for the email body to load in the preview pane.
	// The mock server handles Email/get and returns bodyValues with the full text.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("notes from today's standup"))
	}, teatest.WithDuration(5*time.Second))

	// Step 4: Force a fresh render to capture the complete output.
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 40})

	var out []byte
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		// Wait for the dashboard with the email preview showing the body.
		if bytes.Contains(bts, []byte("From:")) && bytes.Contains(bts, []byte("standup")) {
			out = make([]byte, len(bts))
			copy(out, bts)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	require.NotEmpty(t, out, "expected rendered email reader output")
	teatest.RequireEqualOutput(t, out)
}

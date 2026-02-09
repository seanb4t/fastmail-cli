package tui

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/require"
)

func TestIntegration_DashboardLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// Wait for emails to load — "Weekly Team Standup Notes" appears once
	// both mailboxes and emails have been fetched and rendered.
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Weekly Team Standup Notes"))
	}, teatest.WithDuration(5*time.Second))

	// WaitFor drained the output buffer. Send a WindowSizeMsg to trigger a
	// fresh, complete render of the dashboard into the (now empty) buffer.
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 40})

	var out []byte
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		if bytes.Contains(bts, []byte("Fastmail CLI")) {
			out = make([]byte, len(bts))
			copy(out, bts)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	require.NotEmpty(t, out, "expected rendered dashboard output")
	teatest.RequireEqualOutput(t, out)
}

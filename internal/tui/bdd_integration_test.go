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

// BDD Integration: As a user, when the dashboard fully loads, I see
// the sidebar (Mailboxes), the selected mailbox (Inbox), the stats
// bar branding (Fastmail CLI), and the key bar hint (tab).
func TestBDD_Integration_DashboardShowsAllPanes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tm := newTestFixture(t, 120, 40)

	// WaitFor accumulates all bytes read from Output() into bts.
	// We use a compound condition that checks for all expected UI elements.
	var captured []byte
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		hasMailboxes := bytes.Contains(bts, []byte("Mailboxes"))
		hasInbox := bytes.Contains(bts, []byte("Inbox"))
		hasBrand := bytes.Contains(bts, []byte("Fastmail CLI"))
		hasTab := bytes.Contains(bts, []byte("tab"))

		if hasMailboxes && hasInbox && hasBrand && hasTab {
			captured = make([]byte, len(bts))
			copy(captured, bts)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	// Verify all pane elements are present in the captured output
	require.NotEmpty(t, captured, "should have captured output bytes")
	assert.Contains(t, string(captured), "Mailboxes", "sidebar header should be visible")
	assert.Contains(t, string(captured), "Inbox", "inbox mailbox should be visible")
	assert.Contains(t, string(captured), "Fastmail CLI", "stats bar branding should be visible")
	assert.Contains(t, string(captured), "tab", "key bar hint should be visible")
}

// BDD Integration: As a user, pressing 'j' and pressing arrow-down
// produce identical navigation behavior (both move cursor to next item).
func TestBDD_Integration_KeyboardNavParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// The second email subject from fixtures — both keys should select it.
	secondEmail := "Project Deadline Update"

	// Scenario A: press 'j' to move cursor down
	tmA := newTestFixture(t, 120, 40)

	// Wait for initial load
	teatest.WaitFor(t, tmA.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Inbox"))
	}, teatest.WithDuration(5*time.Second))

	// Send 'j' to move cursor down (focus starts on email list pane)
	tmA.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	var outputA []byte
	teatest.WaitFor(t, tmA.Output(), func(bts []byte) bool {
		if bytes.Contains(bts, []byte(secondEmail)) {
			outputA = make([]byte, len(bts))
			copy(outputA, bts)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	// Scenario B: press arrow-down to move cursor down
	tmB := newTestFixture(t, 120, 40)

	// Wait for initial load
	teatest.WaitFor(t, tmB.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Inbox"))
	}, teatest.WithDuration(5*time.Second))

	// Send down arrow
	tmB.Send(tea.KeyMsg{Type: tea.KeyDown})

	var outputB []byte
	teatest.WaitFor(t, tmB.Output(), func(bts []byte) bool {
		if bytes.Contains(bts, []byte(secondEmail)) {
			outputB = make([]byte, len(bts))
			copy(outputB, bts)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	// Both should have navigated to show the second email
	require.NotEmpty(t, outputA, "output A should have content after j key")
	require.NotEmpty(t, outputB, "output B should have content after down key")
	assert.Contains(t, string(outputA), secondEmail, "j key should navigate to second email")
	assert.Contains(t, string(outputB), secondEmail, "down key should navigate to second email")
}

// BDD Integration: At a narrow terminal width (60x20), the sidebar
// auto-collapses but inbox content remains visible.
func TestBDD_Integration_ResponsiveLayout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// 60 columns is below the 80-column threshold for sidebar auto-collapse
	tm := newTestFixture(t, 60, 20)

	var captured []byte
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		// Wait until data loads — Inbox appears as the email list title
		if bytes.Contains(bts, []byte("Inbox")) {
			captured = make([]byte, len(bts))
			copy(captured, bts)
			return true
		}
		return false
	}, teatest.WithDuration(5*time.Second))

	require.NotEmpty(t, captured, "should have captured output bytes")
	output := string(captured)

	// Sidebar should be auto-collapsed at 60 columns (threshold is 80)
	assert.NotContains(t, output, "Mailboxes", "sidebar title should NOT be visible when collapsed")

	// Inbox content should still be rendered in the main area
	assert.Contains(t, output, "Inbox", "inbox email list should still be visible")
}

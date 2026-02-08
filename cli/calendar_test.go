package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarHelp_ShowsSubcommands(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"calendar", "--help"})

	err := root.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "list")
	assert.Contains(t, output, "show")
	assert.Contains(t, output, "create")
	assert.Contains(t, output, "update")
	assert.Contains(t, output, "delete")
	assert.Contains(t, output, "calendars")
}

func TestCalendarShow_RequiresID(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"calendar", "show"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestCalendarCreate_RequiresSummary(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"calendar", "create", "--start", "2026-01-01T09:00:00Z", "--end", "2026-01-01T10:00:00Z", "--calendar", "default"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

func TestCalendarCreate_RequiresStart(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"calendar", "create", "--summary", "Test Event", "--end", "2026-01-01T10:00:00Z", "--calendar", "default"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

func TestCalendarCreate_RequiresEnd(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"calendar", "create", "--summary", "Test Event", "--start", "2026-01-01T09:00:00Z", "--calendar", "default"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

func TestCalendarCreate_RequiresCalendar(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"calendar", "create", "--summary", "Test Event", "--start", "2026-01-01T09:00:00Z", "--end", "2026-01-01T10:00:00Z"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

func TestCalendarUpdate_RequiresID(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"calendar", "update"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestCalendarDelete_RequiresID(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"calendar", "delete"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestCalendarDelete_RequiresForce(t *testing.T) {
	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"calendar", "delete", "event-123"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deletion canceled")
}

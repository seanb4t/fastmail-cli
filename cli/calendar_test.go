package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCalendarHelp_ShowsSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"calendar", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("calendar --help should not error: %v", err)
	}

	output := buf.String()
	subcommands := []string{"list", "show", "create", "update", "delete", "calendars"}
	for _, sub := range subcommands {
		if !strings.Contains(output, sub) {
			t.Errorf("expected %q subcommand in help, got: %q", sub, output)
		}
	}
}

func TestCalendarCreate_RequiresSummary(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"calendar", "create", "--start", "2026-01-01T10:00", "--end", "2026-01-01T11:00"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("calendar create without --summary should error")
	}
}

func TestCalendarShow_RequiresIDArg(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"calendar", "show"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("calendar show without ID should error")
	}
}

func TestCalendarDelete_RequiresForce(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"calendar", "delete", "event-123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("calendar delete without --force should error")
	}
	if !strings.Contains(err.Error(), "force") {
		t.Errorf("expected 'force' in error, got: %v", err)
	}
}

func TestCalendarUpdate_RequiresIDArg(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"calendar", "update"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("calendar update without ID should error")
	}
}

package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestIdentityHelp_ShowsSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"identity", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("identity --help should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "list") {
		t.Errorf("expected 'list' subcommand in help, got: %q", output)
	}
}

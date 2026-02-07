package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVacationHelp_ShowsSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"vacation", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("vacation --help should not error: %v", err)
	}

	output := buf.String()
	for _, sub := range []string{"show", "set"} {
		if !strings.Contains(output, sub) {
			t.Errorf("expected %q subcommand in help, got: %q", sub, output)
		}
	}
}

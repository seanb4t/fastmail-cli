package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestThreadHelp_ShowsSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"thread", "--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("thread --help should not error: %v", err)
	}
	if !strings.Contains(buf.String(), "show") {
		t.Error("expected 'show' subcommand in help")
	}
}

func TestThreadShow_RequiresIDArg(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"thread", "show"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("thread show without ID should error")
	}
}

package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVacationSet_NoFlags_ReturnsError(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"vacation", "set"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("vacation set with no flags should return an error")
	}
	if !strings.Contains(err.Error(), "at least one option is required") {
		t.Errorf("expected 'at least one option is required' error, got: %v", err)
	}
}

func TestVacationSet_EnableDisableMutuallyExclusive(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"vacation", "set", "--enable", "--disable"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("vacation set with both --enable and --disable should return an error")
	}
	// Cobra's mutually exclusive error says "if any flags in the group [enable disable] are set none of the others can be"
	if !strings.Contains(err.Error(), "if any flags in the group") {
		t.Errorf("expected mutually exclusive flags error, got: %v", err)
	}
}

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

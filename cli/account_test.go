package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestAccountCommand_Help(t *testing.T) {
	var stdout bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"account", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "quota") {
		t.Errorf("expected account help to mention 'quota' subcommand, got: %s", output)
	}
}

func TestAccountQuotaCommand_Help(t *testing.T) {
	var stdout bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"account", "quota", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "storage") || !strings.Contains(output, "quota") {
		t.Errorf("expected quota help to mention storage/quota, got: %s", output)
	}
}

func TestAccountInRootHelp(t *testing.T) {
	var stdout bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "account") {
		t.Errorf("expected root help to mention 'account' subcommand, got: %s", output)
	}
}

package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestMailboxHelp_ShowsSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("mailbox --help should not error: %v", err)
	}

	output := buf.String()

	subcommands := []string{"list", "create", "rename", "delete"}
	for _, sub := range subcommands {
		if !strings.Contains(output, sub) {
			t.Errorf("expected %q subcommand in help, got: %q", sub, output)
		}
	}
}

func TestMailboxCreate_RequiresName(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "create"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mailbox create without --name should error")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected 'required' in error, got: %v", err)
	}
}

func TestMailboxRename_RequiresIDArg(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "rename", "--name", "NewName"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mailbox rename without ID should error")
	}

	if !strings.Contains(err.Error(), "arg") {
		t.Errorf("expected argument error, got: %v", err)
	}
}

func TestMailboxRename_RequiresName(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "rename", "mb-123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mailbox rename without --name should error")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected 'required' in error, got: %v", err)
	}
}

func TestMailboxDelete_RequiresIDArg(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "delete"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mailbox delete without ID should error")
	}

	if !strings.Contains(err.Error(), "arg") {
		t.Errorf("expected argument error, got: %v", err)
	}
}

func TestMailboxDelete_NoForceShowsWarning(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "delete", "mb-123"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("mailbox delete without --force should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Are you sure") {
		t.Errorf("expected confirmation warning, got: %q", output)
	}
	if !strings.Contains(output, "--force") {
		t.Errorf("expected --force hint, got: %q", output)
	}
}

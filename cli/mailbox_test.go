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
		t.Fatal("mailbox create without name should error")
	}

	if !strings.Contains(err.Error(), "arg") {
		t.Errorf("expected argument error, got: %v", err)
	}
}

func TestMailboxRename_RequiresID(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "rename"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mailbox rename without ID should error")
	}

	if !strings.Contains(err.Error(), "arg") {
		t.Errorf("expected argument error, got: %v", err)
	}
}

func TestMailboxRename_RequiresNameFlag(t *testing.T) {
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

func TestMailboxDelete_RequiresID(t *testing.T) {
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

func TestMailboxList_HelpShowsUsage(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "list", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("mailbox list --help should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "List") || !strings.Contains(output, "mailbox") {
		t.Errorf("expected list mailboxes description in help, got: %q", output)
	}
}

func TestMailboxCreate_HelpShowsParentFlag(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "create", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("mailbox create --help should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "parent") {
		t.Error("expected 'parent' flag in create help")
	}
}

func TestMailboxDelete_HelpShowsUsage(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mailbox", "delete", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("mailbox delete --help should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ID") {
		t.Error("expected 'ID' in delete help usage")
	}
}

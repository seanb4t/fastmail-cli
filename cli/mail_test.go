package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestMailHelp_ShowsSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mail", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("mail --help should not error: %v", err)
	}

	output := buf.String()

	// Should show all subcommands
	subcommands := []string{"list", "read", "send", "reply"}
	for _, sub := range subcommands {
		if !strings.Contains(output, sub) {
			t.Errorf("expected %q subcommand in help, got: %q", sub, output)
		}
	}
}

func TestMailRead_RequiresEmailID(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mail", "read"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mail read without email ID should error")
	}

	if !strings.Contains(err.Error(), "arg") {
		t.Errorf("expected argument error, got: %v", err)
	}
}

func TestMailSend_RequiresTo(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mail", "send", "--subject", "Test", "--body", "Body"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mail send without --to should error")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected 'required' in error, got: %v", err)
	}
}

func TestMailSend_RequiresSubject(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mail", "send", "--to", "test@example.com", "--body", "Body"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mail send without --subject should error")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected 'required' in error, got: %v", err)
	}
}

func TestMailSend_RequiresBody(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mail", "send", "--to", "test@example.com", "--subject", "Test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mail send without --body should error")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected 'required' in error, got: %v", err)
	}
}

func TestMailReply_RequiresEmailID(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mail", "reply", "--body", "Reply text"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mail reply without email ID should error")
	}

	// Cobra says "accepts 1 arg(s), received 0"
	if !strings.Contains(err.Error(), "arg") {
		t.Errorf("expected argument error, got: %v", err)
	}
}

func TestMailReply_RequiresBody(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mail", "reply", "email123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("mail reply without --body should error")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected 'required' in error, got: %v", err)
	}
}

func TestMailList_DefaultFlags(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"mail", "list", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("mail list --help should not error: %v", err)
	}

	output := buf.String()

	// Should show default values
	if !strings.Contains(output, "10") {
		t.Error("expected default limit '10' in help")
	}
	if !strings.Contains(output, "Inbox") {
		t.Error("expected default folder 'Inbox' in help")
	}
}

func TestParseAddresses_SimpleEmail(t *testing.T) {
	addrs := parseAddresses([]string{"test@example.com"})
	if len(addrs) != 1 {
		t.Fatalf("expected 1 address, got %d", len(addrs))
	}
	if addrs[0].Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", addrs[0].Email)
	}
	if addrs[0].Name != "" {
		t.Errorf("expected empty name, got %q", addrs[0].Name)
	}
}

func TestParseAddresses_NameAndEmail(t *testing.T) {
	addrs := parseAddresses([]string{"John Doe <john@example.com>"})
	if len(addrs) != 1 {
		t.Fatalf("expected 1 address, got %d", len(addrs))
	}
	if addrs[0].Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %q", addrs[0].Email)
	}
	if addrs[0].Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got %q", addrs[0].Name)
	}
}

func TestParseAddresses_MultipleAddresses(t *testing.T) {
	addrs := parseAddresses([]string{
		"alice@example.com",
		"Bob <bob@example.com>",
	})
	if len(addrs) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(addrs))
	}

	if addrs[0].Email != "alice@example.com" {
		t.Errorf("expected first email 'alice@example.com', got %q", addrs[0].Email)
	}
	if addrs[1].Email != "bob@example.com" {
		t.Errorf("expected second email 'bob@example.com', got %q", addrs[1].Email)
	}
	if addrs[1].Name != "Bob" {
		t.Errorf("expected second name 'Bob', got %q", addrs[1].Name)
	}
}

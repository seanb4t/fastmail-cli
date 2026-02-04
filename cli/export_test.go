package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestExportHelp_ShowsFormats(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"export", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("export --help should not error: %v", err)
	}

	output := buf.String()

	// Should show all formats
	formats := []string{"jsonl", "maildir", "mbox"}
	for _, fmt := range formats {
		if !strings.Contains(output, fmt) {
			t.Errorf("expected %q format in help, got: %q", fmt, output)
		}
	}
}

func TestExportHelp_ShowsFlags(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"export", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("export --help should not error: %v", err)
	}

	output := buf.String()

	// Should show all flags
	flags := []string{"--folder", "--format", "--output"}
	for _, flag := range flags {
		if !strings.Contains(output, flag) {
			t.Errorf("expected %q flag in help, got: %q", flag, output)
		}
	}
}

func TestExportHelp_ShowsDefaults(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"export", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("export --help should not error: %v", err)
	}

	output := buf.String()

	// Should show default values
	if !strings.Contains(output, "Inbox") {
		t.Error("expected default folder 'Inbox' in help")
	}
	if !strings.Contains(output, "jsonl") {
		t.Error("expected default format 'jsonl' in help")
	}
}

func TestExport_InvalidFormat(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"export", "--format", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("export with invalid format should error")
	}

	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' in error, got: %v", err)
	}
}

func TestExport_MaildirRequiresOutput(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"export", "--format", "maildir"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("maildir format without --output should error")
	}

	if !strings.Contains(err.Error(), "--output is required") {
		t.Errorf("expected '--output is required' in error, got: %v", err)
	}
}

func TestExport_ShortFlags(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"export", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("export --help should not error: %v", err)
	}

	output := buf.String()

	// Should show short flags
	if !strings.Contains(output, "-f") {
		t.Error("expected short flag '-f' for folder in help")
	}
	if !strings.Contains(output, "-o") {
		t.Error("expected short flag '-o' for output in help")
	}
}

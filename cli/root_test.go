package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand_Version(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "fastmail-cli") {
		t.Errorf("expected version output to contain 'fastmail-cli', got: %s", output)
	}
}

func TestRootCommand_Help(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "auth") {
		t.Errorf("expected help output to mention 'auth' subcommand, got: %s", output)
	}
	if !strings.Contains(output, "--config") {
		t.Errorf("expected help output to mention '--config' flag, got: %s", output)
	}
	if !strings.Contains(output, "--json") {
		t.Errorf("expected help output to mention '--json' flag, got: %s", output)
	}
	if !strings.Contains(output, "--quiet") {
		t.Errorf("expected help output to mention '--quiet' flag, got: %s", output)
	}
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantFlag string
		wantVal  any
	}{
		{
			name:     "config flag",
			args:     []string{"--config", "/custom/path"},
			wantFlag: "config",
			wantVal:  "/custom/path",
		},
		{
			name:     "json flag",
			args:     []string{"--json"},
			wantFlag: "json",
			wantVal:  true,
		},
		{
			name:     "quiet flag",
			args:     []string{"--quiet"},
			wantFlag: "quiet",
			wantVal:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCommand()
			cmd.SetArgs(tt.args)

			// Execute to parse flags
			_ = cmd.Execute()

			switch v := tt.wantVal.(type) {
			case string:
				got, err := cmd.Flags().GetString(tt.wantFlag)
				if err != nil {
					t.Fatalf("failed to get flag %s: %v", tt.wantFlag, err)
				}
				if got != v {
					t.Errorf("flag %s = %q, want %q", tt.wantFlag, got, v)
				}
			case bool:
				got, err := cmd.Flags().GetBool(tt.wantFlag)
				if err != nil {
					t.Fatalf("failed to get flag %s: %v", tt.wantFlag, err)
				}
				if got != v {
					t.Errorf("flag %s = %v, want %v", tt.wantFlag, got, v)
				}
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// Execute should work as the entry point without panicking
	// We can't easily test stdout in this case, just ensure no error
	err := Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
}

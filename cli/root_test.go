package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCommand_Version(t *testing.T) {
	var stdout bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "fastmail-cli") {
		t.Errorf("expected version output to contain 'fastmail-cli', got: %s", output)
	}
}

func TestRootCommand_Help(t *testing.T) {
	var stdout bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()

	// Should show usage with auth subcommand
	if !strings.Contains(output, "auth") {
		t.Errorf("expected help to mention 'auth' subcommand, got: %s", output)
	}

	// Should show global flags
	for _, flag := range []string{"--config", "--json", "--quiet"} {
		if !strings.Contains(output, flag) {
			t.Errorf("expected help to mention '%s' flag, got: %s", flag, output)
		}
	}
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantJSON  bool
		wantQuiet bool
	}{
		{
			name:      "default flags",
			args:      []string{"--help"},
			wantJSON:  false,
			wantQuiet: false,
		},
		{
			name:      "json flag",
			args:      []string{"--json", "--help"},
			wantJSON:  true,
			wantQuiet: false,
		},
		{
			name:      "quiet flag",
			args:      []string{"--quiet", "--help"},
			wantJSON:  false,
			wantQuiet: true,
		},
		{
			name:      "both flags",
			args:      []string{"--json", "--quiet", "--help"},
			wantJSON:  true,
			wantQuiet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := NewRootCommand()
			root.cmd.SetArgs(tt.args)

			err := root.cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if root.JSONOutput != tt.wantJSON {
				t.Errorf("JSONOutput = %v, want %v", root.JSONOutput, tt.wantJSON)
			}
			if root.Quiet != tt.wantQuiet {
				t.Errorf("Quiet = %v, want %v", root.Quiet, tt.wantQuiet)
			}
		})
	}
}

func TestAuthSubcommands(t *testing.T) {
	// Use temp config to avoid writing to default config path
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"auth help", []string{"auth", "--help"}, false},
		// login requires --token flag or interactive terminal
		{"auth login requires token", []string{"--config", configPath, "auth", "login"}, true},
		// logout and status now work (implemented in auth.go)
		{"auth logout works", []string{"--config", configPath, "auth", "logout"}, false},
		{"auth status works", []string{"--config", configPath, "auth", "status"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			cmd := NewRootCommand()
			cmd.SetOut(&stdout)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

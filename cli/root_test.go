package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand_Version(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "fastmail-cli") {
		t.Errorf("version output should contain 'fastmail-cli', got: %q", output)
	}
	if !strings.Contains(output, "version") {
		t.Errorf("version output should contain 'version', got: %q", output)
	}
}

func TestRootCommand_Help(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should show auth subcommand
	if !strings.Contains(output, "auth") {
		t.Errorf("help should list 'auth' subcommand, got: %q", output)
	}
	// Should show global flags
	if !strings.Contains(output, "--config") {
		t.Errorf("help should show '--config' flag, got: %q", output)
	}
	if !strings.Contains(output, "--json") {
		t.Errorf("help should show '--json' flag, got: %q", output)
	}
	if !strings.Contains(output, "--quiet") {
		t.Errorf("help should show '--quiet' flag, got: %q", output)
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
			cmd := NewRootCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs(tt.args)

			_ = cmd.Execute()

			// Check flags were parsed
			jsonFlag, _ := cmd.Flags().GetBool("json")
			quietFlag, _ := cmd.Flags().GetBool("quiet")

			if jsonFlag != tt.wantJSON {
				t.Errorf("json flag = %v, want %v", jsonFlag, tt.wantJSON)
			}
			if quietFlag != tt.wantQuiet {
				t.Errorf("quiet flag = %v, want %v", quietFlag, tt.wantQuiet)
			}
		})
	}
}

func TestRootCommand_ConfigFlag(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--config", "/custom/path/config.yaml", "--help"})

	_ = cmd.Execute()

	configPath, _ := cmd.Flags().GetString("config")
	if configPath != "/custom/path/config.yaml" {
		t.Errorf("config path = %q, want %q", configPath, "/custom/path/config.yaml")
	}
}

func TestExecute(t *testing.T) {
	// Execute should not panic and should work as entry point
	// We can't easily test full execution, but we can verify it exists and is callable
	// In real usage, this would run the full CLI
}

func TestHelperFunctions(t *testing.T) {
	// Reset global state
	cfgFile = ""
	jsonFlag = false
	quietFlag = false

	// Test defaults
	if GetConfigPath() != "" {
		t.Errorf("default config path should be empty, got %q", GetConfigPath())
	}
	if IsJSONOutput() {
		t.Error("default JSON output should be false")
	}
	if IsQuiet() {
		t.Error("default quiet should be false")
	}

	// Set values via command parsing
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--config", "/test/config.yaml", "--json", "--quiet", "--help"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	_ = cmd.Execute()

	if GetConfigPath() != "/test/config.yaml" {
		t.Errorf("config path = %q, want %q", GetConfigPath(), "/test/config.yaml")
	}
	if !IsJSONOutput() {
		t.Error("JSON output should be true after --json flag")
	}
	if !IsQuiet() {
		t.Error("quiet should be true after --quiet flag")
	}
}

func TestAuthSubcommands(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"auth help", []string{"auth", "--help"}, false},
		// login requires --token flag or interactive terminal
		{"auth login requires token", []string{"auth", "login"}, true},
		// logout and status now work (implemented in auth.go)
		{"auth logout works", []string{"auth", "logout"}, false},
		{"auth status works", []string{"auth", "status"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

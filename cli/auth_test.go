package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthStatus_NoToken(t *testing.T) {
	// Setup: ensure no token exists
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Clear env var
	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	_ = os.Unsetenv("FASTMAIL_TOKEN")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("FASTMAIL_TOKEN", originalEnv)
		}
	}()

	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--config", configPath, "auth", "status"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth status should not error when no token: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Not logged in") {
		t.Errorf("expected 'Not logged in' in output, got: %q", output)
	}
}

func TestAuthStatus_WithToken(t *testing.T) {
	// Setup: set token via env var
	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	_ = os.Setenv("FASTMAIL_TOKEN", "test-token-12345")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("FASTMAIL_TOKEN", originalEnv)
		} else {
			_ = os.Unsetenv("FASTMAIL_TOKEN")
		}
	}()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--config", configPath, "auth", "status"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth status should not error with token: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Logged in") {
		t.Errorf("expected 'Logged in' in output, got: %q", output)
	}
	// Should NOT show the actual token
	if strings.Contains(output, "test-token-12345") {
		t.Error("auth status should not display the actual token")
	}
}

func TestAuthHelp_ShowsSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"auth", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth help should not error: %v", err)
	}

	output := buf.String()
	// Should show all three subcommands
	if !strings.Contains(output, "login") {
		t.Errorf("auth help should show 'login' subcommand, got: %q", output)
	}
	if !strings.Contains(output, "logout") {
		t.Errorf("auth help should show 'logout' subcommand, got: %q", output)
	}
	if !strings.Contains(output, "status") {
		t.Errorf("auth help should show 'status' subcommand, got: %q", output)
	}
}

func TestAuthLogout_RemovesToken(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a config file with a token
	err := os.WriteFile(configPath, []byte("token: test-token\n"), 0600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Clear env var
	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	_ = os.Unsetenv("FASTMAIL_TOKEN")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("FASTMAIL_TOKEN", originalEnv)
		}
	}()

	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--config", configPath, "auth", "logout"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("auth logout should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Logged out") {
		t.Errorf("expected 'Logged out' in output, got: %q", output)
	}

	// Verify token was removed - run status to check
	cmd2 := NewRootCommand()
	buf2 := new(bytes.Buffer)
	cmd2.SetOut(buf2)
	cmd2.SetErr(buf2)
	cmd2.SetArgs([]string{"--config", configPath, "auth", "status"})

	_ = cmd2.Execute()
	if !strings.Contains(buf2.String(), "Not logged in") {
		t.Errorf("after logout, status should show 'Not logged in', got: %q", buf2.String())
	}
}

func TestAuthLogin_StoresToken(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Clear env var
	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	_ = os.Unsetenv("FASTMAIL_TOKEN")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("FASTMAIL_TOKEN", originalEnv)
		}
	}()

	// Test with --token flag (non-interactive mode)
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--config", configPath, "auth", "login", "--token", "my-api-token"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth login should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Logged in") {
		t.Errorf("expected 'Logged in' in output, got: %q", output)
	}

	// Verify token was stored - run status to check
	cmd2 := NewRootCommand()
	buf2 := new(bytes.Buffer)
	cmd2.SetOut(buf2)
	cmd2.SetErr(buf2)
	cmd2.SetArgs([]string{"--config", configPath, "auth", "status"})

	_ = cmd2.Execute()
	if !strings.Contains(buf2.String(), "Logged in") {
		t.Errorf("after login, status should show 'Logged in', got: %q", buf2.String())
	}
}

func TestAuthLogin_RequiresToken(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// When not a terminal and no --token flag, should error
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--config", configPath, "auth", "login"})

	err := cmd.Execute()
	// Should error because no token provided and not interactive
	if err == nil {
		t.Error("auth login without token should error in non-interactive mode")
	}
}

func TestAuthStatus_JSONOutput(t *testing.T) {
	// Clear env var
	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	_ = os.Unsetenv("FASTMAIL_TOKEN")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("FASTMAIL_TOKEN", originalEnv)
		}
	}()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--config", configPath, "--json", "auth", "status"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth status --json should not error: %v", err)
	}

	output := buf.String()
	// Should be valid JSON with logged_in field
	if !strings.Contains(output, `"logged_in"`) {
		t.Errorf("JSON output should contain 'logged_in' field, got: %q", output)
	}
	if !strings.Contains(output, "false") {
		t.Errorf("JSON output should show logged_in: false, got: %q", output)
	}
}

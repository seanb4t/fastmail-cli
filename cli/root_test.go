package cli

import (
	"bytes"
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
		name     string
		args     []string
		checkFn  func(cmd *RootCommand) error
	}{
		{
			name: "config flag",
			args: []string{"--config", "/custom/path.yaml"},
			checkFn: func(cmd *RootCommand) error {
				if cmd.ConfigFile != "/custom/path.yaml" {
					return errorf("expected config '/custom/path.yaml', got '%s'", cmd.ConfigFile)
				}
				return nil
			},
		},
		{
			name: "json flag",
			args: []string{"--json"},
			checkFn: func(cmd *RootCommand) error {
				if !cmd.JSONOutput {
					return errorf("expected JSONOutput true, got false")
				}
				return nil
			},
		},
		{
			name: "quiet flag",
			args: []string{"--quiet"},
			checkFn: func(cmd *RootCommand) error {
				if !cmd.Quiet {
					return errorf("expected Quiet true, got false")
				}
				return nil
			},
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

			if err := tt.checkFn(root); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// Execute() should work as the entry point
	// We can't fully test this without mocking os.Exit,
	// but we can verify it doesn't panic
	err := Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
}

// errorf is a helper to create formatted errors for test checks
func errorf(format string, args ...any) error {
	return &testError{msg: sprintf(format, args...)}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func sprintf(format string, args ...any) string {
	// Simple sprintf implementation without importing fmt in test
	result := format
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			result = strings.Replace(result, "%s", v, 1)
		case bool:
			if v {
				result = strings.Replace(result, "%v", "true", 1)
			} else {
				result = strings.Replace(result, "%v", "false", 1)
			}
		}
	}
	return result
}

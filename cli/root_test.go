package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand_Version(t *testing.T) {
	// Capture output
	out := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(out)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "fastmail-cli")
	assert.Contains(t, output, "version")
}

func TestRootCommand_Help(t *testing.T) {
	// Capture output
	out := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	// Should show usage with subcommands
	assert.Contains(t, output, "fastmail-cli")
	assert.Contains(t, output, "auth")
	assert.Contains(t, output, "--config")
	assert.Contains(t, output, "--json")
	assert.Contains(t, output, "--quiet")
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantFlag string
	}{
		{"config flag", []string{"--config", "/custom/path", "--help"}, "--config"},
		{"json flag", []string{"--json", "--help"}, "--json"},
		{"quiet flag", []string{"--quiet", "--help"}, "--quiet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCommand()
			cmd.SetArgs(tt.args)

			// Flags should be parseable without error
			err := cmd.Execute()
			assert.NoError(t, err)
		})
	}
}

func TestExecute(t *testing.T) {
	// Execute should work as entry point
	// Note: This captures the basic contract that Execute() returns error
	err := Execute()
	// With no args, root command shows help and returns nil
	assert.NoError(t, err)
}

func TestRootCommand_ConfigFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--config", "/custom/config.yaml"})

	// Parse flags
	err := cmd.ParseFlags([]string{"--config", "/custom/config.yaml"})
	require.NoError(t, err)

	// Verify flag value is accessible
	configPath, err := cmd.Flags().GetString("config")
	require.NoError(t, err)
	assert.Equal(t, "/custom/config.yaml", configPath)
}

func TestRootCommand_JSONFlag(t *testing.T) {
	cmd := NewRootCommand()

	err := cmd.ParseFlags([]string{"--json"})
	require.NoError(t, err)

	jsonOutput, err := cmd.Flags().GetBool("json")
	require.NoError(t, err)
	assert.True(t, jsonOutput)
}

func TestRootCommand_QuietFlag(t *testing.T) {
	cmd := NewRootCommand()

	err := cmd.ParseFlags([]string{"--quiet"})
	require.NoError(t, err)

	quiet, err := cmd.Flags().GetBool("quiet")
	require.NoError(t, err)
	assert.True(t, quiet)
}

func TestRootCommand_HasAuthSubcommand(t *testing.T) {
	cmd := NewRootCommand()

	// Find auth subcommand
	var found bool
	for _, sub := range cmd.Commands() {
		if sub.Use == "auth" || strings.HasPrefix(sub.Use, "auth ") {
			found = true
			break
		}
	}
	assert.True(t, found, "root command should have 'auth' subcommand")
}

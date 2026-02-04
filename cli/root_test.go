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
	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "fastmail-cli")
	assert.Contains(t, output, "version")
}

func TestRootCommand_Help(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should show usage
	assert.Contains(t, output, "Usage:")
	// Should mention auth subcommand
	assert.Contains(t, output, "auth")
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flagName string
		check    func(t *testing.T, cmd *rootCmd)
	}{
		{
			name:     "config flag",
			args:     []string{"--config", "/custom/config.yaml"},
			flagName: "config",
			check: func(t *testing.T, cmd *rootCmd) {
				val, err := cmd.cmd.Flags().GetString("config")
				require.NoError(t, err)
				assert.Equal(t, "/custom/config.yaml", val)
			},
		},
		{
			name:     "json flag",
			args:     []string{"--json"},
			flagName: "json",
			check: func(t *testing.T, cmd *rootCmd) {
				val, err := cmd.cmd.Flags().GetBool("json")
				require.NoError(t, err)
				assert.True(t, val)
			},
		},
		{
			name:     "quiet flag",
			args:     []string{"--quiet"},
			flagName: "quiet",
			check: func(t *testing.T, cmd *rootCmd) {
				val, err := cmd.cmd.Flags().GetBool("quiet")
				require.NoError(t, err)
				assert.True(t, val)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newRootCmd()
			root.cmd.SetArgs(tt.args)
			// Parse flags without running
			err := root.cmd.ParseFlags(tt.args)
			require.NoError(t, err)
			tt.check(t, root)
		})
	}
}

func TestRootCommand_Execute(t *testing.T) {
	// Execute should not error with no args
	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestExecute(_ *testing.T) {
	// Test the public Execute() function works
	// We can't easily capture output here, but we can verify it doesn't panic
	// In a real scenario this would test the full CLI flow
	err := Execute()
	// May error if auth not configured, but shouldn't panic
	_ = err
}

func TestRootCommand_AuthSubcommandExists(t *testing.T) {
	root := newRootCmd()

	// Check auth subcommand is registered
	authCmd, _, err := root.cmd.Find([]string{"auth"})
	require.NoError(t, err)
	assert.NotNil(t, authCmd)
	assert.Equal(t, "auth", authCmd.Name())
}

func TestRootCommand_ShortFlagAliases(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flagName string
		expected any
	}{
		{
			name:     "short config flag",
			args:     []string{"-c", "/custom/config.yaml"},
			flagName: "config",
			expected: "/custom/config.yaml",
		},
		{
			name:     "short quiet flag",
			args:     []string{"-q"},
			flagName: "quiet",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newRootCmd()
			root.cmd.SetArgs(tt.args)
			err := root.cmd.ParseFlags(tt.args)
			require.NoError(t, err)

			switch expected := tt.expected.(type) {
			case string:
				val, err := root.cmd.Flags().GetString(tt.flagName)
				require.NoError(t, err)
				assert.Equal(t, expected, val)
			case bool:
				val, err := root.cmd.Flags().GetBool(tt.flagName)
				require.NoError(t, err)
				assert.Equal(t, expected, val)
			}
		})
	}
}

func TestRootCommand_VersionFormat(t *testing.T) {
	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Version should be on its own line and include version info
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.NotEmpty(t, lines)
	// First line should contain version info
	assert.True(t, strings.Contains(lines[0], "fastmail-cli") || strings.Contains(lines[0], "version"))
}

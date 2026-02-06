package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	expected := filepath.Join(homeDir, ".config", "fastmail-cli", "config.yaml")
	assert.Equal(t, expected, path)
}

func TestLoad_DefaultValues(t *testing.T) {
	// Use a temp directory with no config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Clear any environment variables that could interfere
	t.Setenv("FASTMAIL_ENDPOINT", "")
	t.Setenv("FASTMAIL_CARDDAV_ENDPOINT", "")
	t.Setenv("FASTMAIL_OUTPUT_FORMAT", "")
	t.Setenv("FASTMAIL_TOKEN", "")

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// Verify defaults
	assert.Equal(t, "https://api.fastmail.com/jmap/session", cfg.Endpoint)
	assert.Equal(t, "https://carddav.fastmail.com/dav/", cfg.CardDAVEndpoint)
	assert.Equal(t, "auto", cfg.OutputFormat)
	assert.Equal(t, "", cfg.Token)
}

func TestLoad_EnvironmentOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Set environment variables
	t.Setenv("FASTMAIL_ENDPOINT", "https://custom.endpoint.com/jmap/session")
	t.Setenv("FASTMAIL_OUTPUT_FORMAT", "json")
	t.Setenv("FASTMAIL_TOKEN", "env-token-123")

	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, "https://custom.endpoint.com/jmap/session", cfg.Endpoint)
	assert.Equal(t, "json", cfg.OutputFormat)
	assert.Equal(t, "env-token-123", cfg.Token)
}

func TestLoad_FileConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write config file
	configContent := `endpoint: https://file.endpoint.com/jmap/session
output_format: text
token: file-token-456
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Clear environment variables
	t.Setenv("FASTMAIL_ENDPOINT", "")
	t.Setenv("FASTMAIL_OUTPUT_FORMAT", "")
	t.Setenv("FASTMAIL_TOKEN", "")

	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, "https://file.endpoint.com/jmap/session", cfg.Endpoint)
	assert.Equal(t, "text", cfg.OutputFormat)
	assert.Equal(t, "file-token-456", cfg.Token)
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write config file
	configContent := `endpoint: https://file.endpoint.com/jmap/session
output_format: text
token: file-token-456
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set ONE environment variable - it should override just that value
	t.Setenv("FASTMAIL_ENDPOINT", "https://env.endpoint.com/jmap/session")
	t.Setenv("FASTMAIL_OUTPUT_FORMAT", "")
	t.Setenv("FASTMAIL_TOKEN", "")

	cfg, err := Load(configPath)
	require.NoError(t, err)

	// Env should override file for endpoint
	assert.Equal(t, "https://env.endpoint.com/jmap/session", cfg.Endpoint)
	// File values should remain for the rest
	assert.Equal(t, "text", cfg.OutputFormat)
	assert.Equal(t, "file-token-456", cfg.Token)
}

func TestLoad_PreservesCardDAVEndpointFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `endpoint: https://api.example.com/jmap/session
carddav_endpoint: https://carddav.override.test/dav/
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	t.Setenv("FASTMAIL_ENDPOINT", "")
	t.Setenv("FASTMAIL_CARDDAV_ENDPOINT", "")

	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, "https://api.example.com/jmap/session", cfg.Endpoint)
	assert.Equal(t, "https://carddav.override.test/dav/", cfg.CardDAVEndpoint)
}

func TestLoad_PreservesCardDAVEndpointFromEnv(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `endpoint: https://api.example.com/jmap/session
carddav_endpoint: https://carddav.file.test/dav/
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	t.Setenv("FASTMAIL_ENDPOINT", "https://api.env.test/jmap/session")
	t.Setenv("FASTMAIL_CARDDAV_ENDPOINT", "https://carddav.env.test/dav/")

	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, "https://api.env.test/jmap/session", cfg.Endpoint)
	assert.Equal(t, "https://carddav.env.test/dav/", cfg.CardDAVEndpoint)
}

func TestDeriveCardDAVEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected string
	}{
		{
			name:     "empty defaults",
			endpoint: "",
			expected: defaultCardDAVEndpoint,
		},
		{
			name:     "api host converts to carddav host",
			endpoint: "https://api.fastmail.com/jmap/session",
			expected: "https://carddav.fastmail.com/dav/",
		},
		{
			name:     "non-api host preserved",
			endpoint: "https://mail.example.com/jmap/session",
			expected: "https://mail.example.com/dav/",
		},
		{
			name:     "invalid url defaults",
			endpoint: "://bad",
			expected: defaultCardDAVEndpoint,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, deriveCardDAVEndpoint(tc.endpoint))
		})
	}
}

func TestConfig_Validate_ValidOutputFormats(t *testing.T) {
	validFormats := []string{"auto", "json", "text"}

	for _, format := range validFormats {
		t.Run(format, func(t *testing.T) {
			cfg := &Config{
				Endpoint:     "https://api.fastmail.com/jmap/session",
				OutputFormat: format,
			}
			err := cfg.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestConfig_Validate_InvalidOutputFormat(t *testing.T) {
	cfg := &Config{
		Endpoint:     "https://api.fastmail.com/jmap/session",
		OutputFormat: "invalid",
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "output_format")
}

func TestConfig_Validate_EmptyEndpoint(t *testing.T) {
	cfg := &Config{
		Endpoint:     "",
		OutputFormat: "auto",
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint")
}

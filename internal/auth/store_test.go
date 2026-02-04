package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_EnvToken(t *testing.T) {
	// Environment variable should take highest priority
	t.Setenv("FASTMAIL_TOKEN", "env-token-123")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	store := NewStore(configPath)
	store.DisableKeychain() // Don't use system keychain in tests

	token, err := store.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "env-token-123", token)

	assert.True(t, store.HasToken())
}

func TestStore_FileToken(t *testing.T) {
	// Clear environment variable
	t.Setenv("FASTMAIL_TOKEN", "")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write token to config file
	configContent := `token: file-token-456
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	store := NewStore(configPath)
	store.DisableKeychain() // Don't use system keychain in tests

	token, err := store.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "file-token-456", token)

	assert.True(t, store.HasToken())
}

func TestStore_NoToken(t *testing.T) {
	// Clear environment variable
	t.Setenv("FASTMAIL_TOKEN", "")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml") // No file exists

	store := NewStore(configPath)
	store.DisableKeychain() // Don't use system keychain in tests

	token, err := store.GetToken()
	assert.Error(t, err)
	assert.Equal(t, "", token)
	assert.Contains(t, err.Error(), "no token")

	assert.False(t, store.HasToken())
}

func TestStore_EnvOverridesFile(t *testing.T) {
	// Env should take priority over file
	t.Setenv("FASTMAIL_TOKEN", "env-token-priority")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write different token to file
	configContent := `token: file-token-should-be-ignored
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	store := NewStore(configPath)
	store.DisableKeychain()

	token, err := store.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "env-token-priority", token)
}

func TestStore_SetToken(t *testing.T) {
	t.Setenv("FASTMAIL_TOKEN", "")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	store := NewStore(configPath)
	store.DisableKeychain()

	// Set a new token
	err := store.SetToken("new-token-789")
	require.NoError(t, err)

	// Verify it can be retrieved
	token, err := store.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "new-token-789", token)

	// Verify file was created with correct permissions
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestStore_DeleteToken(t *testing.T) {
	t.Setenv("FASTMAIL_TOKEN", "")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write token to file
	configContent := `token: token-to-delete
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	store := NewStore(configPath)
	store.DisableKeychain()

	// Verify token exists first
	assert.True(t, store.HasToken())

	// Delete the token
	err = store.DeleteToken()
	require.NoError(t, err)

	// Verify token is gone
	assert.False(t, store.HasToken())

	_, err = store.GetToken()
	assert.Error(t, err)
}

func TestStore_EmptyEnvToken(t *testing.T) {
	// Empty env var should fall through to file
	t.Setenv("FASTMAIL_TOKEN", "")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `token: file-fallback-token
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	store := NewStore(configPath)
	store.DisableKeychain()

	token, err := store.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "file-fallback-token", token)
}

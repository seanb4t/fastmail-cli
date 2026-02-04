// Package auth provides authentication and credential storage.
//
// Token sources (in priority order, highest to lowest):
//  1. Environment variable (FASTMAIL_TOKEN)
//  2. System keychain (via go-keyring)
//  3. Config file (~/.config/fastmail-cli/config.yaml)
package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

const (
	// EnvToken is the environment variable name for the API token.
	EnvToken = "FASTMAIL_TOKEN"

	// KeyringService is the service name used in system keychain.
	KeyringService = "fastmail-cli"

	// KeyringUser is the user name used in system keychain.
	KeyringUser = "api-token"
)

// Store manages credential storage with layered lookup.
type Store struct {
	configPath      string
	keychainEnabled bool
}

// NewStore creates a new Store with the given config file path.
func NewStore(configPath string) *Store {
	return &Store{
		configPath:      configPath,
		keychainEnabled: true,
	}
}

// DisableKeychain disables keychain access for testing.
func (s *Store) DisableKeychain() {
	s.keychainEnabled = false
}

// GetToken retrieves the token from the highest-priority source available.
// Priority: environment > keychain > config file.
func (s *Store) GetToken() (string, error) {
	// 1. Check environment variable
	if token := os.Getenv(EnvToken); token != "" {
		return token, nil
	}

	// 2. Check system keychain (if enabled)
	if s.keychainEnabled {
		token, err := keyring.Get(KeyringService, KeyringUser)
		if err == nil && token != "" {
			return token, nil
		}
		// Ignore keychain errors - fall through to file
	}

	// 3. Check config file
	token, err := s.getFileToken()
	if err == nil && token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no token found: set %s environment variable, use keychain, or add token to config file", EnvToken)
}

// SetToken stores the token. Uses keychain if enabled, otherwise file.
func (s *Store) SetToken(token string) error {
	if s.keychainEnabled {
		err := keyring.Set(KeyringService, KeyringUser, token)
		if err == nil {
			return nil
		}
		// Fall through to file if keychain fails
	}

	return s.setFileToken(token)
}

// DeleteToken removes the token from all storage locations.
func (s *Store) DeleteToken() error {
	var lastErr error

	// Delete from keychain if enabled
	if s.keychainEnabled {
		if err := keyring.Delete(KeyringService, KeyringUser); err != nil {
			// Ignore "not found" errors
			if !errors.Is(err, keyring.ErrNotFound) {
				lastErr = err
			}
		}
	}

	// Delete from config file
	if err := s.deleteFileToken(); err != nil {
		lastErr = err
	}

	return lastErr
}

// HasToken returns true if a token is available from any source.
func (s *Store) HasToken() bool {
	// Check environment
	if os.Getenv(EnvToken) != "" {
		return true
	}

	// Check keychain
	if s.keychainEnabled {
		token, err := keyring.Get(KeyringService, KeyringUser)
		if err == nil && token != "" {
			return true
		}
	}

	// Check file
	token, err := s.getFileToken()
	return err == nil && token != ""
}

// getFileToken reads the token from the config file.
func (s *Store) getFileToken() (string, error) {
	v := viper.New()
	v.SetConfigFile(s.configPath)

	if err := v.ReadInConfig(); err != nil {
		return "", err
	}

	return v.GetString("token"), nil
}

// setFileToken writes the token to the config file.
func (s *Store) setFileToken(token string) error {
	// Ensure directory exists
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Read existing config if present
	v := viper.New()
	v.SetConfigFile(s.configPath)
	_ = v.ReadInConfig() // Ignore error if file doesn't exist

	// Set the token
	v.Set("token", token)

	// Write with restricted permissions
	if err := v.WriteConfigAs(s.configPath); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	// Ensure correct permissions (viper may not set them)
	if err := os.Chmod(s.configPath, 0600); err != nil {
		return fmt.Errorf("setting config file permissions: %w", err)
	}

	return nil
}

// deleteFileToken removes the token from the config file.
func (s *Store) deleteFileToken() error {
	v := viper.New()
	v.SetConfigFile(s.configPath)

	if err := v.ReadInConfig(); err != nil {
		// File doesn't exist, nothing to delete
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Remove token key and rewrite
	v.Set("token", "")

	return v.WriteConfig()
}

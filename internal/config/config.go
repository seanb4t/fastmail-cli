// Package config provides configuration management using Viper.
//
// Config sources (in priority order, highest to lowest):
//  1. Environment variables (FASTMAIL_*)
//  2. Config file (~/.config/fastmail-cli/config.yaml)
//  3. Default values
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const defaultCardDAVEndpoint = "https://carddav.fastmail.com/dav/"

// Config holds the application configuration.
type Config struct {
	// Endpoint is the JMAP session endpoint URL.
	Endpoint string `mapstructure:"endpoint"`

	// OutputFormat controls output rendering: auto, json, or text.
	OutputFormat string `mapstructure:"output_format"`

	// Token is the API token for authentication.
	Token string `mapstructure:"token"`

	// CardDAVEndpoint is the CardDAV server URL for contacts.
	CardDAVEndpoint string `mapstructure:"carddav_endpoint"`

	// CardDAVUsername is the username for CardDAV authentication.
	CardDAVUsername string `mapstructure:"carddav_username"`
}

// DefaultConfigPath returns the default configuration file path.
func DefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}
	return filepath.Join(homeDir, ".config", "fastmail-cli", "config.yaml")
}

// Load reads configuration from the given path and environment variables.
// Environment variables take precedence over file values.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("endpoint", "https://api.fastmail.com/jmap/session")
	v.SetDefault("output_format", "auto")
	v.SetDefault("token", "")
	v.SetDefault("carddav_username", "")

	// Configure environment variable binding
	v.SetEnvPrefix("FASTMAIL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Try to read config file if it exists
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		// Ignore "file not found" errors - config file is optional
		var configFileNotFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundErr) {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("reading config file: %w", err)
			}
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if cfg.CardDAVEndpoint == "" {
		cfg.CardDAVEndpoint = deriveCardDAVEndpoint(cfg.Endpoint)
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func deriveCardDAVEndpoint(endpoint string) string {
	if endpoint == "" {
		return defaultCardDAVEndpoint
	}

	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return defaultCardDAVEndpoint
	}

	host := parsed.Host
	if strings.HasPrefix(host, "api.") {
		host = "carddav." + strings.TrimPrefix(host, "api.")
	}

	return fmt.Sprintf("%s://%s/dav/", parsed.Scheme, host)
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	validFormats := map[string]bool{
		"auto": true,
		"json": true,
		"text": true,
	}
	if !validFormats[c.OutputFormat] {
		return fmt.Errorf("output_format must be one of: auto, json, text (got %q)", c.OutputFormat)
	}

	return nil
}

# internal/config

Configuration management using viper.

## Purpose

Load and manage application configuration from:
- Configuration files (YAML, JSON)
- Environment variables (FASTMAIL_* prefix)
- Command-line flags

## Key Types

| Type | Description |
|------|-------------|
| `Config` | Root configuration structure |
| `Loader` | Configuration loading interface |

## Conventions

- Viper for configuration loading
- Environment variables override file config
- Validate config on load, fail fast on invalid
- Use struct tags for mapping

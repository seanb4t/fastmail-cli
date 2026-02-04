# internal/auth

Authentication and credential storage.

## Purpose

Manage Fastmail API credentials with secure storage:
- System keychain (preferred)
- File-based fallback for headless systems

## Key Types

| Type | Description |
|------|-------------|
| `Store` | Credential storage interface |
| `KeychainStore` | System keychain implementation |
| `FileStore` | File-based fallback implementation |

## Conventions

- Keychain preferred, file fallback for CI/headless
- Use go-keyring for cross-platform keychain access
- Never log or print credentials
- Validate tokens before storing

# internal/

Private packages for fastmail-cli implementation.

## Packages

| Package | Description |
|---------|-------------|
| `config/` | Configuration loading via viper |
| `auth/` | Authentication and credential storage |
| `jmap/` | JMAP protocol client implementation |
| `dav/` | CardDAV/CalDAV client implementation |
| `output/` | Output formatting (JSON, table, etc.) |
| `mocks/` | Generated mock implementations |

## Conventions

- Internal packages are not importable by external code
- Interfaces defined in consumer packages
- Implementations in separate packages
- Use dependency injection for testability
- Context as first parameter for cancellable operations

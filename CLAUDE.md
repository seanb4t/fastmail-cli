# fastmail-cli

Go CLI for Fastmail access and surfacing to Agentic Systems.

## Quick Reference

| Command | Description |
|---------|-------------|
| `task build` | Build the binary to `./bin/fastmail-cli` |
| `task test` | Run all tests with race detection |
| `task lint` | Run golangci-lint |
| `task mocks` | Generate mocks with mockery |
| `task all` | Lint, test, and build |

## Architecture

```
cmd/
  fastmail-cli/     # CLI entrypoint
internal/
  client/           # Fastmail JMAP client
  config/           # Configuration management
  commands/         # CLI command implementations
pkg/
  jmap/             # JMAP protocol types (if public)
```

## Development

### Prerequisites

- Go 1.22+
- [Task](https://taskfile.dev/) - task runner
- [golangci-lint](https://golangci-lint.run/) - linter
- [mockery](https://vektra.github.io/mockery/) - mock generator

### Build

```bash
task build
./bin/fastmail-cli --help
```

### Test

```bash
task test                 # Run tests
task test:coverage        # Generate coverage report
```

## Coding Standards

- Use `internal/` for private packages
- Interfaces in consumer packages, implementations separate
- Table-driven tests with descriptive names
- Error wrapping with `fmt.Errorf("context: %w", err)`
- Context as first parameter for cancellable operations

## Configuration

Configuration via environment variables and/or config file:

| Variable | Description |
|----------|-------------|
| `FASTMAIL_API_TOKEN` | Fastmail API token |
| `FASTMAIL_ACCOUNT_ID` | Account ID for JMAP requests |

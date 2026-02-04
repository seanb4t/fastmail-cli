# cmd/

Binary entrypoints for the fastmail-cli project.

## Packages

| Package | Description |
|---------|-------------|
| `fastmail-cli/` | Main CLI binary entrypoint |

## Conventions

- Each subdirectory is a separate binary
- `main.go` must be minimal - delegate to `cli.Execute()`
- No business logic in cmd packages
- Import only `cli` package and standard library

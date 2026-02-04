# cli/

Command-line interface implementation using cobra.

## Purpose

Define CLI commands and flags:
- Root command with global flags
- Subcommands for each feature area
- Help and completion generation

## Key Types

| Type | Description |
|------|-------------|
| `Execute()` | Run the root command |
| Root command | Global flags: --config, --json, --quiet |

## Conventions

- Use cobra for command definition
- Global flags on root, specific flags on subcommands
- RunE for commands that can error
- Short help under 80 chars, long help in separate files

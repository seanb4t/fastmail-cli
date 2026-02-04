# internal/output

Output formatting for CLI responses.

## Purpose

Format command output in various styles:
- JSON for machine consumption
- Table for human readability
- Quiet mode for scripting

## Key Types

| Type | Description |
|------|-------------|
| `Formatter` | Output formatting interface |
| `JSONFormatter` | JSON output implementation |
| `TableFormatter` | Human-readable table output |

## Conventions

- Honor --json and --quiet flags
- Errors to stderr, output to stdout
- Consistent field ordering in JSON
- No color when piped (isatty check)

# internal/jmap

JMAP protocol client implementation.

## Purpose

Handle JMAP API communication with Fastmail:
- Session management and capability discovery
- Request/response handling
- Method call batching

## Key Types

| Type | Description |
|------|-------------|
| `Client` | JMAP client interface |
| `Session` | Active JMAP session with capabilities |
| `Request` | JMAP request builder |
| `Response` | JMAP response parser |

## Conventions

- Use context for request cancellation
- Batch related method calls
- Handle back-references in requests
- Validate capabilities before method calls

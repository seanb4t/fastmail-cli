# Auth Status Token Validation Design

## Overview

Enhance `auth status` to validate tokens against the FastMail JMAP API instead of just checking for token existence. The command will make a network request to `/jmap/session` and report the actual authentication state.

## Exit Codes

| Code | Meaning | Text Output | JSON Output |
|------|---------|-------------|-------------|
| 0 | Authenticated | `Authenticated as user@fastmail.com` | `{"authenticated": true, "username": "..."}` |
| 1 | No token stored | `Not logged in` | `{"authenticated": false, "reason": "no_token"}` |
| 2 | Token invalid/revoked | `Token expired or revoked` | `{"authenticated": false, "reason": "invalid_token"}` |
| 3 | Network error | `Cannot reach FastMail API: <error>` | `{"authenticated": false, "reason": "network_error", "error": "..."}` |

## Implementation

### File Changes

**`cli/auth.go`**

Update `newAuthStatusCommand` to:

1. Check if a token exists via `store.GetToken()`
   - If no token → exit 1 with "Not logged in"

2. Create a JMAP client and call `Authenticate(ctx)`
   - Success → exit 0 with username from session
   - HTTP 401/403 → exit 2 with "Token expired or revoked"
   - Other network error → exit 3 with error details

### Flow

```go
token, err := store.GetToken()
if err != nil || token == "" {
    return ExitCodeNoToken
}

client := jmap.NewClient(JMAPEndpoint, token)
session, err := client.Authenticate(ctx)
if err != nil {
    if isAuthError(err) {
        return ExitCodeInvalidToken
    }
    return ExitCodeNetworkError
}

fmt.Printf("Authenticated as %s\n", session.Username)
return ExitCodeSuccess
```

### Dependencies

- Import `internal/jmap` to create a client
- JMAP endpoint: `https://api.fastmail.com/jmap/session` (constant or from config)

## Testing

### Unit Tests (`cli/auth_test.go`)

| Test Case | Setup | Expected |
|-----------|-------|----------|
| No token stored | Mock store returns empty | Exit 1, "Not logged in" |
| Valid token | Mock JMAP returns session | Exit 0, shows username |
| Invalid token | Mock JMAP returns 401 | Exit 2, "Token expired or revoked" |
| Network failure | Mock JMAP returns connection error | Exit 3, shows error |
| JSON output mode | Valid token + `--output json` | JSON with `authenticated: true` |

### Testability

Add `WithHTTPClient(*http.Client)` option to `jmap.NewClient` to allow injecting a mock HTTP client with custom `RoundTripper`.

## Files Modified

- `cli/auth.go` - update `newAuthStatusCommand`
- `cli/auth_test.go` - add test cases
- `internal/jmap/client.go` - add `WithHTTPClient` for testability

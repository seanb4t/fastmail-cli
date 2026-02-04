# Auth Status Token Validation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make `auth status` validate tokens against FastMail's JMAP API with distinct exit codes.

**Architecture:** The command will use the existing JMAP client to call `Authenticate()`. We'll add exit code constants and error classification to distinguish no-token, invalid-token, and network errors.

**Tech Stack:** Go, Cobra CLI, internal/jmap client, httptest for mocking

---

## Task 1: Add Exit Code Constants

**Files:**
- Create: `cli/exitcodes.go`

**Step 1: Create exit codes file**

```go
package cli

// Exit codes for CLI commands
const (
	ExitSuccess      = 0 // Successful operation
	ExitNoToken      = 1 // No token stored
	ExitInvalidToken = 2 // Token expired or revoked
	ExitNetworkError = 3 // Cannot reach API
)
```

**Step 2: Commit**

```bash
git add cli/exitcodes.go
git commit -m "feat(auth): add exit code constants for auth status"
```

---

## Task 2: Add HTTP Client Option to JMAP Client

**Files:**
- Modify: `internal/jmap/client.go:14-31`

**Step 1: Write failing test**

Create `internal/jmap/client_test.go` (add to existing):

```go
func TestClient_WithHTTPClient(t *testing.T) {
	called := false
	mockTransport := &mockRoundTripper{
		fn: func(req *http.Request) (*http.Response, error) {
			called = true
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"username":"test@example.com","apiUrl":"https://api.fastmail.com/jmap/api/"}`)),
			}, nil
		},
	}

	client := NewClient("https://api.fastmail.com/jmap/session", "test-token",
		WithHTTPClient(&http.Client{Transport: mockTransport}))

	_, err := client.Authenticate(context.Background())
	require.NoError(t, err)
	assert.True(t, called, "custom HTTP client should be used")
}

type mockRoundTripper struct {
	fn func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/jmap/... -run TestClient_WithHTTPClient -v`
Expected: FAIL - `WithHTTPClient` undefined

**Step 3: Implement WithHTTPClient option**

Modify `internal/jmap/client.go`:

```go
// ClientOption configures a Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new JMAP client for the given endpoint and access token.
func NewClient(endpoint, accessToken string, opts ...ClientOption) *Client {
	c := &Client{
		endpoint:    endpoint,
		accessToken: accessToken,
		httpClient:  &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/jmap/... -run TestClient_WithHTTPClient -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/jmap/client.go internal/jmap/client_test.go
git commit -m "feat(jmap): add WithHTTPClient option for testability"
```

---

## Task 3: Add JMAP Endpoint Constant

**Files:**
- Create: `internal/jmap/endpoints.go`

**Step 1: Create endpoints file**

```go
package jmap

// DefaultSessionURL is the FastMail JMAP session endpoint.
const DefaultSessionURL = "https://api.fastmail.com/jmap/session"
```

**Step 2: Commit**

```bash
git add internal/jmap/endpoints.go
git commit -m "feat(jmap): add default session URL constant"
```

---

## Task 4: Write Failing Tests for Auth Status Validation

**Files:**
- Modify: `cli/auth_test.go`

**Step 1: Add test for no token case (update existing)**

Update `TestAuthStatus_NoToken` to expect exit code 1:

```go
func TestAuthStatus_NoToken(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	_ = os.Unsetenv("FASTMAIL_TOKEN")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("FASTMAIL_TOKEN", originalEnv)
		}
	}()

	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--config", configPath, "auth", "status"})

	err := cmd.Execute()
	// Now expects an error for no token
	if err == nil {
		t.Fatal("auth status should error when no token")
	}

	output := buf.String()
	if !strings.Contains(output, "Not logged in") {
		t.Errorf("expected 'Not logged in' in output, got: %q", output)
	}
}
```

**Step 2: Add test for valid token**

```go
func TestAuthStatus_ValidToken(t *testing.T) {
	// This test requires mocking the JMAP client
	// For now, skip - will be implemented with integration test helper
	t.Skip("requires JMAP mock infrastructure")
}
```

**Step 3: Add test for invalid token**

```go
func TestAuthStatus_InvalidToken(t *testing.T) {
	t.Skip("requires JMAP mock infrastructure")
}
```

**Step 4: Add test for network error**

```go
func TestAuthStatus_NetworkError(t *testing.T) {
	t.Skip("requires JMAP mock infrastructure")
}
```

**Step 5: Run tests to verify they fail/skip**

Run: `go test ./cli/... -run TestAuthStatus -v`
Expected: NoToken test FAILS (behavior changed), others SKIP

**Step 6: Commit test scaffolding**

```bash
git add cli/auth_test.go
git commit -m "test(auth): add scaffolding for auth status validation tests"
```

---

## Task 5: Implement Auth Status Validation

**Files:**
- Modify: `cli/auth.go:111-145`

**Step 1: Update imports**

Add to imports in `cli/auth.go`:

```go
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"

	"github.com/seanb4t/fastmail-cli/internal/auth"
	"github.com/seanb4t/fastmail-cli/internal/config"
	"github.com/seanb4t/fastmail-cli/internal/jmap"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)
```

**Step 2: Create AuthStatusError type**

Add before `newAuthStatusCommand`:

```go
// AuthStatusError wraps an error with an exit code.
type AuthStatusError struct {
	Code    int
	Message string
}

func (e *AuthStatusError) Error() string {
	return e.Message
}
```

**Step 3: Rewrite newAuthStatusCommand**

Replace the entire function:

```go
// newAuthStatusCommand creates the auth status command.
func newAuthStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  "Validate your FastMail API token and show authentication status.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			configPath := GetConfigPath()
			if configPath == "" {
				configPath = config.DefaultConfigPath()
			}

			store := auth.NewStore(configPath)
			store.DisableKeychain()

			// Check if token exists
			token, err := store.GetToken()
			if err != nil || token == "" {
				if IsJSONOutput() {
					return outputJSON(cmd, map[string]any{
						"authenticated": false,
						"reason":        "no_token",
					})
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Not logged in")
				return &AuthStatusError{Code: ExitNoToken, Message: "no token stored"}
			}

			// Validate token against API
			client := jmap.NewClient(jmap.DefaultSessionURL, token)
			session, err := client.Authenticate(ctx)
			if err != nil {
				return handleAuthError(cmd, err)
			}

			// Success
			username := session.Username
			if IsJSONOutput() {
				return outputJSON(cmd, map[string]any{
					"authenticated": true,
					"username":      username,
				})
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Authenticated as %s\n", username)
			return nil
		},
	}
}

func handleAuthError(cmd *cobra.Command, err error) error {
	// Check for auth errors (401, 403)
	if isAuthError(err) {
		if IsJSONOutput() {
			_ = outputJSON(cmd, map[string]any{
				"authenticated": false,
				"reason":        "invalid_token",
			})
		} else {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Token expired or revoked")
		}
		return &AuthStatusError{Code: ExitInvalidToken, Message: "token invalid"}
	}

	// Network error
	if IsJSONOutput() {
		_ = outputJSON(cmd, map[string]any{
			"authenticated": false,
			"reason":        "network_error",
			"error":         err.Error(),
		})
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cannot reach FastMail API: %v\n", err)
	}
	return &AuthStatusError{Code: ExitNetworkError, Message: err.Error()}
}

func isAuthError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "401") || strings.Contains(errStr, "403")
}

func outputJSON(cmd *cobra.Command, data map[string]any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
```

**Step 4: Add strings import**

Ensure `strings` is in the imports.

**Step 5: Run tests**

Run: `go test ./cli/... -run TestAuthStatus -v`
Expected: NoToken test should PASS now

**Step 6: Commit**

```bash
git add cli/auth.go
git commit -m "feat(auth): validate token against FastMail API in auth status"
```

---

## Task 6: Handle Exit Codes in Root Command

**Files:**
- Modify: `cli/root.go` (or main.go if that's where Execute is called)

**Step 1: Find where Execute is called**

Look for `cmd.Execute()` in main.go or similar.

**Step 2: Add exit code handling**

```go
func Execute() {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		var authErr *AuthStatusError
		if errors.As(err, &authErr) {
			os.Exit(authErr.Code)
		}
		os.Exit(1)
	}
}
```

**Step 3: Run full test suite**

Run: `go test ./... -v`
Expected: All tests pass

**Step 4: Commit**

```bash
git add cli/root.go
git commit -m "feat(cli): handle auth status exit codes"
```

---

## Task 7: Update Documentation

**Files:**
- Modify: `docs/site/cli/auth.md`

**Step 1: Update auth status docs**

Add exit codes section:

```markdown
### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Authenticated successfully |
| 1 | No token stored |
| 2 | Token expired or revoked |
| 3 | Cannot reach FastMail API |
```

**Step 2: Update examples**

```markdown
### Examples

```bash
# Check authentication status
fm auth status
# Output: Authenticated as user@fastmail.com

# Check status in scripts
if fm auth status; then
  echo "Ready to use"
else
  echo "Please run: fm auth login"
fi
```

**Step 3: Commit**

```bash
git add docs/site/cli/auth.md
git commit -m "docs(auth): document auth status exit codes"
```

---

## Summary

| Task | Description | Commit |
|------|-------------|--------|
| 1 | Add exit code constants | `feat(auth): add exit code constants` |
| 2 | Add WithHTTPClient option | `feat(jmap): add WithHTTPClient option` |
| 3 | Add JMAP endpoint constant | `feat(jmap): add default session URL` |
| 4 | Add test scaffolding | `test(auth): add scaffolding for validation tests` |
| 5 | Implement validation | `feat(auth): validate token against FastMail API` |
| 6 | Handle exit codes | `feat(cli): handle auth status exit codes` |
| 7 | Update documentation | `docs(auth): document auth status exit codes` |

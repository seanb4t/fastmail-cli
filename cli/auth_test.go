package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockRoundTripper allows injecting custom HTTP responses for testing.
type mockRoundTripper struct {
	fn func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

// withMockJMAP sets up a mock HTTP client for auth status tests.
func withMockJMAP(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) func() {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(server.Close)

	// Create a mock transport that redirects to our test server
	mockTransport := &mockRoundTripper{
		fn: func(req *http.Request) (*http.Response, error) {
			// Forward to the test server
			client := server.Client()
			newReq, err := http.NewRequestWithContext(req.Context(), req.Method, server.URL+req.URL.Path, req.Body)
			if err != nil {
				return nil, err
			}
			newReq.Header = req.Header
			return client.Transport.RoundTrip(newReq)
		},
	}

	authStatusHTTPClient = &http.Client{Transport: mockTransport}
	return func() {
		authStatusHTTPClient = nil
	}
}

func TestAuthStatus_NoToken(t *testing.T) {
	// Setup: ensure no token exists
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Clear env var
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

	// Verify exit code
	var authErr *AuthStatusError
	if ok := errorAs(err, &authErr); !ok {
		t.Errorf("expected AuthStatusError, got: %T", err)
	} else if authErr.Code != ExitNoToken {
		t.Errorf("expected exit code %d, got: %d", ExitNoToken, authErr.Code)
	}
}

func TestAuthStatus_ValidToken(t *testing.T) {
	cleanup := withMockJMAP(t, func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("expected Bearer auth, got: %q", auth)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"username": "test@fastmail.com",
			"apiUrl":   "https://api.fastmail.com/jmap/api/",
		})
	})
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a token to the config
	err := os.WriteFile(configPath, []byte("token: valid-test-token\n"), 0600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Clear env var to use file token
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

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("auth status should not error with valid token: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Authenticated as test@fastmail.com") {
		t.Errorf("expected 'Authenticated as test@fastmail.com' in output, got: %q", output)
	}

	// Should NOT show the actual token
	if strings.Contains(output, "valid-test-token") {
		t.Error("auth status should not display the actual token")
	}
}

func TestAuthStatus_InvalidToken(t *testing.T) {
	cleanup := withMockJMAP(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "unauthorized"}`))
	})
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write an invalid token to the config
	err := os.WriteFile(configPath, []byte("token: invalid-token\n"), 0600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Clear env var
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

	err = cmd.Execute()
	if err == nil {
		t.Fatal("auth status should error with invalid token")
	}

	output := buf.String()
	if !strings.Contains(output, "Token expired or revoked") {
		t.Errorf("expected 'Token expired or revoked' in output, got: %q", output)
	}

	// Verify exit code
	var authErr *AuthStatusError
	if ok := errorAs(err, &authErr); !ok {
		t.Errorf("expected AuthStatusError, got: %T", err)
	} else if authErr.Code != ExitInvalidToken {
		t.Errorf("expected exit code %d, got: %d", ExitInvalidToken, authErr.Code)
	}
}

func TestAuthStatus_NetworkError(t *testing.T) {
	// Use a mock that returns network error
	mockTransport := &mockRoundTripper{
		fn: func(_ *http.Request) (*http.Response, error) {
			return nil, &mockNetError{msg: "connection refused"}
		},
	}
	authStatusHTTPClient = &http.Client{Transport: mockTransport}
	defer func() { authStatusHTTPClient = nil }()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a token to the config
	err := os.WriteFile(configPath, []byte("token: some-token\n"), 0600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Clear env var
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

	err = cmd.Execute()
	if err == nil {
		t.Fatal("auth status should error on network failure")
	}

	output := buf.String()
	if !strings.Contains(output, "Cannot reach FastMail API") {
		t.Errorf("expected 'Cannot reach FastMail API' in output, got: %q", output)
	}

	// Verify exit code
	var authErr *AuthStatusError
	if ok := errorAs(err, &authErr); !ok {
		t.Errorf("expected AuthStatusError, got: %T", err)
	} else if authErr.Code != ExitNetworkError {
		t.Errorf("expected exit code %d, got: %d", ExitNetworkError, authErr.Code)
	}
}

// mockNetError implements net.Error for testing network failures.
type mockNetError struct {
	msg string
}

func (e *mockNetError) Error() string   { return e.msg }
func (e *mockNetError) Timeout() bool   { return false }
func (e *mockNetError) Temporary() bool { return false }

func TestAuthStatus_ValidToken_JSON(t *testing.T) {
	cleanup := withMockJMAP(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"username": "user@fastmail.com",
			"apiUrl":   "https://api.fastmail.com/jmap/api/",
		})
	})
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte("token: test-token\n"), 0600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

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
	cmd.SetArgs([]string{"--config", configPath, "--json", "auth", "status"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("auth status --json should not error with valid token: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"authenticated"`) {
		t.Errorf("JSON output should contain 'authenticated' field, got: %q", output)
	}
	if !strings.Contains(output, `"username"`) {
		t.Errorf("JSON output should contain 'username' field, got: %q", output)
	}
	if !strings.Contains(output, "true") {
		t.Errorf("JSON output should show authenticated: true, got: %q", output)
	}
}

func TestAuthStatus_NoToken_JSON(t *testing.T) {
	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	_ = os.Unsetenv("FASTMAIL_TOKEN")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("FASTMAIL_TOKEN", originalEnv)
		}
	}()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--config", configPath, "--json", "auth", "status"})

	err := cmd.Execute()
	// Should error even with JSON output
	if err == nil {
		t.Fatal("auth status --json should error when no token")
	}

	output := buf.String()
	if !strings.Contains(output, `"authenticated"`) {
		t.Errorf("JSON output should contain 'authenticated' field, got: %q", output)
	}
	if !strings.Contains(output, `"reason"`) {
		t.Errorf("JSON output should contain 'reason' field, got: %q", output)
	}
	if !strings.Contains(output, "no_token") {
		t.Errorf("JSON output should show reason: no_token, got: %q", output)
	}
}

func TestAuthHelp_ShowsSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"auth", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth help should not error: %v", err)
	}

	output := buf.String()
	// Should show all three subcommands
	if !strings.Contains(output, "login") {
		t.Errorf("auth help should show 'login' subcommand, got: %q", output)
	}
	if !strings.Contains(output, "logout") {
		t.Errorf("auth help should show 'logout' subcommand, got: %q", output)
	}
	if !strings.Contains(output, "status") {
		t.Errorf("auth help should show 'status' subcommand, got: %q", output)
	}
}

func TestAuthLogout_RemovesToken(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a config file with a token
	err := os.WriteFile(configPath, []byte("token: test-token\n"), 0600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Clear env var
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
	cmd.SetArgs([]string{"--config", configPath, "auth", "logout"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("auth logout should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Logged out") {
		t.Errorf("expected 'Logged out' in output, got: %q", output)
	}

	// Verify token was removed - run status to check (expect error now)
	cmd2 := NewRootCommand()
	buf2 := new(bytes.Buffer)
	cmd2.SetOut(buf2)
	cmd2.SetErr(buf2)
	cmd2.SetArgs([]string{"--config", configPath, "auth", "status"})

	_ = cmd2.Execute()
	if !strings.Contains(buf2.String(), "Not logged in") {
		t.Errorf("after logout, status should show 'Not logged in', got: %q", buf2.String())
	}
}

func TestAuthLogin_StoresToken(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Clear env var
	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	_ = os.Unsetenv("FASTMAIL_TOKEN")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("FASTMAIL_TOKEN", originalEnv)
		}
	}()

	// Test with --token flag (non-interactive mode)
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--config", configPath, "auth", "login", "--token", "my-api-token"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth login should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Logged in") {
		t.Errorf("expected 'Logged in' in output, got: %q", output)
	}

	// Verify token was stored by reading the config file directly
	// (not by running status, which would try to validate against API)
	configContent, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if !strings.Contains(string(configContent), "token:") {
		t.Errorf("config file should contain token, got: %q", string(configContent))
	}
}

func TestAuthLogin_RequiresToken(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// When not a terminal and no --token flag, should error
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--config", configPath, "auth", "login"})

	err := cmd.Execute()
	// Should error because no token provided and not interactive
	if err == nil {
		t.Error("auth login without token should error in non-interactive mode")
	}
}

// errorAs is a helper to check error type (like errors.As but simpler for tests).
func errorAs(err error, target any) bool {
	if err == nil {
		return false
	}
	// Use errors.As for proper error unwrapping.
	if t, ok := target.(**AuthStatusError); ok {
		return errors.As(err, t)
	}
	return false
}

// Ensure exported httpClient is set and reset properly.
func init() {
	// Set up a way to inject mock HTTP client for testing
	// The authStatusHTTPClient variable in auth.go handles this
	_ = io.Discard // Use io to prevent unused import error
}

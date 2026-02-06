package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"

	"github.com/seanb4t/fastmail-cli/internal/auth"
	"github.com/seanb4t/fastmail-cli/internal/jmap"
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

func mockKeyringError(t *testing.T, err error) {
	t.Helper()
	keyring.MockInitWithError(err)
	t.Cleanup(keyring.MockInit)
}

// withMockJMAPLogin sets up a mock HTTP client for auth login tests.
func withMockJMAPLogin(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) func() {
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

	authLoginHTTPClient = &http.Client{Transport: mockTransport}
	return func() {
		authLoginHTTPClient = nil
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

func TestAuthStatus_UsesConfiguredEndpoint(t *testing.T) {
	customPath := "/custom/session"
	cleanup := withMockJMAP(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != customPath {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
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
	endpoint := "http://example.test" + customPath

	configContent := strings.Join([]string{
		"token: valid-test-token",
		"endpoint: " + endpoint,
		"",
	}, "\n")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
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
	cmd.SetArgs([]string{"--config", configPath, "auth", "status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth status should not error with configured endpoint: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Authenticated as test@fastmail.com") {
		t.Errorf("expected auth success output, got: %q", output)
	}
}

func TestAuthStatus_LogsResolvedEndpoint(t *testing.T) {
	customPath := "/debug/session"
	cleanup := withMockJMAP(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != customPath {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
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
	endpoint := "http://example.test" + customPath

	configContent := strings.Join([]string{
		"token: valid-test-token",
		"endpoint: " + endpoint,
		"",
	}, "\n")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
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
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"--config", configPath, "auth", "status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth status should not error with configured endpoint: %v", err)
	}

	if !strings.Contains(errBuf.String(), endpoint) {
		t.Errorf("expected endpoint in stderr, got: %q", errBuf.String())
	}
}

func TestAuthStatus_DefaultEndpointResolved(t *testing.T) {
	cleanup := withMockJMAP(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jmap/session" {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
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

	configContent := "token: default-endpoint-token" + "\n" + ""
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
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
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"--config", configPath, "auth", "status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth status should not error with default endpoint: %v", err)
	}

	stderr := errBuf.String()
	if !strings.Contains(stderr, "Resolved JMAP endpoint") {
		t.Errorf("expected resolved endpoint log in stderr, got: %q", stderr)
	}
	if !strings.Contains(stderr, "https://api.fastmail.com/jmap/session") {
		t.Errorf("expected default endpoint in stderr, got: %q", stderr)
	}
}

func TestAuthStatus_ConfigLoadFailureUsesDefaultEndpoint(t *testing.T) {
	cleanup := withMockJMAP(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jmap/session" {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
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

	configContent := strings.Join([]string{
		"token: default-endpoint-token",
		"output_format: bogus",
		"",
	}, "\n")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
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
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"--config", configPath, "auth", "status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth status should not error with invalid config: %v", err)
	}

	if !strings.Contains(outBuf.String(), "Authenticated as test@fastmail.com") {
		t.Errorf("expected auth success output, got: %q", outBuf.String())
	}

	stderr := errBuf.String()
	if !strings.Contains(stderr, "Warning: failed to load config") {
		t.Errorf("expected config warning in stderr, got: %q", stderr)
	}
	if !strings.Contains(stderr, "Resolved JMAP endpoint") {
		t.Errorf("expected resolved endpoint log in stderr, got: %q", stderr)
	}
	if !strings.Contains(stderr, jmap.DefaultSessionURL) {
		t.Errorf("expected default endpoint in stderr, got: %q", stderr)
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

func TestAuthStatus_AuthError(t *testing.T) {
	cleanup := withMockJMAP(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "server_error"}`))
	})
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a token to the config
	err := os.WriteFile(configPath, []byte("token: server-error-token\n"), 0600)
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
		t.Fatal("auth status should error on auth failure")
	}

	output := buf.String()
	if !strings.Contains(output, "Authentication failed") {
		t.Errorf("expected 'Authentication failed' in output, got: %q", output)
	}

	// Verify exit code
	var authErr *AuthStatusError
	if ok := errorAs(err, &authErr); !ok {
		t.Errorf("expected AuthStatusError, got: %T", err)
	} else if authErr.Code != ExitAuthError {
		t.Errorf("expected exit code %d, got: %d", ExitAuthError, authErr.Code)
	}
}

func TestAuthStatus_InvalidJSONResponse(t *testing.T) {
	cleanup := withMockJMAP(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{not json"))
	})
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a token to the config
	err := os.WriteFile(configPath, []byte("token: invalid-json-token\n"), 0600)
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
		t.Fatal("auth status should error on invalid JSON response")
	}

	output := buf.String()
	if !strings.Contains(output, "Authentication failed") {
		t.Errorf("expected 'Authentication failed' in output, got: %q", output)
	}

	// Verify exit code
	var authErr *AuthStatusError
	if ok := errorAs(err, &authErr); !ok {
		t.Errorf("expected AuthStatusError, got: %T", err)
	} else if authErr.Code != ExitAuthError {
		t.Errorf("expected exit code %d, got: %d", ExitAuthError, authErr.Code)
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

func TestAuthStatus_AuthError_JSON(t *testing.T) {
	cleanup := withMockJMAP(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "server_error"}`))
	})
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte("token: json-error-token\n"), 0600)
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
	if err == nil {
		t.Fatal("auth status --json should error on auth failure")
	}

	output := buf.String()
	if !strings.Contains(output, `"reason"`) {
		t.Errorf("JSON output should contain 'reason' field, got: %q", output)
	}
	if !strings.Contains(output, "auth_error") {
		t.Errorf("JSON output should show reason: auth_error, got: %q", output)
	}

	// Verify exit code
	var authErr *AuthStatusError
	if ok := errorAs(err, &authErr); !ok {
		t.Errorf("expected AuthStatusError, got: %T", err)
	} else if authErr.Code != ExitAuthError {
		t.Errorf("expected exit code %d, got: %d", ExitAuthError, authErr.Code)
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

func TestAuthLogout_ErrorsOnKeychainDeleteFailure(t *testing.T) {
	mockKeyringError(t, errors.New("keychain unavailable"))

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
	if err == nil {
		t.Fatalf("auth logout should return error when keychain delete fails")
	}
	if !strings.Contains(err.Error(), "removing token") {
		t.Fatalf("expected logout error to mention removing token, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Warning: keychain unavailable") || !strings.Contains(output, "falling back to config file") {
		t.Errorf("expected keychain fallback warning in output, got: %q", output)
	}

	configContent, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if strings.Contains(string(configContent), "test-token") {
		t.Errorf("expected token removed from config file, got: %q", string(configContent))
	}
}

func TestAuthLogin_InvalidTokenDoesNotStore(t *testing.T) {
	mockKeyringError(t, errors.New("keychain unavailable"))

	cleanup := withMockJMAPLogin(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "unauthorized"}`))
	})
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	endpoint := "http://example.test/jmap/session"

	configContent := "endpoint: " + endpoint + "\n" + ""
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
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
	cmd.SetArgs([]string{"--config", configPath, "auth", "login", "--token", "bad-token"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("auth login should error on invalid token")
	}
	if !strings.Contains(err.Error(), "invalid token") {
		t.Errorf("expected invalid token error, got: %v", err)
	}

	updatedConfig, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if strings.Contains(string(updatedConfig), "token:") {
		t.Errorf("token should not be stored on invalid login, got: %q", string(updatedConfig))
	}
}

func TestAuthLogin_ValidTokenAuthenticatesAndStores(t *testing.T) {
	mockKeyringError(t, errors.New("keychain unavailable"))

	loginCleanup := withMockJMAPLogin(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"username": "test@fastmail.com",
			"apiUrl":   "https://api.fastmail.com/jmap/api/",
		})
	})
	defer loginCleanup()

	statusCleanup := withMockJMAP(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"username": "test@fastmail.com",
			"apiUrl":   "https://api.fastmail.com/jmap/api/",
		})
	})
	defer statusCleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	endpoint := "http://example.test/jmap/session"

	configContent := "endpoint: " + endpoint + "\n" + ""
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
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
	cmd.SetArgs([]string{"--config", configPath, "auth", "login", "--token", "good-token"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth login should not error: %v", err)
	}

	updatedConfig, err := os.ReadFile(filepath.Clean(configPath))
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if !strings.Contains(string(updatedConfig), "token: good-token") {
		t.Errorf("expected token stored in config, got: %q", string(updatedConfig))
	}

	statusCmd := NewRootCommand()
	statusBuf := new(bytes.Buffer)
	statusCmd.SetOut(statusBuf)
	statusCmd.SetErr(statusBuf)
	statusCmd.SetArgs([]string{"--config", configPath, "auth", "status"})

	if err := statusCmd.Execute(); err != nil {
		t.Fatalf("auth status should not error after login: %v", err)
	}
	if !strings.Contains(statusBuf.String(), "Authenticated as test@fastmail.com") {
		t.Errorf("expected auth status success, got: %q", statusBuf.String())
	}
}

func TestAuthLogin_StoresTokenInKeychainByDefault(t *testing.T) {
	keyring.MockInit()
	t.Cleanup(func() {
		_ = keyring.DeleteAll(auth.KeyringService)
	})

	cleanup := withMockJMAPLogin(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"username": "test@fastmail.com",
			"apiUrl":   "https://api.fastmail.com/jmap/api/",
		})
	})
	defer cleanup()

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
	cmd.SetArgs([]string{"--config", configPath, "auth", "login", "--token", "my-api-token"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth login should not error: %v", err)
	}

	token, err := keyring.Get(auth.KeyringService, auth.KeyringUser)
	if err != nil {
		t.Fatalf("expected token in keychain, got error: %v", err)
	}
	if token != "my-api-token" {
		t.Errorf("expected keychain token %q, got: %q", "my-api-token", token)
	}

	if _, err := os.Stat(configPath); err == nil {
		t.Errorf("config file should not be created when keychain succeeds")
	} else if !os.IsNotExist(err) {
		t.Fatalf("unexpected config stat error: %v", err)
	}
}

func TestAuthLogin_StoresToken(t *testing.T) {
	mockKeyringError(t, errors.New("keychain unavailable"))

	cleanup := withMockJMAPLogin(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"username": "test@fastmail.com",
			"apiUrl":   "https://api.fastmail.com/jmap/api/",
		})
	})
	defer cleanup()

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
	if !strings.Contains(output, "Warning: keychain unavailable") || !strings.Contains(output, "falling back to config file") {
		t.Errorf("expected keychain fallback warning in output, got: %q", output)
	}
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

func TestAuthLogin_QuietSuppressesKeychainWarning(t *testing.T) {
	mockKeyringError(t, errors.New("keychain unavailable"))

	cleanup := withMockJMAPLogin(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"username": "test@fastmail.com",
			"apiUrl":   "https://api.fastmail.com/jmap/api/",
		})
	})
	defer cleanup()

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
	cmd.SetArgs([]string{"--config", configPath, "--quiet", "auth", "login", "--token", "my-api-token"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth login should not error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Warning: keychain unavailable") || strings.Contains(output, "falling back to config file") {
		t.Errorf("expected no keychain fallback warning in quiet mode, got: %q", output)
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
	keyring.MockInit()
}

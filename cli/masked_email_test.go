package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// maskedEmailSessionResponse returns a valid JMAP session with MaskedEmail capability.
func maskedEmailSessionResponse(apiURL string) string {
	return `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {},
			"urn:ietf:params:jmap:mail": {},
			"https://www.fastmail.com/dev/maskedemail": {}
		},
		"accounts": {
			"acc1": {
				"name": "test@example.com",
				"isPersonal": true,
				"isReadOnly": false,
				"accountCapabilities": {
					"urn:ietf:params:jmap:mail": {},
					"https://www.fastmail.com/dev/maskedemail": {}
				}
			}
		},
		"primaryAccounts": {"urn:ietf:params:jmap:mail": "acc1"},
		"username": "test@example.com",
		"apiUrl": "` + apiURL + `",
		"downloadUrl": "https://example.com/download",
		"uploadUrl": "https://example.com/upload",
		"eventSourceUrl": "https://example.com/events",
		"state": "s1"
	}`
}

func setupMaskedEmailTestEnv(t *testing.T, sessionURL string) (string, func()) {
	t.Helper()

	// Create temp directory for config
	tempDir, err := os.MkdirTemp("", "fastmail-cli-test-*")
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `endpoint: "` + sessionURL + `"`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set token via env var
	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	os.Setenv("FASTMAIL_TOKEN", "test-token")

	cleanup := func() {
		os.RemoveAll(tempDir)
		if originalEnv != "" {
			os.Setenv("FASTMAIL_TOKEN", originalEnv)
		} else {
			os.Unsetenv("FASTMAIL_TOKEN")
		}
	}

	return configPath, cleanup
}

func TestMaskedEmailListCommand(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["MaskedEmail/get", {
					"accountId": "acc1",
					"state": "me1",
					"list": [
						{
							"id": "me-1",
							"email": "abc123@fastmail.com",
							"state": "enabled",
							"forDomain": "example.com",
							"description": "Shopping site",
							"createdAt": "2024-01-15T10:30:00Z"
						},
						{
							"id": "me-2",
							"email": "xyz789@fastmail.com",
							"state": "disabled",
							"forDomain": "newsletter.com",
							"description": "Newsletter",
							"createdAt": "2024-01-10T08:00:00Z"
						}
					],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(maskedEmailSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupMaskedEmailTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "masked-email", "list"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "abc123@fastmail.com")
	assert.Contains(t, output, "xyz789@fastmail.com")
	assert.Contains(t, output, "enabled")
	assert.Contains(t, output, "disabled")
}

func TestMaskedEmailListCommand_JSON(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["MaskedEmail/get", {
					"accountId": "acc1",
					"state": "me1",
					"list": [
						{
							"id": "me-1",
							"email": "test@fastmail.com",
							"state": "enabled",
							"forDomain": "example.com"
						}
					],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(maskedEmailSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupMaskedEmailTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "--json", "masked-email", "list"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify it's valid JSON
	var result []map[string]any
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "test@fastmail.com", result[0]["email"])
}

func TestMaskedEmailCreateCommand(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)

		// Verify create payload has the domain
		create := args["create"].(map[string]any)
		newME := create["new"].(map[string]any)
		assert.Equal(t, "shop.example.com", newME["forDomain"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["MaskedEmail/set", {
					"accountId": "acc1",
					"oldState": "me1",
					"newState": "me2",
					"created": {
						"new": {
							"id": "me-new",
							"email": "newaddr@fastmail.com",
							"state": "enabled",
							"forDomain": "shop.example.com"
						}
					}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(maskedEmailSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupMaskedEmailTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "masked-email", "create", "--for-domain", "shop.example.com"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "newaddr@fastmail.com")
}

func TestMaskedEmailEnableCommand(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)

		// Verify update payload
		update := args["update"].(map[string]any)
		meUpdate := update["me-123"].(map[string]any)
		assert.Equal(t, "enabled", meUpdate["state"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["MaskedEmail/set", {
					"accountId": "acc1",
					"oldState": "me1",
					"newState": "me2",
					"updated": {"me-123": null}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(maskedEmailSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupMaskedEmailTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "masked-email", "enable", "me-123"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Enabled")
}

func TestMaskedEmailDisableCommand(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)

		// Verify update payload
		update := args["update"].(map[string]any)
		meUpdate := update["me-456"].(map[string]any)
		assert.Equal(t, "disabled", meUpdate["state"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["MaskedEmail/set", {
					"accountId": "acc1",
					"oldState": "me1",
					"newState": "me2",
					"updated": {"me-456": null}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(maskedEmailSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupMaskedEmailTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "masked-email", "disable", "me-456"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Disabled")
}

func TestMaskedEmailDeleteCommand(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)

		// Verify destroy payload
		destroy := args["destroy"].([]any)
		assert.Contains(t, destroy, "me-to-delete")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["MaskedEmail/set", {
					"accountId": "acc1",
					"oldState": "me1",
					"newState": "me2",
					"destroyed": ["me-to-delete"]
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(maskedEmailSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupMaskedEmailTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "masked-email", "delete", "me-to-delete"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Deleted")
}

func TestMaskedEmailEnableCommand_RequiresID(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"masked-email", "enable"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestMaskedEmailDisableCommand_RequiresID(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"masked-email", "disable"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestMaskedEmailDeleteCommand_RequiresID(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"masked-email", "delete"})

	err := cmd.Execute()
	assert.Error(t, err)
}

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

// identitySessionResponse returns a valid JMAP session with submission capability.
func identitySessionResponse(apiURL string) string {
	return `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {},
			"urn:ietf:params:jmap:mail": {},
			"urn:ietf:params:jmap:submission": {}
		},
		"accounts": {
			"acc1": {
				"name": "test@example.com",
				"isPersonal": true,
				"isReadOnly": false,
				"accountCapabilities": {
					"urn:ietf:params:jmap:mail": {}
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

func setupIdentityTestEnv(t *testing.T, sessionURL string) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "fastmail-cli-identity-test-*")
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `endpoint: "` + sessionURL + `"`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	_ = os.Setenv("FASTMAIL_TOKEN", "test-token")

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
		if originalEnv != "" {
			_ = os.Setenv("FASTMAIL_TOKEN", originalEnv)
		} else {
			_ = os.Unsetenv("FASTMAIL_TOKEN")
		}
	}

	return configPath, cleanup
}

func TestIdentityCommand_HasSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"identity", "--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "list")
	assert.Contains(t, output, "set")
}

func TestIdentityListCommand(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Identity/get", {
					"accountId": "acc1",
					"state": "i1",
					"list": [
						{
							"id": "id-1",
							"name": "Test User",
							"email": "test@example.com",
							"textSignature": "-- \nSent from CLI",
							"mayDelete": false
						},
						{
							"id": "id-2",
							"name": "Alt Identity",
							"email": "alt@example.com",
							"mayDelete": true
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
		_, _ = w.Write([]byte(identitySessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupIdentityTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "identity", "list"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "id-1")
	assert.Contains(t, output, "Test User")
	assert.Contains(t, output, "test@example.com")
	assert.Contains(t, output, "id-2")
	assert.Contains(t, output, "Alt Identity")
}

func TestIdentityListCommand_JSON(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Identity/get", {
					"accountId": "acc1",
					"state": "i1",
					"list": [
						{
							"id": "id-1",
							"name": "Test User",
							"email": "test@example.com",
							"replyTo": [{"name": "Reply", "email": "reply@example.com"}],
							"textSignature": "-- \nSig",
							"htmlSignature": "<p>Sig</p>",
							"mayDelete": false
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
		_, _ = w.Write([]byte(identitySessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupIdentityTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "--json", "identity", "list"})

	err := cmd.Execute()
	require.NoError(t, err)

	var result []map[string]any
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "id-1", result[0]["id"])
	assert.Equal(t, "Test User", result[0]["name"])
	assert.Equal(t, "test@example.com", result[0]["email"])
	assert.Equal(t, "-- \nSig", result[0]["text_signature"])
}

func TestIdentityListCommand_Quiet(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Identity/get", {
					"accountId": "acc1",
					"state": "i1",
					"list": [{"id": "id-1", "name": "Test", "email": "t@e.com"}],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(identitySessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupIdentityTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "--quiet", "identity", "list"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Empty(t, out.String())
}

func TestIdentitySetCommand(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)

		update := args["update"].(map[string]any)
		idUpdate := update["id-1"].(map[string]any)
		assert.Equal(t, "New Name", idUpdate["name"])
		assert.Equal(t, "New sig", idUpdate["textSignature"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Identity/set", {
					"accountId": "acc1",
					"oldState": "i1",
					"newState": "i2",
					"updated": {"id-1": null}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(identitySessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupIdentityTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "identity", "set", "id-1", "--name", "New Name", "--signature", "New sig"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Identity updated")
}

func TestIdentitySetCommand_JSON(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Identity/set", {
					"accountId": "acc1",
					"oldState": "i1",
					"newState": "i2",
					"updated": {"id-1": null}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(identitySessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupIdentityTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "--json", "identity", "set", "id-1", "--name", "New Name"})

	err := cmd.Execute()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "id-1", result["id"])
	assert.Equal(t, "updated", result["status"])
}

func TestIdentitySetCommand_RequiresID(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"identity", "set"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestIdentitySetCommand_RequiresAtLeastOneFlag(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Identity/set", {
					"accountId": "acc1",
					"oldState": "i1",
					"newState": "i1"
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(identitySessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupIdentityTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "identity", "set", "id-1"})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one")
}

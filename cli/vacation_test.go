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

// vacationSessionResponse returns a valid JMAP session with mail capability.
func vacationSessionResponse(apiURL string) string {
	return `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {},
			"urn:ietf:params:jmap:mail": {}
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

func setupVacationTestEnv(t *testing.T, sessionURL string) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "fastmail-cli-vacation-test-*")
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

func TestVacationCommand_HasSubcommands(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"vacation", "--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "status")
	assert.Contains(t, output, "enable")
	assert.Contains(t, output, "disable")
}

func TestVacationStatusCommand_Enabled(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/get", {
					"accountId": "acc1",
					"state": "v1",
					"list": [{
						"id": "singleton",
						"isEnabled": true,
						"fromDate": "2024-01-15T00:00:00Z",
						"toDate": "2024-01-20T00:00:00Z",
						"subject": "Out of Office",
						"textBody": "I'm away until Jan 20"
					}],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(vacationSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupVacationTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "vacation", "status"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "ENABLED")
	assert.Contains(t, output, "Out of Office")
	assert.Contains(t, output, "I'm away until Jan 20")
}

func TestVacationStatusCommand_Disabled(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/get", {
					"accountId": "acc1",
					"state": "v1",
					"list": [{
						"id": "singleton",
						"isEnabled": false
					}],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(vacationSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupVacationTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "vacation", "status"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "DISABLED")
}

func TestVacationStatusCommand_JSON(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/get", {
					"accountId": "acc1",
					"state": "v1",
					"list": [{
						"id": "singleton",
						"isEnabled": true,
						"subject": "OOO",
						"textBody": "Away"
					}],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(vacationSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupVacationTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "--json", "vacation", "status"})

	err := cmd.Execute()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(out.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, true, result["is_enabled"])
	assert.Equal(t, "OOO", result["subject"])
}

func TestVacationEnableCommand(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)

		update := args["update"].(map[string]any)
		singletonUpdate := update["singleton"].(map[string]any)
		assert.Equal(t, true, singletonUpdate["isEnabled"])
		assert.Equal(t, "Out of Office", singletonUpdate["subject"])
		assert.Equal(t, "I'm away", singletonUpdate["textBody"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/set", {
					"accountId": "acc1",
					"oldState": "v1",
					"newState": "v2",
					"updated": {"singleton": null}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(vacationSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupVacationTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "vacation", "enable", "--subject", "Out of Office", "--body", "I'm away"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Vacation response enabled")
}

func TestVacationEnableCommand_RequiresSubject(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"vacation", "enable", "--body", "I'm away"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestVacationEnableCommand_RequiresBody(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"vacation", "enable", "--subject", "OOO"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestVacationDisableCommand(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)

		update := args["update"].(map[string]any)
		singletonUpdate := update["singleton"].(map[string]any)
		assert.Equal(t, false, singletonUpdate["isEnabled"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["VacationResponse/set", {
					"accountId": "acc1",
					"oldState": "v1",
					"newState": "v2",
					"updated": {"singleton": null}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(vacationSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	configPath, cleanup := setupVacationTestEnv(t, sessionServer.URL)
	defer cleanup()

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "vacation", "disable"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Vacation response disabled")
}

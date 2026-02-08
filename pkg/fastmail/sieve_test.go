package fastmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sieveSessionResponse returns a valid JMAP session with SieveScript capability.
func sieveSessionResponse(apiURL string) string {
	return `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {},
			"urn:ietf:params:jmap:mail": {},
			"urn:ietf:params:jmap:sieve": {}
		},
		"accounts": {
			"acc1": {
				"name": "test@example.com",
				"isPersonal": true,
				"isReadOnly": false,
				"accountCapabilities": {
					"urn:ietf:params:jmap:mail": {},
					"urn:ietf:params:jmap:sieve": {}
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

func TestSieveService_List(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		assert.Equal(t, "SieveScript/get", methodName)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["SieveScript/get", {
					"accountId": "acc1",
					"state": "sv1",
					"list": [
						{
							"id": "sieve-1",
							"name": "Main Filter",
							"blobId": "blob-1",
							"isActive": true,
							"createdAt": "2024-01-15T10:30:00Z"
						},
						{
							"id": "sieve-2",
							"name": "Backup Filter",
							"blobId": "blob-2",
							"isActive": false,
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
		_, _ = w.Write([]byte(sieveSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	scripts, err := client.Sieve().List(ctx)
	require.NoError(t, err)
	require.Len(t, scripts, 2)

	assert.Equal(t, "sieve-1", scripts[0].ID)
	assert.Equal(t, "Main Filter", scripts[0].Name)
	assert.True(t, scripts[0].IsActive)

	assert.Equal(t, "sieve-2", scripts[1].ID)
	assert.False(t, scripts[1].IsActive)
}

func TestSieveService_Get(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)

		ids := args["ids"].([]any)
		assert.Contains(t, ids, "sieve-1")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["SieveScript/get", {
					"accountId": "acc1",
					"state": "sv1",
					"list": [
						{
							"id": "sieve-1",
							"name": "Main Filter",
							"blobId": "blob-1",
							"script": "require \"fileinto\"; if header :is \"from\" \"test@example.com\" { fileinto \"Test\"; }",
							"isActive": true
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
		_, _ = w.Write([]byte(sieveSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	script, err := client.Sieve().Get(ctx, "sieve-1")
	require.NoError(t, err)
	require.NotNil(t, script)

	assert.Equal(t, "sieve-1", script.ID)
	assert.Equal(t, "Main Filter", script.Name)
	assert.Contains(t, script.Script, "fileinto")
	assert.True(t, script.IsActive)
}

func TestSieveService_Create(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "SieveScript/set", methodName)

		create := args["create"].(map[string]any)
		newScript := create["new"].(map[string]any)
		assert.Equal(t, "My Filter", newScript["name"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["SieveScript/set", {
					"accountId": "acc1",
					"oldState": "sv1",
					"newState": "sv2",
					"created": {
						"new": {
							"id": "sieve-new",
							"name": "My Filter",
							"blobId": "blob-new",
							"isActive": false,
							"createdAt": "2024-01-25T12:00:00Z"
						}
					}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sieveSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	opts := CreateSieveScriptOptions{
		Name:   "My Filter",
		Script: `require "fileinto";`,
	}

	script, err := client.Sieve().Create(ctx, opts)
	require.NoError(t, err)
	require.NotNil(t, script)

	assert.Equal(t, "sieve-new", script.ID)
	assert.Equal(t, "My Filter", script.Name)
}

// runSieveActiveChangeTest is a shared helper for Activate/Deactivate service tests.
func runSieveActiveChangeTest(t *testing.T, scriptID string, expectedActive bool, callService func(context.Context, *Client, string) error) {
	t.Helper()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "SieveScript/set", methodName)

		update := args["update"].(map[string]any)
		scriptUpdate := update[scriptID].(map[string]any)
		assert.Equal(t, expectedActive, scriptUpdate["isActive"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["SieveScript/set", {
					"accountId": "acc1",
					"oldState": "sv1",
					"newState": "sv2",
					"updated": {
						"` + scriptID + `": null
					}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sieveSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := callService(ctx, client, scriptID)
	require.NoError(t, err)
}

func TestSieveService_Activate(t *testing.T) {
	runSieveActiveChangeTest(t, "sieve-inactive", true,
		func(ctx context.Context, c *Client, id string) error {
			return c.Sieve().Activate(ctx, id)
		})
}

func TestSieveService_Deactivate(t *testing.T) {
	runSieveActiveChangeTest(t, "sieve-active", false,
		func(ctx context.Context, c *Client, id string) error {
			return c.Sieve().Deactivate(ctx, id)
		})
}

func TestSieveService_Delete(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "SieveScript/set", methodName)

		destroy := args["destroy"].([]any)
		assert.Contains(t, destroy, "sieve-to-delete")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["SieveScript/set", {
					"accountId": "acc1",
					"oldState": "sv1",
					"newState": "sv2",
					"destroyed": ["sieve-to-delete"]
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sieveSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.Sieve().Delete(ctx, "sieve-to-delete")
	require.NoError(t, err)
}

func TestSieveService_Validate_Valid(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		assert.Equal(t, "SieveScript/validate", methodName)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["SieveScript/validate", {
					"accountId": "acc1"
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sieveSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	result, err := client.Sieve().Validate(ctx, `require "fileinto";`)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsValid)
}

func TestSieveService_Validate_Invalid(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["SieveScript/validate", {
					"accountId": "acc1",
					"error": {
						"type": "invalidScript",
						"description": "line 1: unknown command 'badcommand'"
					}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sieveSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	result, err := client.Sieve().Validate(ctx, "badcommand;")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsValid)
	assert.Equal(t, "invalidScript", result.ErrorType)
	assert.Contains(t, result.Description, "badcommand")
}

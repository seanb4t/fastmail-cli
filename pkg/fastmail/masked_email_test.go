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

func TestMaskedEmailService_List(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		assert.Equal(t, "MaskedEmail/get", methodName)

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
							"createdAt": "2024-01-15T10:30:00Z",
							"lastMessageAt": "2024-01-20T15:00:00Z"
						},
						{
							"id": "me-2",
							"email": "xyz789@fastmail.com",
							"state": "disabled",
							"forDomain": "newsletter.com",
							"description": "Newsletter signup",
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

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	emails, err := client.MaskedEmail().List(ctx)
	require.NoError(t, err)
	require.Len(t, emails, 2)

	assert.Equal(t, "me-1", emails[0].ID)
	assert.Equal(t, "abc123@fastmail.com", emails[0].Email)
	assert.Equal(t, MaskedEmailStateEnabled, emails[0].State)
	assert.Equal(t, "example.com", emails[0].ForDomain)
	assert.Equal(t, "Shopping site", emails[0].Description)

	assert.Equal(t, "me-2", emails[1].ID)
	assert.Equal(t, MaskedEmailStateDisabled, emails[1].State)
}

func TestMaskedEmailService_Create(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "MaskedEmail/set", methodName)

		// Verify create payload
		create := args["create"].(map[string]any)
		newME := create["new"].(map[string]any)
		assert.Equal(t, "shop.example.com", newME["forDomain"])
		assert.Equal(t, "enabled", newME["state"])

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
							"id": "me-new-123",
							"email": "newaddr@fastmail.com",
							"state": "enabled",
							"forDomain": "shop.example.com",
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
		_, _ = w.Write([]byte(maskedEmailSessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	opts := CreateMaskedEmailOptions{
		ForDomain:   "shop.example.com",
		Description: "Shopping site",
	}

	maskedEmail, err := client.MaskedEmail().Create(ctx, opts)
	require.NoError(t, err)
	require.NotNil(t, maskedEmail)

	assert.Equal(t, "me-new-123", maskedEmail.ID)
	assert.Equal(t, "newaddr@fastmail.com", maskedEmail.Email)
	assert.Equal(t, MaskedEmailStateEnabled, maskedEmail.State)
}

func TestMaskedEmailService_Enable(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "MaskedEmail/set", methodName)

		// Verify update payload
		update := args["update"].(map[string]any)
		meUpdate := update["me-disabled-1"].(map[string]any)
		assert.Equal(t, "enabled", meUpdate["state"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["MaskedEmail/set", {
					"accountId": "acc1",
					"oldState": "me1",
					"newState": "me2",
					"updated": {
						"me-disabled-1": {
							"id": "me-disabled-1",
							"state": "enabled"
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

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.MaskedEmail().Enable(ctx, "me-disabled-1")
	require.NoError(t, err)
}

func TestMaskedEmailService_Disable(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "MaskedEmail/set", methodName)

		// Verify update payload
		update := args["update"].(map[string]any)
		meUpdate := update["me-enabled-1"].(map[string]any)
		assert.Equal(t, "disabled", meUpdate["state"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["MaskedEmail/set", {
					"accountId": "acc1",
					"oldState": "me1",
					"newState": "me2",
					"updated": {
						"me-enabled-1": {
							"id": "me-enabled-1",
							"state": "disabled"
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

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.MaskedEmail().Disable(ctx, "me-enabled-1")
	require.NoError(t, err)
}

func TestMaskedEmailService_Delete(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "MaskedEmail/set", methodName)

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

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.MaskedEmail().Delete(ctx, "me-to-delete")
	require.NoError(t, err)
}

func TestMaskedEmailService_Delete_NotFound(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["MaskedEmail/set", {
					"accountId": "acc1",
					"oldState": "me1",
					"newState": "me1",
					"notDestroyed": {
						"nonexistent": {
							"type": "notFound",
							"description": "Masked email not found"
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

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.MaskedEmail().Delete(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notFound")
}

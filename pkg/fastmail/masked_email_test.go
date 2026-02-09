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

func TestMaskedEmailService_Get(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "MaskedEmail/get", methodName)
		assert.Equal(t, []any{"me-1"}, args["ids"])

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

	email, err := client.MaskedEmail().Get(ctx, "me-1")
	require.NoError(t, err)
	require.NotNil(t, email)

	assert.Equal(t, "me-1", email.ID)
	assert.Equal(t, "abc123@fastmail.com", email.Email)
	assert.Equal(t, MaskedEmailStateEnabled, email.State)
	assert.Equal(t, "example.com", email.ForDomain)
	assert.Equal(t, "Shopping site", email.Description)
}

func TestMaskedEmailService_Get_NotFound(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["MaskedEmail/get", {
					"accountId": "acc1",
					"state": "me1",
					"list": [],
					"notFound": ["nonexistent"]
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

	email, err := client.MaskedEmail().Get(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, email)
	assert.Contains(t, err.Error(), "not found")
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

// runMaskedEmailStateChangeServiceTest is a shared helper for Enable/Disable service tests
// to avoid code duplication. It sets up a mock API that verifies the MaskedEmail/set
// method call and state update payload.
func runMaskedEmailStateChangeServiceTest(t *testing.T, meID, expectedState string, callService func(context.Context, *Client, string) error) {
	t.Helper()

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
		meUpdate := update[meID].(map[string]any)
		assert.Equal(t, expectedState, meUpdate["state"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["MaskedEmail/set", {
					"accountId": "acc1",
					"oldState": "me1",
					"newState": "me2",
					"updated": {
						"` + meID + `": {
							"id": "` + meID + `",
							"state": "` + expectedState + `"
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

	err := callService(ctx, client, meID)
	require.NoError(t, err)
}

func TestMaskedEmailService_Enable(t *testing.T) {
	runMaskedEmailStateChangeServiceTest(t, "me-disabled-1", "enabled",
		func(ctx context.Context, c *Client, id string) error {
			return c.MaskedEmail().Enable(ctx, id)
		})
}

func TestMaskedEmailService_Disable(t *testing.T) {
	runMaskedEmailStateChangeServiceTest(t, "me-enabled-1", "disabled",
		func(ctx context.Context, c *Client, id string) error {
			return c.MaskedEmail().Disable(ctx, id)
		})
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

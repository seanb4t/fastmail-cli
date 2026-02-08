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

func TestIdentityService_List(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "Identity/get", methodName)
		assert.Equal(t, "acc1", args["accountId"])

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
							"bcc": [{"email": "bcc@example.com"}],
							"textSignature": "-- \nSent from CLI",
							"htmlSignature": "<p>Sent from CLI</p>",
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

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	identities, err := client.Identity().List(ctx)
	require.NoError(t, err)
	require.Len(t, identities, 2)

	first := identities[0]
	assert.Equal(t, "id-1", first.ID)
	assert.Equal(t, "Test User", first.Name)
	assert.Equal(t, "test@example.com", first.Email)
	require.Len(t, first.ReplyTo, 1)
	assert.Equal(t, "reply@example.com", first.ReplyTo[0].Email)
	assert.Equal(t, "Reply", first.ReplyTo[0].Name)
	require.Len(t, first.Bcc, 1)
	assert.Equal(t, "bcc@example.com", first.Bcc[0].Email)
	assert.Equal(t, "-- \nSent from CLI", first.TextSignature)
	assert.Equal(t, "<p>Sent from CLI</p>", first.HTMLSignature)
	assert.False(t, first.MayDelete)

	second := identities[1]
	assert.Equal(t, "id-2", second.ID)
	assert.Equal(t, "Alt Identity", second.Name)
	assert.Equal(t, "alt@example.com", second.Email)
	assert.True(t, second.MayDelete)
	assert.Empty(t, second.ReplyTo)
}

func TestIdentityService_List_Empty(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Identity/get", {
					"accountId": "acc1",
					"state": "i1",
					"list": [],
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

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	identities, err := client.Identity().List(ctx)
	require.NoError(t, err)
	assert.Empty(t, identities)
}

func TestIdentityService_Update(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "Identity/set", methodName)
		assert.Equal(t, "acc1", args["accountId"])

		update := args["update"].(map[string]any)
		idUpdate := update["id-1"].(map[string]any)
		assert.Equal(t, "Updated Name", idUpdate["name"])
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

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	name := "Updated Name"
	sig := "New sig"
	err := client.Identity().Update(ctx, "id-1", UpdateIdentityOptions{
		Name:          &name,
		TextSignature: &sig,
	})
	require.NoError(t, err)
}

func TestIdentityService_Update_ReplyTo(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)

		update := args["update"].(map[string]any)
		idUpdate := update["id-1"].(map[string]any)

		replyTo := idUpdate["replyTo"].([]any)
		require.Len(t, replyTo, 1)
		replyAddr := replyTo[0].(map[string]any)
		assert.Equal(t, "reply@example.com", replyAddr["email"])

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

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.Identity().Update(ctx, "id-1", UpdateIdentityOptions{
		ReplyTo: []EmailAddress{{Email: "reply@example.com"}},
	})
	require.NoError(t, err)
}

func TestIdentityService_Update_Error(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Identity/set", {
					"accountId": "acc1",
					"oldState": "i1",
					"newState": "i1",
					"notUpdated": {
						"id-1": {
							"type": "invalidProperties",
							"description": "name is too long"
						}
					}
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

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	name := "Bad Name"
	err := client.Identity().Update(ctx, "id-1", UpdateIdentityOptions{
		Name: &name,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalidProperties")
}

func TestIdentityService_Update_NoFields(t *testing.T) {
	client := NewClient("http://unused", "test-token")
	ctx := context.Background()

	err := client.Identity().Update(ctx, "id-1", UpdateIdentityOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no fields")
}

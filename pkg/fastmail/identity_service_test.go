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

// sessionResponseWithSubmission returns a valid JMAP session that includes the submission capability.
func sessionResponseWithSubmission(apiURL string) string {
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
				"accountCapabilities": {"urn:ietf:params:jmap:mail": {}}
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

		assert.Equal(t, "Identity/get", methodName)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Identity/get", {
					"accountId": "acc1",
					"state": "id1",
					"list": [
						{
							"id": "ident-1",
							"name": "Test User",
							"email": "test@example.com",
							"replyTo": [{"name": "Test User", "email": "reply@example.com"}],
							"bcc": null,
							"textSignature": "-- \nTest User",
							"htmlSignature": "<p>Test User</p>",
							"mayDelete": false
						},
						{
							"id": "ident-2",
							"name": "Work Alias",
							"email": "work@example.com",
							"replyTo": null,
							"bcc": [{"name": "", "email": "archive@example.com"}],
							"textSignature": "",
							"htmlSignature": "",
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
		_, _ = w.Write([]byte(sessionResponseWithSubmission(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	identities, err := client.Identity().List(ctx)
	require.NoError(t, err)
	require.Len(t, identities, 2)

	assert.Equal(t, "ident-1", identities[0].ID)
	assert.Equal(t, "Test User", identities[0].Name)
	assert.Equal(t, "test@example.com", identities[0].Email)
	require.Len(t, identities[0].ReplyTo, 1)
	assert.Equal(t, "reply@example.com", identities[0].ReplyTo[0].Email)
	assert.Equal(t, "-- \nTest User", identities[0].TextSignature)
	assert.Equal(t, "<p>Test User</p>", identities[0].HTMLSignature)
	assert.False(t, identities[0].MayDelete)

	assert.Equal(t, "ident-2", identities[1].ID)
	assert.Equal(t, "Work Alias", identities[1].Name)
	assert.Equal(t, "work@example.com", identities[1].Email)
	assert.Nil(t, identities[1].ReplyTo)
	require.Len(t, identities[1].BCC, 1)
	assert.Equal(t, "archive@example.com", identities[1].BCC[0].Email)
	assert.True(t, identities[1].MayDelete)
}

func TestIdentityService_List_Empty(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Identity/get", {
					"accountId": "acc1",
					"state": "id1",
					"list": [],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponseWithSubmission(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	identities, err := client.Identity().List(ctx)
	require.NoError(t, err)
	assert.Empty(t, identities)
}

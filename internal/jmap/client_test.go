package jmap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validSessionJSON returns a valid JMAP session response for testing.
func validSessionJSON() string {
	return `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {
				"maxSizeUpload": 50000000,
				"maxConcurrentUpload": 4,
				"maxSizeRequest": 10000000,
				"maxConcurrentRequests": 4,
				"maxCallsInRequest": 16,
				"maxObjectsInGet": 500,
				"maxObjectsInSet": 500,
				"collationAlgorithms": ["i;ascii-casemap"]
			},
			"urn:ietf:params:jmap:mail": {}
		},
		"accounts": {
			"u12345": {
				"name": "user@example.com",
				"isPersonal": true,
				"isReadOnly": false,
				"accountCapabilities": {
					"urn:ietf:params:jmap:mail": {}
				}
			}
		},
		"primaryAccounts": {
			"urn:ietf:params:jmap:mail": "u12345"
		},
		"username": "user@example.com",
		"apiUrl": "https://api.example.com/jmap/api/",
		"downloadUrl": "https://api.example.com/jmap/download/{accountId}/{blobId}/{name}",
		"uploadUrl": "https://api.example.com/jmap/upload/{accountId}/",
		"eventSourceUrl": "https://api.example.com/jmap/eventsource/",
		"state": "abc123"
	}`
}

func TestClient_Authenticate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Authorization header
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer test-token-123", auth)

		// Verify request method and path
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(validSessionJSON()))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token-123")
	session, err := client.Authenticate(context.Background())

	require.NoError(t, err)
	require.NotNil(t, session)

	// Verify session was parsed correctly
	assert.Equal(t, "https://api.example.com/jmap/api/", session.APIURL)
	assert.Equal(t, "user@example.com", session.Username)
	assert.Equal(t, "abc123", session.State)

	// Verify accounts
	require.Len(t, session.Accounts, 1)
	acc, ok := session.Accounts["u12345"]
	require.True(t, ok)
	assert.Equal(t, "user@example.com", acc.Name)
	assert.True(t, acc.IsPersonal)
	assert.False(t, acc.IsReadOnly)

	// Verify primaryAccounts
	assert.Equal(t, "u12345", session.PrimaryAccounts[CapMail])
}

func TestClient_AuthenticateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-token")
	session, err := client.Authenticate(context.Background())

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "401")
}

func TestClient_Session_Cached(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(validSessionJSON()))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx := context.Background()

	// First call should hit the server
	session1, err := client.Session(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second call should use cache
	session2, err := client.Session(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount) // Still 1, no new request

	// Should be same session
	assert.Equal(t, session1, session2)
}

func TestSession_MailAccountID(t *testing.T) {
	session := &Session{
		PrimaryAccounts: map[string]string{
			CapMail: "mail-account-123",
		},
	}

	assert.Equal(t, "mail-account-123", session.MailAccountID())
}

func TestSession_MailAccountID_NotFound(t *testing.T) {
	session := &Session{
		PrimaryAccounts: map[string]string{},
	}

	assert.Equal(t, "", session.MailAccountID())
}

func TestSession_HasCapability(t *testing.T) {
	session := &Session{
		Capabilities: map[string]json.RawMessage{
			CapCore: json.RawMessage(`{}`),
			CapMail: json.RawMessage(`{}`),
		},
	}

	assert.True(t, session.HasCapability(CapCore))
	assert.True(t, session.HasCapability(CapMail))
	assert.False(t, session.HasCapability(CapContacts))
}

func TestNewClient(t *testing.T) {
	client := NewClient("https://api.example.com/jmap/session", "my-token")

	assert.NotNil(t, client)
	assert.Equal(t, "https://api.example.com/jmap/session", client.endpoint)
	assert.Equal(t, "my-token", client.accessToken)
	assert.NotNil(t, client.httpClient)
}

package jmap

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRoundTripper implements http.RoundTripper for testing.
type mockRoundTripper struct {
	fn func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

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
	var httpErr *HTTPError
	if assert.ErrorAs(t, err, &httpErr) {
		assert.Equal(t, http.StatusUnauthorized, httpErr.StatusCode)
	}
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

func TestClient_WithHTTPClient(t *testing.T) {
	called := false
	mockTransport := &mockRoundTripper{
		fn: func(_ *http.Request) (*http.Response, error) {
			called = true
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(validSessionJSON())),
			}, nil
		},
	}

	client := NewClient("https://api.fastmail.com/jmap/session", "test-token",
		WithHTTPClient(&http.Client{Transport: mockTransport}))

	_, err := client.Authenticate(context.Background())
	require.NoError(t, err)
	assert.True(t, called, "custom HTTP client should be used")
}

func TestClient_Call(t *testing.T) {
	// Track requests to verify behavior
	var receivedRequest map[string]any
	var receivedAuth string

	// First create the API server that will handle the Call
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request body
		err := json.NewDecoder(r.Body).Decode(&receivedRequest)
		require.NoError(t, err)

		// Return a valid JMAP response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"sessionState": "state123",
			"methodResponses": [
				["Mailbox/get", {"accountId": "A1", "state": "m1", "list": [{"id": "inbox"}], "notFound": []}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	// Session server returns apiUrl pointing to our API server
	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"capabilities": {"urn:ietf:params:jmap:core": {}},
			"accounts": {"u1": {"name": "test", "isPersonal": true, "isReadOnly": false, "accountCapabilities": {}}},
			"primaryAccounts": {},
			"username": "test@example.com",
			"apiUrl": "` + apiServer.URL + `",
			"downloadUrl": "https://example.com/download",
			"uploadUrl": "https://example.com/upload",
			"eventSourceUrl": "https://example.com/events",
			"state": "s1"
		}`))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token-abc")
	ctx := context.Background()

	// Build a request
	req := NewRequest().WithCapabilities(CapCore, CapMail)
	req.Invoke("Mailbox/get", map[string]any{
		"accountId": "A1",
		"ids":       nil,
	})

	// Execute the call
	resp, err := client.Call(ctx, req)
	require.NoError(t, err)

	// Verify Authorization header was sent
	assert.Equal(t, "Bearer test-token-abc", receivedAuth)

	// Verify request structure
	assert.Equal(t, []any{CapCore, CapMail}, receivedRequest["using"])
	methodCalls := receivedRequest["methodCalls"].([]any)
	require.Len(t, methodCalls, 1)

	// Verify response
	assert.Equal(t, "state123", resp.SessionState)
	require.Len(t, resp.MethodResponses, 1)
	assert.Equal(t, "Mailbox/get", resp.MethodResponses[0].Name)
	assert.Equal(t, "0", resp.MethodResponses[0].CallID)
}

func TestClient_Call_HTTPError(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "server error"}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"capabilities": {},
			"accounts": {},
			"primaryAccounts": {},
			"username": "test",
			"apiUrl": "` + apiServer.URL + `",
			"downloadUrl": "",
			"uploadUrl": "",
			"eventSourceUrl": "",
			"state": ""
		}`))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "token")
	ctx := context.Background()

	req := NewRequest()
	req.Invoke("Test/method", map[string]any{})

	resp, err := client.Call(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	var httpErr *HTTPError
	if assert.ErrorAs(t, err, &httpErr) {
		assert.Equal(t, http.StatusInternalServerError, httpErr.StatusCode)
	}
}

func TestClient_Call_SessionError(t *testing.T) {
	// Session server returns error
	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "bad-token")
	ctx := context.Background()

	req := NewRequest()
	req.Invoke("Test/method", map[string]any{})

	resp, err := client.Call(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "getting session")
}

func TestClient_DownloadBlob(t *testing.T) {
	blobContent := "binary-blob-data-here"

	// Download server
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path contains the expected components
		assert.Contains(t, r.URL.Path, "u12345")
		assert.Contains(t, r.URL.Path, "blob-abc")
		assert.Contains(t, r.URL.Path, "test.pdf")
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(blobContent))
	}))
	defer downloadServer.Close()

	// Session server with downloadUrl pointing to our download server
	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"capabilities": {"urn:ietf:params:jmap:core": {}, "urn:ietf:params:jmap:mail": {}},
			"accounts": {"u12345": {"name": "test", "isPersonal": true, "isReadOnly": false, "accountCapabilities": {"urn:ietf:params:jmap:mail": {}}}},
			"primaryAccounts": {"urn:ietf:params:jmap:mail": "u12345"},
			"username": "test@example.com",
			"apiUrl": "https://api.example.com/jmap/api/",
			"downloadUrl": "` + downloadServer.URL + `/{accountId}/{blobId}/{name}",
			"uploadUrl": "https://api.example.com/jmap/upload/{accountId}/",
			"eventSourceUrl": "https://api.example.com/jmap/eventsource/",
			"state": "abc123"
		}`))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	reader, err := client.DownloadBlob(ctx, "u12345", "blob-abc", "test.pdf")
	require.NoError(t, err)
	require.NotNil(t, reader)
	defer func() { _ = reader.Close() }()

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, blobContent, string(data))
}

func TestClient_DownloadBlob_HTTPError(t *testing.T) {
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "not found"}`))
	}))
	defer downloadServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"capabilities": {},
			"accounts": {},
			"primaryAccounts": {},
			"username": "test",
			"apiUrl": "https://example.com/api",
			"downloadUrl": "` + downloadServer.URL + `/{accountId}/{blobId}/{name}",
			"uploadUrl": "",
			"eventSourceUrl": "",
			"state": ""
		}`))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	reader, err := client.DownloadBlob(ctx, "acc1", "bad-blob", "file.txt")
	assert.Error(t, err)
	assert.Nil(t, reader)
	var httpErr *HTTPError
	if assert.ErrorAs(t, err, &httpErr) {
		assert.Equal(t, http.StatusNotFound, httpErr.StatusCode)
	}
}

func TestClient_UploadBlob(t *testing.T) {
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/pdf", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Contains(t, r.URL.Path, "u12345")

		// Read the uploaded data
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, "fake-pdf-data", string(body))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"accountId": "u12345",
			"blobId": "blob-new-123",
			"type": "application/pdf",
			"size": 13
		}`))
	}))
	defer uploadServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"capabilities": {"urn:ietf:params:jmap:core": {}, "urn:ietf:params:jmap:mail": {}},
			"accounts": {"u12345": {"name": "test", "isPersonal": true, "isReadOnly": false, "accountCapabilities": {"urn:ietf:params:jmap:mail": {}}}},
			"primaryAccounts": {"urn:ietf:params:jmap:mail": "u12345"},
			"username": "test@example.com",
			"apiUrl": "https://api.example.com/jmap/api/",
			"downloadUrl": "https://api.example.com/jmap/download/{accountId}/{blobId}/{name}",
			"uploadUrl": "` + uploadServer.URL + `/{accountId}/",
			"eventSourceUrl": "https://api.example.com/jmap/eventsource/",
			"state": "abc123"
		}`))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	resp, err := client.UploadBlob(ctx, "u12345", strings.NewReader("fake-pdf-data"), "application/pdf")
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "u12345", resp.AccountID)
	assert.Equal(t, "blob-new-123", resp.BlobID)
	assert.Equal(t, "application/pdf", resp.Type)
	assert.Equal(t, uint64(13), resp.Size)
}

func TestClient_UploadBlob_HTTPError(t *testing.T) {
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, _ = w.Write([]byte(`{"error": "too large"}`))
	}))
	defer uploadServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"capabilities": {},
			"accounts": {},
			"primaryAccounts": {},
			"username": "test",
			"apiUrl": "https://example.com/api",
			"downloadUrl": "",
			"uploadUrl": "` + uploadServer.URL + `/{accountId}/",
			"eventSourceUrl": "",
			"state": ""
		}`))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	resp, err := client.UploadBlob(ctx, "acc1", strings.NewReader("data"), "application/pdf")
	assert.Error(t, err)
	assert.Nil(t, resp)
	var httpErr *HTTPError
	if assert.ErrorAs(t, err, &httpErr) {
		assert.Equal(t, http.StatusRequestEntityTooLarge, httpErr.StatusCode)
	}
}

func TestBuildDownloadURL(t *testing.T) {
	template := "https://api.example.com/jmap/download/{accountId}/{blobId}/{name}"
	url := buildDownloadURL(template, "acc-123", "blob-456", "report.pdf")
	assert.Equal(t, "https://api.example.com/jmap/download/acc-123/blob-456/report.pdf", url)
}

func TestBuildUploadURL(t *testing.T) {
	template := "https://api.example.com/jmap/upload/{accountId}/"
	url := buildUploadURL(template, "acc-123")
	assert.Equal(t, "https://api.example.com/jmap/upload/acc-123/", url)
}

func TestClient_Call_JMAPError(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["error", {"type": "unknownMethod", "description": "Method not found"}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"capabilities": {},
			"accounts": {},
			"primaryAccounts": {},
			"username": "test",
			"apiUrl": "` + apiServer.URL + `",
			"downloadUrl": "",
			"uploadUrl": "",
			"eventSourceUrl": "",
			"state": ""
		}`))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "token")
	ctx := context.Background()

	req := NewRequest()
	req.Invoke("Unknown/method", map[string]any{})

	// Call succeeds (HTTP 200), but response contains JMAP error
	resp, err := client.Call(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	result, err := resp.GetResult("0")
	require.NoError(t, err)
	assert.True(t, result.IsError())

	jmapErr := result.Error()
	assert.Equal(t, "unknownMethod", jmapErr.Type)
}

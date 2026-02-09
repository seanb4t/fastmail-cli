package fastmail

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSessionServer(t *testing.T) *httptest.Server {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(server.URL + "/jmap/api")))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	called := false
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(server.URL + "/jmap/api")))
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, "test-token", WithHTTPClient(server.Client()))
	err := client.Connect(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "acc1", client.accountID)
	assert.True(t, called, "custom HTTP client should be used")
}

func TestNewClient_DefaultClient(t *testing.T) {
	server := newTestSessionServer(t)

	client := NewClient(server.URL, "test-token")
	err := client.Connect(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "acc1", client.accountID)
}

package dav

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCardDAVClient(t *testing.T) {
	t.Run("with default endpoint", func(t *testing.T) {
		client, err := NewCardDAVClient("", "user@example.com", "token123")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "/dav/principals/user/user@example.com/", client.principal)
	})

	t.Run("with custom endpoint", func(t *testing.T) {
		client, err := NewCardDAVClient("https://custom.dav.example.com/", "user@example.com", "token123")
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("returns underlying client", func(t *testing.T) {
		client, err := NewCardDAVClient("", "user@example.com", "token123")
		require.NoError(t, err)
		assert.NotNil(t, client.Client())
	})
}

func TestNewCalDAVClient(t *testing.T) {
	t.Run("with default endpoint", func(t *testing.T) {
		client, err := NewCalDAVClient("", "user@example.com", "token123")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "/dav/principals/user/user@example.com/", client.principal)
	})

	t.Run("with custom endpoint", func(t *testing.T) {
		client, err := NewCalDAVClient("https://custom.caldav.example.com/", "user@example.com", "token123")
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("returns underlying client", func(t *testing.T) {
		client, err := NewCalDAVClient("", "user@example.com", "token123")
		require.NoError(t, err)
		assert.NotNil(t, client.Client())
	})
}

func TestBuildPrincipal(t *testing.T) {
	tests := []struct {
		user     string
		expected string
	}{
		{"user@example.com", "/dav/principals/user/user@example.com/"},
		{"john.doe@fastmail.com", "/dav/principals/user/john.doe@fastmail.com/"},
		{"test@domain.co.uk", "/dav/principals/user/test@domain.co.uk/"},
	}

	for _, tt := range tests {
		t.Run(tt.user, func(t *testing.T) {
			result := buildPrincipal(tt.user)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFastmailEndpoints(t *testing.T) {
	assert.Equal(t, "https://carddav.fastmail.com/", FastmailCardDAVEndpoint)
	assert.Equal(t, "https://caldav.fastmail.com/", FastmailCalDAVEndpoint)
}

func TestCardDAVClient_FindAddressBookHomeSet(t *testing.T) {
	// Create a mock WebDAV server that returns a PROPFIND response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Basic Auth header is present
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok, "Basic auth should be present")
		assert.Equal(t, "user@example.com", user)
		assert.Equal(t, "token123", pass)

		// Return a WebDAV multistatus response for address-book-home-set
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<D:multistatus xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:carddav">
  <D:response>
    <D:href>/dav/principals/user/user@example.com/</D:href>
    <D:propstat>
      <D:prop>
        <C:addressbook-home-set>
          <D:href>/dav/addressbooks/user/user@example.com/</D:href>
        </C:addressbook-home-set>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>`))
	}))
	defer server.Close()

	client, err := NewCardDAVClient(server.URL+"/", "user@example.com", "token123")
	require.NoError(t, err)

	homeSet, err := client.FindAddressBookHomeSet(context.Background())
	require.NoError(t, err)
	assert.Contains(t, homeSet, "addressbooks")
}

func TestCalDAVClient_FindCalendarHomeSet(t *testing.T) {
	// Create a mock WebDAV server that returns a PROPFIND response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Basic Auth header is present
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok, "Basic auth should be present")
		assert.Equal(t, "user@example.com", user)
		assert.Equal(t, "token123", pass)

		// Return a WebDAV multistatus response for calendar-home-set
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<D:multistatus xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:response>
    <D:href>/dav/principals/user/user@example.com/</D:href>
    <D:propstat>
      <D:prop>
        <C:calendar-home-set>
          <D:href>/dav/calendars/user/user@example.com/</D:href>
        </C:calendar-home-set>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>`))
	}))
	defer server.Close()

	client, err := NewCalDAVClient(server.URL+"/", "user@example.com", "token123")
	require.NoError(t, err)

	homeSet, err := client.FindCalendarHomeSet(context.Background())
	require.NoError(t, err)
	assert.Contains(t, homeSet, "calendars")
}

func TestCardDAVClient_FindAddressBookHomeSet_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := NewCardDAVClient(server.URL+"/", "user@example.com", "badtoken")
	require.NoError(t, err)

	_, err = client.FindAddressBookHomeSet(context.Background())
	assert.Error(t, err)
}

func TestCalDAVClient_FindCalendarHomeSet_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := NewCalDAVClient(server.URL+"/", "user@example.com", "badtoken")
	require.NoError(t, err)

	_, err = client.FindCalendarHomeSet(context.Background())
	assert.Error(t, err)
}

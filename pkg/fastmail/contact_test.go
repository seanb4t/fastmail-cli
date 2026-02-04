package fastmail_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCardDAVServer creates a test server that simulates Fastmail's CardDAV.
func mockCardDAVServer(_ *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "PROPFIND" && r.URL.Path == "/":
			// Return address book discovery response
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = w.Write([]byte(`<?xml version="1.0"?>
<d:multistatus xmlns:d="DAV:" xmlns:card="urn:ietf:params:xml:ns:carddav">
  <d:response>
    <d:href>/dav/addressbooks/user/test/</d:href>
    <d:propstat>
      <d:prop><d:resourcetype><d:collection/></d:resourcetype></d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
  <d:response>
    <d:href>/dav/addressbooks/user/test/Default/</d:href>
    <d:propstat>
      <d:prop>
        <d:resourcetype><d:collection/><card:addressbook/></d:resourcetype>
        <d:displayname>Default</d:displayname>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
</d:multistatus>`))

		case r.Method == "REPORT":
			// Return contacts list
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = w.Write([]byte(`<?xml version="1.0"?>
<d:multistatus xmlns:d="DAV:" xmlns:card="urn:ietf:params:xml:ns:carddav">
  <d:response>
    <d:href>/dav/addressbooks/user/test/Default/c1.vcf</d:href>
    <d:propstat>
      <d:prop>
        <d:getetag>"etag1"</d:getetag>
        <card:address-data>BEGIN:VCARD
VERSION:3.0
UID:c1
FN:Alice Wonder
N:Wonder;Alice;;;
EMAIL:alice@example.com
END:VCARD</card:address-data>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
  <d:response>
    <d:href>/dav/addressbooks/user/test/Default/c2.vcf</d:href>
    <d:propstat>
      <d:prop>
        <d:getetag>"etag2"</d:getetag>
        <card:address-data>BEGIN:VCARD
VERSION:3.0
UID:c2
FN:Bob Builder
N:Builder;Bob;;;
EMAIL:bob@example.com
END:VCARD</card:address-data>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
</d:multistatus>`))

		case r.Method == "GET":
			// Return single contact
			w.Header().Set("Content-Type", "text/vcard")
			w.Header().Set("ETag", `"single-etag"`)
			_, _ = w.Write([]byte(`BEGIN:VCARD
VERSION:3.0
UID:c1
FN:Alice Wonder
N:Wonder;Alice;;;
EMAIL:alice@example.com
TEL:+1-555-1234
END:VCARD`))

		case r.Method == "PUT":
			w.Header().Set("ETag", `"new-etag"`)
			// If-None-Match: * indicates a create, otherwise it's an update
			if r.Header.Get("If-None-Match") == "*" {
				w.WriteHeader(http.StatusCreated)
			} else {
				w.WriteHeader(http.StatusNoContent)
			}

		case r.Method == "DELETE":
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestContactsService_List(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	client := fastmail.NewContactsClient(server.URL, "test@example.com", "password")

	contacts, err := client.List(context.Background())
	require.NoError(t, err)
	require.Len(t, contacts, 2)

	assert.Equal(t, "Alice Wonder", contacts[0].Name)
	assert.Equal(t, "alice@example.com", contacts[0].Email)

	assert.Equal(t, "Bob Builder", contacts[1].Name)
	assert.Equal(t, "bob@example.com", contacts[1].Email)
}

func TestContactsService_Get(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	client := fastmail.NewContactsClient(server.URL, "test@example.com", "password")

	contact, err := client.Get(context.Background(), "c1")
	require.NoError(t, err)

	assert.Equal(t, "c1", contact.ID)
	assert.Equal(t, "Alice Wonder", contact.Name)
	assert.Equal(t, "alice@example.com", contact.Email)
	assert.Equal(t, "+1-555-1234", contact.Phone)
}

func TestContactsService_Create(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	client := fastmail.NewContactsClient(server.URL, "test@example.com", "password")

	contact := &fastmail.Contact{
		Name:  "New Contact",
		Email: "new@example.com",
		Phone: "+1-555-0000",
	}

	err := client.Create(context.Background(), contact)
	require.NoError(t, err)
	assert.NotEmpty(t, contact.ID) // ID should be generated
}

func TestContactsService_Update(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	client := fastmail.NewContactsClient(server.URL, "test@example.com", "password")

	contact := &fastmail.Contact{
		ID:    "c1",
		Name:  "Updated Name",
		Email: "updated@example.com",
	}

	err := client.Update(context.Background(), contact)
	require.NoError(t, err)
}

func TestContactsService_Delete(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	client := fastmail.NewContactsClient(server.URL, "test@example.com", "password")

	err := client.Delete(context.Background(), "c1")
	require.NoError(t, err)
}

func TestContactsService_Search(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	client := fastmail.NewContactsClient(server.URL, "test@example.com", "password")

	// Search filters on the client side
	contacts, err := client.Search(context.Background(), "Alice")
	require.NoError(t, err)
	require.Len(t, contacts, 1)
	assert.Equal(t, "Alice Wonder", contacts[0].Name)
}

package fastmail

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seanb4t/fastmail-cli/internal/dav"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContactsService_List(t *testing.T) {
	t.Run("returns empty list when no DAV client", func(t *testing.T) {
		service := &ContactsService{}

		contacts, err := service.List(context.Background())

		require.NoError(t, err)
		assert.Empty(t, contacts)
	})
}

func TestContactsService_Get(t *testing.T) {
	t.Run("returns error when no DAV client", func(t *testing.T) {
		service := &ContactsService{}

		_, err := service.Get(context.Background(), "some-id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	})
}

func TestContactsService_Search(t *testing.T) {
	t.Run("returns empty list when no DAV client", func(t *testing.T) {
		service := &ContactsService{}

		contacts, err := service.Search(context.Background(), "query")

		require.NoError(t, err)
		assert.Empty(t, contacts)
	})
}

func TestContact_Type(t *testing.T) {
	t.Run("contact has expected fields", func(t *testing.T) {
		contact := Contact{
			ID:      "abc123",
			Name:    "John Doe",
			Email:   "john@example.com",
			Phone:   "+1-555-1234",
			Address: "123 Main St, City, ST 12345",
		}

		assert.Equal(t, "abc123", contact.ID)
		assert.Equal(t, "John Doe", contact.Name)
		assert.Equal(t, "john@example.com", contact.Email)
		assert.Equal(t, "+1-555-1234", contact.Phone)
		assert.Equal(t, "123 Main St, City, ST 12345", contact.Address)
	})
}

// Mock WebDAV responses for CardDAV operations

const mockAddressBooksResponse = `<?xml version="1.0" encoding="UTF-8"?>
<D:multistatus xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:carddav">
  <D:response>
    <D:href>/dav/addressbooks/user/user@example.com/Default/</D:href>
    <D:propstat>
      <D:prop>
        <D:resourcetype>
          <D:collection/>
          <C:addressbook/>
        </D:resourcetype>
        <D:displayname>Default</D:displayname>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>`

const mockContactsQueryResponse = `<?xml version="1.0" encoding="UTF-8"?>
<D:multistatus xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:carddav">
  <D:response>
    <D:href>/dav/addressbooks/user/user@example.com/Default/contact1.vcf</D:href>
    <D:propstat>
      <D:prop>
        <D:getetag>"abc123"</D:getetag>
        <C:address-data>BEGIN:VCARD
VERSION:3.0
UID:contact1
FN:John Doe
EMAIL:john@example.com
TEL:+1-555-1234
END:VCARD</C:address-data>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
  <D:response>
    <D:href>/dav/addressbooks/user/user@example.com/Default/contact2.vcf</D:href>
    <D:propstat>
      <D:prop>
        <D:getetag>"def456"</D:getetag>
        <C:address-data>BEGIN:VCARD
VERSION:3.0
UID:contact2
FN:Jane Smith
EMAIL:jane@example.com
END:VCARD</C:address-data>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>`

func createMockCardDAVServer(_ *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")

		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = w.Write([]byte(mockAddressBooksResponse))
		case "REPORT":
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = w.Write([]byte(mockContactsQueryResponse))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestContactsService_ListWithDAVClient(t *testing.T) {
	server := createMockCardDAVServer(t)
	defer server.Close()

	davClient := dav.NewCardDAVClient(server.URL+"/", "user@example.com", "token123")

	service := &ContactsService{davClient: davClient}

	t.Run("lists contacts from address book", func(t *testing.T) {
		contacts, err := service.List(context.Background())

		require.NoError(t, err)
		assert.Len(t, contacts, 2)

		// Verify first contact
		assert.Equal(t, "contact1", contacts[0].ID)
		assert.Equal(t, "John Doe", contacts[0].Name)
		assert.Equal(t, "john@example.com", contacts[0].Email)
		assert.Equal(t, "+1-555-1234", contacts[0].Phone)

		// Verify second contact
		assert.Equal(t, "contact2", contacts[1].ID)
		assert.Equal(t, "Jane Smith", contacts[1].Name)
		assert.Equal(t, "jane@example.com", contacts[1].Email)
	})
}

func TestContactsService_SearchWithDAVClient(t *testing.T) {
	server := createMockCardDAVServer(t)
	defer server.Close()

	davClient := dav.NewCardDAVClient(server.URL+"/", "user@example.com", "token123")

	service := &ContactsService{davClient: davClient}

	t.Run("searches contacts by name", func(t *testing.T) {
		contacts, err := service.Search(context.Background(), "John")

		require.NoError(t, err)
		assert.Len(t, contacts, 1)
		assert.Equal(t, "John Doe", contacts[0].Name)
	})

	t.Run("searches contacts by email", func(t *testing.T) {
		contacts, err := service.Search(context.Background(), "jane@")

		require.NoError(t, err)
		assert.Len(t, contacts, 1)
		assert.Equal(t, "Jane Smith", contacts[0].Name)
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		contacts, err := service.Search(context.Background(), "nonexistent-xyz-123")

		require.NoError(t, err)
		assert.Empty(t, contacts)
	})
}

func TestContactsService_GetWithDAVClient(t *testing.T) {
	server := createMockCardDAVServer(t)
	defer server.Close()

	davClient := dav.NewCardDAVClient(server.URL+"/", "user@example.com", "token123")

	service := &ContactsService{davClient: davClient}

	t.Run("gets contact by ID", func(t *testing.T) {
		contact, err := service.Get(context.Background(), "contact1")

		require.NoError(t, err)
		require.NotNil(t, contact)
		assert.Equal(t, "contact1", contact.ID)
		assert.Equal(t, "John Doe", contact.Name)
	})

	t.Run("returns error for non-existent contact", func(t *testing.T) {
		_, err := service.Get(context.Background(), "nonexistent-contact")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

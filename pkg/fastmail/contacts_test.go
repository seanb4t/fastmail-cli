package fastmail

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestContactsService_Create(t *testing.T) {
	t.Run("returns error when no DAV client", func(t *testing.T) {
		service := &ContactsService{}
		contact := &Contact{
			FullName: "John Doe",
		}

		err := service.Create(context.Background(), contact)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	})
}

func TestContactsService_Update(t *testing.T) {
	t.Run("returns error when no DAV client", func(t *testing.T) {
		service := &ContactsService{}
		contact := &Contact{
			ID:       "some-id",
			FullName: "Jane Doe",
		}

		err := service.Update(context.Background(), contact)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	})
}

func TestContactsService_Delete(t *testing.T) {
	t.Run("returns error when no DAV client", func(t *testing.T) {
		service := &ContactsService{}

		err := service.Delete(context.Background(), "some-id")

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
			ID:       "abc123",
			FullName: "John Doe",
			Emails:   []string{"john@example.com", "johndoe@work.com"},
			Phones:   []string{"+1-555-1234"},
		}

		assert.Equal(t, "abc123", contact.ID)
		assert.Equal(t, "John Doe", contact.FullName)
		assert.Len(t, contact.Emails, 2)
		assert.Len(t, contact.Phones, 1)
	})
}

// Mock WebDAV responses for CardDAV operations

const mockAddressBookHomeSetResponse = `<?xml version="1.0" encoding="UTF-8"?>
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
</D:multistatus>`

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
EMAIL:jsmith@work.com
END:VCARD</C:address-data>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>`

func createMockCardDAVServer(_ *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")

		switch {
		case strings.Contains(r.URL.Path, "principals"):
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = w.Write([]byte(mockAddressBookHomeSetResponse))
		case strings.Contains(r.URL.Path, "addressbooks") && r.Method == "PROPFIND":
			if strings.HasSuffix(r.URL.Path, "/") && !strings.Contains(r.URL.Path, "Default") {
				w.WriteHeader(http.StatusMultiStatus)
				_, _ = w.Write([]byte(mockAddressBooksResponse))
			} else {
				w.WriteHeader(http.StatusMultiStatus)
				_, _ = w.Write([]byte(mockContactsQueryResponse))
			}
		case r.Method == "REPORT":
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

	davClient, err := dav.NewCardDAVClient(server.URL+"/", "user@example.com", "token123")
	require.NoError(t, err)

	service := &ContactsService{davClient: davClient}

	t.Run("lists contacts from address book", func(t *testing.T) {
		contacts, err := service.List(context.Background())

		require.NoError(t, err)
		assert.Len(t, contacts, 2)

		// Verify first contact
		assert.Equal(t, "contact1", contacts[0].ID)
		assert.Equal(t, "John Doe", contacts[0].FullName)
		assert.Contains(t, contacts[0].Emails, "john@example.com")
		assert.Contains(t, contacts[0].Phones, "+1-555-1234")

		// Verify second contact
		assert.Equal(t, "contact2", contacts[1].ID)
		assert.Equal(t, "Jane Smith", contacts[1].FullName)
		assert.Len(t, contacts[1].Emails, 2)
	})
}

func TestContactsService_SearchWithDAVClient(t *testing.T) {
	server := createMockCardDAVServer(t)
	defer server.Close()

	davClient, err := dav.NewCardDAVClient(server.URL+"/", "user@example.com", "token123")
	require.NoError(t, err)

	service := &ContactsService{davClient: davClient}

	t.Run("searches contacts by name", func(t *testing.T) {
		contacts, err := service.Search(context.Background(), "John")

		require.NoError(t, err)
		// Mock server returns all contacts for now, search filters client-side
		assert.GreaterOrEqual(t, len(contacts), 1)
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

	davClient, err := dav.NewCardDAVClient(server.URL+"/", "user@example.com", "token123")
	require.NoError(t, err)

	service := &ContactsService{davClient: davClient}

	t.Run("gets contact by ID", func(t *testing.T) {
		contact, err := service.Get(context.Background(), "contact1")

		require.NoError(t, err)
		require.NotNil(t, contact)
		assert.Equal(t, "contact1", contact.ID)
		assert.Equal(t, "John Doe", contact.FullName)
	})

	t.Run("returns error for non-existent contact", func(t *testing.T) {
		_, err := service.Get(context.Background(), "nonexistent-contact")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func createMockCardDAVServerWithWrite(_ *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")

		switch r.Method {
		case "PUT":
			// Create/Update contact
			w.Header().Set("ETag", `"new-etag"`)
			w.WriteHeader(http.StatusCreated)
		case "DELETE":
			// Delete contact
			w.WriteHeader(http.StatusNoContent)
		case "PROPFIND":
			if strings.Contains(r.URL.Path, "principals") {
				w.WriteHeader(http.StatusMultiStatus)
				_, _ = w.Write([]byte(mockAddressBookHomeSetResponse))
			} else if strings.Contains(r.URL.Path, "addressbooks") {
				if !strings.Contains(r.URL.Path, "Default") || strings.HasSuffix(r.URL.Path, "user@example.com/") {
					w.WriteHeader(http.StatusMultiStatus)
					_, _ = w.Write([]byte(mockAddressBooksResponse))
				} else {
					w.WriteHeader(http.StatusMultiStatus)
					_, _ = w.Write([]byte(mockContactsQueryResponse))
				}
			}
		case "REPORT":
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = w.Write([]byte(mockContactsQueryResponse))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestContactsService_CreateWithDAVClient(t *testing.T) {
	server := createMockCardDAVServerWithWrite(t)
	defer server.Close()

	davClient, err := dav.NewCardDAVClient(server.URL+"/", "user@example.com", "token123")
	require.NoError(t, err)

	service := &ContactsService{davClient: davClient}

	t.Run("creates new contact", func(t *testing.T) {
		contact := &Contact{
			FullName: "New Person",
			Emails:   []string{"new@example.com"},
		}

		err := service.Create(context.Background(), contact)

		require.NoError(t, err)
		// After create, the contact should have an ID assigned
		assert.NotEmpty(t, contact.ID)
	})

	t.Run("requires full name", func(t *testing.T) {
		contact := &Contact{
			Emails: []string{"noname@example.com"},
		}

		err := service.Create(context.Background(), contact)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "full name")
	})
}

func TestContactsService_DeleteWithDAVClient(t *testing.T) {
	server := createMockCardDAVServerWithWrite(t)
	defer server.Close()

	davClient, err := dav.NewCardDAVClient(server.URL+"/", "user@example.com", "token123")
	require.NoError(t, err)

	service := &ContactsService{davClient: davClient}

	t.Run("deletes contact by ID", func(t *testing.T) {
		// First, get a contact to know its path
		contact, err := service.Get(context.Background(), "contact1")
		require.NoError(t, err)

		err = service.Delete(context.Background(), contact.ID)

		require.NoError(t, err)
	})

	t.Run("returns error for non-existent contact", func(t *testing.T) {
		err := service.Delete(context.Background(), "nonexistent-contact")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestContactsService_UpdateWithDAVClient(t *testing.T) {
	server := createMockCardDAVServerWithWrite(t)
	defer server.Close()

	davClient, err := dav.NewCardDAVClient(server.URL+"/", "user@example.com", "token123")
	require.NoError(t, err)

	service := &ContactsService{davClient: davClient}

	t.Run("updates existing contact", func(t *testing.T) {
		contact := &Contact{
			ID:       "contact1",
			FullName: "John Doe Updated",
			Emails:   []string{"john.updated@example.com"},
		}

		err := service.Update(context.Background(), contact)

		require.NoError(t, err)
	})

	t.Run("returns error for non-existent contact", func(t *testing.T) {
		contact := &Contact{
			ID:       "nonexistent-contact",
			FullName: "Ghost Person",
		}

		err := service.Update(context.Background(), contact)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("requires full name", func(t *testing.T) {
		contact := &Contact{
			ID:     "contact1",
			Emails: []string{"noname@example.com"},
		}

		err := service.Update(context.Background(), contact)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "full name")
	})
}

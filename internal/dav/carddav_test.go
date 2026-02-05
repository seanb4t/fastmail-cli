package dav_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seanb4t/fastmail-cli/internal/dav"
)

func TestCardDAVClient_ListAddressBooks(t *testing.T) {
	// Minimal CardDAV PROPFIND response for address book discovery
	propfindResponse := `<?xml version="1.0" encoding="UTF-8"?>
<d:multistatus xmlns:d="DAV:" xmlns:card="urn:ietf:params:xml:ns:carddav">
  <d:response>
    <d:href>/dav/addressbooks/user/test@example.com/</d:href>
    <d:propstat>
      <d:prop>
        <d:resourcetype><d:collection/></d:resourcetype>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
  <d:response>
    <d:href>/dav/addressbooks/user/test@example.com/Default/</d:href>
    <d:propstat>
      <d:prop>
        <d:resourcetype><d:collection/><card:addressbook/></d:resourcetype>
        <d:displayname>Default</d:displayname>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
</d:multistatus>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PROPFIND", r.Method)
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(propfindResponse))
	}))
	defer server.Close()

	client := dav.NewCardDAVClient(server.URL, "test@example.com", "password")

	books, err := client.ListAddressBooks(context.Background())
	require.NoError(t, err)
	require.Len(t, books, 1)
	assert.Equal(t, "Default", books[0].Name)
	assert.Contains(t, books[0].Path, "Default")
}

func TestCardDAVClient_ListContacts(t *testing.T) {
	// CardDAV REPORT response with vCard data
	reportResponse := `<?xml version="1.0" encoding="UTF-8"?>
<d:multistatus xmlns:d="DAV:" xmlns:card="urn:ietf:params:xml:ns:carddav">
  <d:response>
    <d:href>/dav/addressbooks/user/test@example.com/Default/contact1.vcf</d:href>
    <d:propstat>
      <d:prop>
        <d:getetag>"abc123"</d:getetag>
        <card:address-data>BEGIN:VCARD
VERSION:3.0
UID:contact1
FN:John Doe
N:Doe;John;;;
EMAIL:john@example.com
TEL:+1-555-1234
END:VCARD</card:address-data>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
</d:multistatus>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(reportResponse))
	}))
	defer server.Close()

	client := dav.NewCardDAVClient(server.URL, "test@example.com", "password")

	contacts, err := client.ListContacts(context.Background(), "/dav/addressbooks/user/test@example.com/Default/")
	require.NoError(t, err)
	require.Len(t, contacts, 1)

	c := contacts[0]
	assert.Equal(t, "contact1", c.UID)
	assert.Equal(t, "John Doe", c.FormattedName)
	assert.Equal(t, "Doe", c.FamilyName)
	assert.Equal(t, "John", c.GivenName)
	require.Len(t, c.Emails, 1)
	assert.Equal(t, "john@example.com", c.Emails[0])
	require.Len(t, c.Phones, 1)
	assert.Equal(t, "+1-555-1234", c.Phones[0])
}

func TestCardDAVClient_GetContact(t *testing.T) {
	vcardData := `BEGIN:VCARD
VERSION:3.0
UID:contact1
FN:Jane Smith
N:Smith;Jane;;;
EMAIL;TYPE=WORK:jane.work@example.com
EMAIL;TYPE=HOME:jane.home@example.com
TEL;TYPE=CELL:+1-555-9999
ADR;TYPE=HOME:;;123 Main St;Springfield;IL;62701;USA
END:VCARD`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "text/vcard")
		w.Header().Set("ETag", `"etag123"`)
		_, _ = w.Write([]byte(vcardData))
	}))
	defer server.Close()

	client := dav.NewCardDAVClient(server.URL, "test@example.com", "password")

	contact, err := client.GetContact(context.Background(), "/contact1.vcf")
	require.NoError(t, err)

	assert.Equal(t, "contact1", contact.UID)
	assert.Equal(t, "Jane Smith", contact.FormattedName)
	assert.Equal(t, "Smith", contact.FamilyName)
	assert.Equal(t, "Jane", contact.GivenName)
	require.Len(t, contact.Emails, 2)
	assert.Contains(t, contact.Emails, "jane.work@example.com")
	assert.Contains(t, contact.Emails, "jane.home@example.com")
	require.Len(t, contact.Phones, 1)
	assert.Equal(t, "+1-555-9999", contact.Phones[0])
	require.Len(t, contact.Addresses, 1)
	assert.Equal(t, "123 Main St", contact.Addresses[0].Street)
	assert.Equal(t, "Springfield", contact.Addresses[0].City)
	assert.Equal(t, "IL", contact.Addresses[0].Region)
	assert.Equal(t, "62701", contact.Addresses[0].PostalCode)
	assert.Equal(t, "USA", contact.Addresses[0].Country)
}

func TestCardDAVClient_CreateContact(t *testing.T) {
	var capturedBody string
	var capturedMethod string
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		capturedBody = string(body)
		w.Header().Set("ETag", `"newetag"`)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := dav.NewCardDAVClient(server.URL, "test@example.com", "password")

	contact := &dav.Contact{
		UID:           "new-contact-uid",
		FormattedName: "New Person",
		GivenName:     "New",
		FamilyName:    "Person",
		Emails:        []string{"new@example.com"},
		Phones:        []string{"+1-555-0000"},
	}

	err := client.CreateContact(context.Background(), "/addressbook/", contact)
	require.NoError(t, err)

	assert.Equal(t, "PUT", capturedMethod)
	assert.Contains(t, capturedPath, "new-contact-uid")
	assert.Contains(t, capturedBody, "BEGIN:VCARD")
	assert.Contains(t, capturedBody, "FN:New Person")
	assert.Contains(t, capturedBody, "new@example.com")
}

func TestCardDAVClient_UpdateContact(t *testing.T) {
	var capturedBody string
	var capturedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		capturedBody = string(body)
		w.Header().Set("ETag", `"updatedetag"`)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := dav.NewCardDAVClient(server.URL, "test@example.com", "password")

	contact := &dav.Contact{
		UID:           "existing-uid",
		Path:          "/addressbook/existing-uid.vcf",
		FormattedName: "Updated Name",
		GivenName:     "Updated",
		FamilyName:    "Name",
		Emails:        []string{"updated@example.com"},
		ETag:          `"oldetag"`,
	}

	err := client.UpdateContact(context.Background(), contact)
	require.NoError(t, err)

	assert.Equal(t, "PUT", capturedMethod)
	assert.Contains(t, capturedBody, "FN:Updated Name")
}

func TestCardDAVClient_DeleteContact(t *testing.T) {
	var capturedMethod string
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := dav.NewCardDAVClient(server.URL, "test@example.com", "password")

	err := client.DeleteContact(context.Background(), "/addressbook/contact.vcf")
	require.NoError(t, err)

	assert.Equal(t, "DELETE", capturedMethod)
	assert.Equal(t, "/addressbook/contact.vcf", capturedPath)
}

func TestContact_VCardRoundTrip(t *testing.T) {
	// Test that a contact can be serialized to vCard and parsed back
	original := &dav.Contact{
		UID:           "roundtrip-uid",
		FormattedName: "Round Trip",
		GivenName:     "Round",
		FamilyName:    "Trip",
		Emails:        []string{"round@example.com", "trip@example.com"},
		Phones:        []string{"+1-111-1111", "+1-222-2222"},
		Addresses: []dav.Address{
			{
				Street:     "456 Oak Ave",
				City:       "Chicago",
				Region:     "IL",
				PostalCode: "60601",
				Country:    "USA",
			},
		},
	}

	vcardData, err := original.ToVCard()
	require.NoError(t, err)
	assert.True(t, strings.Contains(vcardData, "BEGIN:VCARD"))
	assert.True(t, strings.Contains(vcardData, "FN:Round Trip"))

	parsed, err := dav.ParseVCard(vcardData)
	require.NoError(t, err)

	assert.Equal(t, original.UID, parsed.UID)
	assert.Equal(t, original.FormattedName, parsed.FormattedName)
	assert.Equal(t, original.GivenName, parsed.GivenName)
	assert.Equal(t, original.FamilyName, parsed.FamilyName)
	assert.ElementsMatch(t, original.Emails, parsed.Emails)
	assert.ElementsMatch(t, original.Phones, parsed.Phones)
	require.Len(t, parsed.Addresses, 1)
	assert.Equal(t, original.Addresses[0].Street, parsed.Addresses[0].Street)
	assert.Equal(t, original.Addresses[0].City, parsed.Addresses[0].City)
}

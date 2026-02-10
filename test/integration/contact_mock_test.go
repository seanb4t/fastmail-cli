//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// MockContact holds contact test data for mock CardDAV server responses.
type MockContact struct {
	ID    string
	Name  string
	Email string
	Phone string
}

// contactDomainData holds typed state for contact scenarios.
type contactDomainData struct {
	Contacts []MockContact
}

func getContactData(w *World) *contactDomainData {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	data, ok := w.DomainData["contacts"].(*contactDomainData)
	if !ok {
		data = &contactDomainData{}
		w.DomainData["contacts"] = data
	}
	return data
}

// newMockCardDAVServer creates a mock CardDAV HTTP server backed by the given contacts.
func newMockCardDAVServer(contacts []MockContact) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "PROPFIND":
			servePropfindResponse(w)

		case r.Method == "REPORT":
			serveReportResponse(w, contacts)

		case r.Method == "GET":
			serveGetContactResponse(w, r, contacts)

		case r.Method == "PUT":
			servePutContactResponse(w, r)

		case r.Method == "DELETE":
			serveDeleteContactResponse(w)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func servePutContactResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("ETag", `"etag-new"`)
	// Create (If-None-Match: *) returns 201 Created; update returns 204 No Content.
	if r.Header.Get("If-None-Match") == "*" {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func serveDeleteContactResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func servePropfindResponse(w http.ResponseWriter) {
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
}

func serveReportResponse(w http.ResponseWriter, contacts []MockContact) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusMultiStatus)

	var responses strings.Builder
	for _, c := range contacts {
		responses.WriteString(contactToMultistatusEntry(c))
	}

	_, _ = fmt.Fprintf(w, `<?xml version="1.0"?>
<d:multistatus xmlns:d="DAV:" xmlns:card="urn:ietf:params:xml:ns:carddav">
%s</d:multistatus>`, responses.String())
}

func serveGetContactResponse(w http.ResponseWriter, r *http.Request, contacts []MockContact) {
	// Extract contact ID from path: /dav/addressbooks/user/test/Default/{id}.vcf
	path := r.URL.Path
	for _, c := range contacts {
		if strings.HasSuffix(path, "/"+c.ID+".vcf") {
			w.Header().Set("Content-Type", "text/vcard")
			w.Header().Set("ETag", `"etag-`+c.ID+`"`)
			_, _ = w.Write([]byte(contactToVCard(c)))
			return
		}
	}
	http.NotFound(w, r)
}

func contactToMultistatusEntry(c MockContact) string {
	return fmt.Sprintf(`  <d:response>
    <d:href>/dav/addressbooks/user/test/Default/%s.vcf</d:href>
    <d:propstat>
      <d:prop>
        <d:getetag>"etag-%s"</d:getetag>
        <card:address-data>%s</card:address-data>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
`, c.ID, c.ID, contactToVCard(c))
}

func contactToVCard(c MockContact) string {
	// Split name into given/family for the N field
	parts := strings.SplitN(c.Name, " ", 2)
	given := ""
	family := ""
	if len(parts) >= 1 {
		given = parts[0]
	}
	if len(parts) >= 2 {
		family = parts[1]
	}

	var b strings.Builder
	b.WriteString("BEGIN:VCARD\r\n")
	b.WriteString("VERSION:3.0\r\n")
	b.WriteString("UID:" + c.ID + "\r\n")
	b.WriteString("FN:" + c.Name + "\r\n")
	b.WriteString("N:" + family + ";" + given + ";;;\r\n")
	if c.Email != "" {
		b.WriteString("EMAIL:" + c.Email + "\r\n")
	}
	if c.Phone != "" {
		b.WriteString("TEL:" + c.Phone + "\r\n")
	}
	b.WriteString("END:VCARD\r\n")
	return b.String()
}

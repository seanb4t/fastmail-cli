package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func setupContactsTestEnv(t *testing.T, carddavURL string) (string, func()) {
	t.Helper()

	// Create temp directory for config
	tempDir, err := os.MkdirTemp("", "fastmail-cli-test-*")
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `carddav_endpoint: "` + carddavURL + `"
carddav_username: "test@example.com"`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set token via env var
	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	os.Setenv("FASTMAIL_TOKEN", "test-token")

	cleanup := func() {
		os.RemoveAll(tempDir)
		if originalEnv != "" {
			os.Setenv("FASTMAIL_TOKEN", originalEnv)
		} else {
			os.Unsetenv("FASTMAIL_TOKEN")
		}
	}

	return configPath, cleanup
}

func TestContactsListCommand(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	configPath, cleanup := setupContactsTestEnv(t, server.URL)
	defer cleanup()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "contacts", "list"})

	err := root.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Alice Wonder")
	assert.Contains(t, output, "Bob Builder")
}

func TestContactsListCommand_Search(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	configPath, cleanup := setupContactsTestEnv(t, server.URL)
	defer cleanup()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "contacts", "list", "--search", "Alice"})

	err := root.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Alice Wonder")
	assert.NotContains(t, output, "Bob Builder")
}

func TestContactsListCommand_JSON(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	configPath, cleanup := setupContactsTestEnv(t, server.URL)
	defer cleanup()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "--json", "contacts", "list"})

	err := root.Execute()
	require.NoError(t, err)

	var contacts []fastmail.Contact
	err = json.Unmarshal(out.Bytes(), &contacts)
	require.NoError(t, err)
	assert.Len(t, contacts, 2)
}

func TestContactsShowCommand(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	configPath, cleanup := setupContactsTestEnv(t, server.URL)
	defer cleanup()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "contacts", "show", "c1"})

	err := root.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Alice Wonder")
	assert.Contains(t, output, "alice@example.com")
	assert.Contains(t, output, "+1-555-1234")
}

func TestContactsShowCommand_JSON(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	configPath, cleanup := setupContactsTestEnv(t, server.URL)
	defer cleanup()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "--json", "contacts", "show", "c1"})

	err := root.Execute()
	require.NoError(t, err)

	var contact fastmail.Contact
	err = json.Unmarshal(out.Bytes(), &contact)
	require.NoError(t, err)
	assert.Equal(t, "Alice Wonder", contact.Name)
	assert.Equal(t, "alice@example.com", contact.Email)
}

func TestContactsCreateCommand(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	configPath, cleanup := setupContactsTestEnv(t, server.URL)
	defer cleanup()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "contacts", "create", "--name", "New Person", "--email", "new@example.com"})

	err := root.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Created:")
	assert.Contains(t, output, "New Person")
}

func TestContactsCreateCommand_MissingName(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	configPath, cleanup := setupContactsTestEnv(t, server.URL)
	defer cleanup()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "contacts", "create", "--email", "test@example.com"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

func TestContactsUpdateCommand(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	configPath, cleanup := setupContactsTestEnv(t, server.URL)
	defer cleanup()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "contacts", "update", "c1", "--name", "Updated Name"})

	err := root.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Updated")
	assert.Contains(t, output, "c1")
}

func TestContactsDeleteCommand_NoForce(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	configPath, cleanup := setupContactsTestEnv(t, server.URL)
	defer cleanup()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "contacts", "delete", "c1"})

	err := root.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Are you sure")
	assert.Contains(t, output, "--force")
}

func TestContactsDeleteCommand_WithForce(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	configPath, cleanup := setupContactsTestEnv(t, server.URL)
	defer cleanup()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "contacts", "delete", "c1", "--force"})

	err := root.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Deleted")
	assert.Contains(t, output, "c1")
}

func TestContactsCommand_NoAuth(t *testing.T) {
	// Create temp directory for config without token
	tempDir, err := os.MkdirTemp("", "fastmail-cli-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `carddav_endpoint: "https://example.com"
carddav_username: "test@example.com"`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Unset token env var
	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	os.Unsetenv("FASTMAIL_TOKEN")
	defer func() {
		if originalEnv != "" {
			os.Setenv("FASTMAIL_TOKEN", originalEnv)
		}
	}()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "contacts", "list"})

	err = root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting token")
}

func TestContactsCommand_NoUsername(t *testing.T) {
	server := mockCardDAVServer(t)
	defer server.Close()

	// Create temp directory for config without username
	tempDir, err := os.MkdirTemp("", "fastmail-cli-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `carddav_endpoint: "` + server.URL + `"`
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set token via env var
	originalEnv := os.Getenv("FASTMAIL_TOKEN")
	os.Setenv("FASTMAIL_TOKEN", "test-token")
	defer func() {
		if originalEnv != "" {
			os.Setenv("FASTMAIL_TOKEN", originalEnv)
		} else {
			os.Unsetenv("FASTMAIL_TOKEN")
		}
	}()

	root := NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"--config", configPath, "contacts", "list"})

	err = root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "carddav_username not configured")
}

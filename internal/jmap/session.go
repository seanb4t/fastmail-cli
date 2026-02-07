// Package jmap provides a client for the JMAP protocol.
package jmap

import "encoding/json"

// JMAP capability URIs.
const (
	CapCore             = "urn:ietf:params:jmap:core"
	CapMail             = "urn:ietf:params:jmap:mail"
	CapContacts         = "urn:ietf:params:jmap:contacts"
	CapCalendar         = "urn:ietf:params:jmap:calendars"
	CapVacationResponse = "urn:ietf:params:jmap:vacationresponse"
)

// Session represents a JMAP session resource.
// See: https://datatracker.ietf.org/doc/html/rfc8620#section-2
type Session struct {
	// Capabilities maps capability URIs to their configuration objects.
	Capabilities map[string]json.RawMessage `json:"capabilities"`

	// Accounts maps account IDs to Account objects.
	Accounts map[string]Account `json:"accounts"`

	// PrimaryAccounts maps capability URIs to the primary account ID for that capability.
	PrimaryAccounts map[string]string `json:"primaryAccounts"`

	// Username associated with this session.
	Username string `json:"username"`

	// APIURL is the URL to use for JMAP API requests.
	APIURL string `json:"apiUrl"`

	// DownloadURL is the URL template for downloading blobs.
	DownloadURL string `json:"downloadUrl"`

	// UploadURL is the URL template for uploading blobs.
	UploadURL string `json:"uploadUrl"`

	// EventSourceURL is the URL for push events.
	EventSourceURL string `json:"eventSourceUrl"`

	// State represents the current state of this session on the server.
	State string `json:"state"`
}

// Account represents a JMAP account.
type Account struct {
	// Name is a human-friendly name for the account.
	Name string `json:"name"`

	// IsPersonal is true if this account belongs to the authenticated user.
	IsPersonal bool `json:"isPersonal"`

	// IsReadOnly is true if the account is read-only.
	IsReadOnly bool `json:"isReadOnly"`

	// AccountCapabilities maps capability URIs to account-specific capability data.
	AccountCapabilities map[string]json.RawMessage `json:"accountCapabilities"`
}

// MailAccountID returns the primary account ID for mail capabilities.
// Returns empty string if no mail account is configured.
func (s *Session) MailAccountID() string {
	return s.PrimaryAccounts[CapMail]
}

// HasCapability returns true if the session supports the given capability.
func (s *Session) HasCapability(capability string) bool {
	_, ok := s.Capabilities[capability]
	return ok
}

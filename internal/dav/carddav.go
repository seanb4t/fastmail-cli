// Package dav provides WebDAV and CardDAV client functionality.
package dav

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/emersion/go-vcard"
	"github.com/samber/oops"
)

// Contact represents a contact from a CardDAV address book.
type Contact struct {
	UID           string
	Path          string
	ETag          string
	FormattedName string
	GivenName     string
	FamilyName    string
	Emails        []string
	Phones        []string
	Addresses     []Address
}

// Address represents a physical address.
type Address struct {
	Street     string
	City       string
	Region     string
	PostalCode string
	Country    string
}

// AddressBook represents a CardDAV address book.
type AddressBook struct {
	Name string
	Path string
}

// CardDAVClient provides CardDAV operations.
type CardDAVClient struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

// NewCardDAVClient creates a new CardDAV client.
func NewCardDAVClient(baseURL, username, password string) *CardDAVClient {
	return &CardDAVClient{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		username: username,
		password: password,
		client:   &http.Client{},
	}
}

// ListAddressBooks discovers address books for the authenticated user.
func (c *CardDAVClient) ListAddressBooks(ctx context.Context) ([]AddressBook, error) {
	propfindBody := `<?xml version="1.0" encoding="UTF-8"?>
<d:propfind xmlns:d="DAV:" xmlns:card="urn:ietf:params:xml:ns:carddav">
  <d:prop>
    <d:resourcetype/>
    <d:displayname/>
  </d:prop>
</d:propfind>`

	req, err := http.NewRequestWithContext(ctx, "PROPFIND", c.baseURL+"/", strings.NewReader(propfindBody))
	if err != nil {
		return nil, oops.Wrapf(err, "creating request")
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Depth", "1")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMultiStatus {
		return nil, oops.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, oops.Wrapf(err, "reading response")
	}

	return parseAddressBooksFromMultistatus(body)
}

// ListContacts returns all contacts from an address book.
func (c *CardDAVClient) ListContacts(ctx context.Context, addressBookPath string) ([]Contact, error) {
	reportBody := `<?xml version="1.0" encoding="UTF-8"?>
<card:addressbook-query xmlns:d="DAV:" xmlns:card="urn:ietf:params:xml:ns:carddav">
  <d:prop>
    <d:getetag/>
    <card:address-data/>
  </d:prop>
</card:addressbook-query>`

	url := c.baseURL + addressBookPath
	req, err := http.NewRequestWithContext(ctx, "REPORT", url, strings.NewReader(reportBody))
	if err != nil {
		return nil, oops.Wrapf(err, "creating request")
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Depth", "1")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMultiStatus {
		return nil, oops.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, oops.Wrapf(err, "reading response")
	}

	return parseContactsFromMultistatus(body)
}

// GetContact retrieves a single contact by path.
func (c *CardDAVClient) GetContact(ctx context.Context, contactPath string) (*Contact, error) {
	url := c.baseURL + contactPath
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, oops.Wrapf(err, "creating request")
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, oops.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, oops.Wrapf(err, "reading response")
	}

	contact, err := ParseVCard(string(body))
	if err != nil {
		return nil, oops.Wrapf(err, "parsing vCard")
	}

	contact.Path = contactPath
	contact.ETag = resp.Header.Get("ETag")

	return contact, nil
}

// CreateContact creates a new contact in the address book.
func (c *CardDAVClient) CreateContact(ctx context.Context, addressBookPath string, contact *Contact) error {
	vcardData, err := contact.ToVCard()
	if err != nil {
		return oops.Wrapf(err, "serializing contact")
	}

	contactPath := path.Join(addressBookPath, contact.UID+".vcf")
	url := c.baseURL + contactPath

	req, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(vcardData))
	if err != nil {
		return oops.Wrapf(err, "creating request")
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "text/vcard; charset=utf-8")
	req.Header.Set("If-None-Match", "*") // Only create if doesn't exist

	resp, err := c.client.Do(req)
	if err != nil {
		return oops.Wrapf(err, "executing request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return oops.Errorf("unexpected status: %d", resp.StatusCode)
	}

	contact.Path = contactPath
	contact.ETag = resp.Header.Get("ETag")

	return nil
}

// UpdateContact updates an existing contact.
func (c *CardDAVClient) UpdateContact(ctx context.Context, contact *Contact) error {
	if contact.Path == "" {
		return oops.Errorf("contact path is required for update")
	}

	vcardData, err := contact.ToVCard()
	if err != nil {
		return oops.Wrapf(err, "serializing contact")
	}

	url := c.baseURL + contact.Path
	req, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(vcardData))
	if err != nil {
		return oops.Wrapf(err, "creating request")
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "text/vcard; charset=utf-8")
	if contact.ETag != "" {
		req.Header.Set("If-Match", contact.ETag)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return oops.Wrapf(err, "executing request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return oops.Errorf("unexpected status: %d", resp.StatusCode)
	}

	contact.ETag = resp.Header.Get("ETag")

	return nil
}

// DeleteContact removes a contact.
func (c *CardDAVClient) DeleteContact(ctx context.Context, contactPath string) error {
	url := c.baseURL + contactPath
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, http.NoBody)
	if err != nil {
		return oops.Wrapf(err, "creating request")
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return oops.Wrapf(err, "executing request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return oops.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// ToVCard serializes a Contact to vCard format.
func (c *Contact) ToVCard() (string, error) {
	card := make(vcard.Card)

	card.SetValue(vcard.FieldVersion, "3.0")
	card.SetValue(vcard.FieldUID, c.UID)
	card.SetValue(vcard.FieldFormattedName, c.FormattedName)

	// Name field: family;given;additional;prefix;suffix
	nameField := &vcard.Field{
		Value: c.FamilyName + ";" + c.GivenName + ";;;",
	}
	card[vcard.FieldName] = []*vcard.Field{nameField}

	// Emails
	for _, email := range c.Emails {
		card.Add(vcard.FieldEmail, &vcard.Field{Value: email})
	}

	// Phones
	for _, phone := range c.Phones {
		card.Add(vcard.FieldTelephone, &vcard.Field{Value: phone})
	}

	// Addresses
	for _, addr := range c.Addresses {
		// ADR format: PO Box;Extended Address;Street;City;Region;Postal Code;Country
		addrValue := fmt.Sprintf(";;%s;%s;%s;%s;%s",
			addr.Street, addr.City, addr.Region, addr.PostalCode, addr.Country)
		card.Add(vcard.FieldAddress, &vcard.Field{Value: addrValue})
	}

	var buf bytes.Buffer
	enc := vcard.NewEncoder(&buf)
	if err := enc.Encode(card); err != nil {
		return "", oops.Wrapf(err, "encoding vCard")
	}

	return buf.String(), nil
}

// ParseVCard parses vCard data into a Contact.
func ParseVCard(data string) (*Contact, error) {
	dec := vcard.NewDecoder(strings.NewReader(data))
	card, err := dec.Decode()
	if err != nil {
		return nil, oops.Wrapf(err, "decoding vCard")
	}

	contact := &Contact{
		UID:           card.Value(vcard.FieldUID),
		FormattedName: card.Value(vcard.FieldFormattedName),
	}

	// Parse name components
	if nameFields := card[vcard.FieldName]; len(nameFields) > 0 {
		name := nameFields[0]
		parts := strings.Split(name.Value, ";")
		if len(parts) >= 1 {
			contact.FamilyName = parts[0]
		}
		if len(parts) >= 2 {
			contact.GivenName = parts[1]
		}
	}

	// Parse emails
	for _, field := range card[vcard.FieldEmail] {
		if field.Value != "" {
			contact.Emails = append(contact.Emails, field.Value)
		}
	}

	// Parse phones
	for _, field := range card[vcard.FieldTelephone] {
		if field.Value != "" {
			contact.Phones = append(contact.Phones, field.Value)
		}
	}

	// Parse addresses
	for _, field := range card[vcard.FieldAddress] {
		if field.Value != "" {
			addr := parseAddress(field.Value)
			contact.Addresses = append(contact.Addresses, addr)
		}
	}

	return contact, nil
}

func parseAddress(value string) Address {
	// ADR format: PO Box;Extended;Street;City;Region;Postal;Country
	parts := strings.Split(value, ";")
	addr := Address{}
	if len(parts) >= 3 {
		addr.Street = parts[2]
	}
	if len(parts) >= 4 {
		addr.City = parts[3]
	}
	if len(parts) >= 5 {
		addr.Region = parts[4]
	}
	if len(parts) >= 6 {
		addr.PostalCode = parts[5]
	}
	if len(parts) >= 7 {
		addr.Country = parts[6]
	}
	return addr
}

// XML structures for parsing CardDAV responses

type multistatus struct {
	XMLName   xml.Name   `xml:"multistatus"`
	Responses []response `xml:"response"`
}

type response struct {
	Href     string    `xml:"href"`
	Propstat *propstat `xml:"propstat"`
}

type propstat struct {
	Prop   prop   `xml:"prop"`
	Status string `xml:"status"`
}

type prop struct {
	ResourceType resourceType `xml:"resourcetype"`
	DisplayName  string       `xml:"displayname"`
	GetETag      string       `xml:"getetag"`
	AddressData  string       `xml:"address-data"`
}

type resourceType struct {
	Collection  *struct{} `xml:"collection"`
	AddressBook *struct{} `xml:"addressbook"`
}

func parseAddressBooksFromMultistatus(data []byte) ([]AddressBook, error) {
	var ms multistatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return nil, oops.Wrapf(err, "parsing multistatus")
	}

	var books []AddressBook
	for _, resp := range ms.Responses {
		if resp.Propstat == nil {
			continue
		}
		// Only include if it's an addressbook collection
		if resp.Propstat.Prop.ResourceType.AddressBook != nil {
			books = append(books, AddressBook{
				Name: resp.Propstat.Prop.DisplayName,
				Path: resp.Href,
			})
		}
	}

	return books, nil
}

func parseContactsFromMultistatus(data []byte) ([]Contact, error) {
	var ms multistatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return nil, oops.Wrapf(err, "parsing multistatus")
	}

	var contacts []Contact
	for _, resp := range ms.Responses {
		if resp.Propstat == nil {
			continue
		}

		vcardData := resp.Propstat.Prop.AddressData
		if vcardData == "" {
			continue
		}

		contact, err := ParseVCard(vcardData)
		if err != nil {
			continue // Skip invalid vcards
		}

		contact.Path = resp.Href
		contact.ETag = resp.Propstat.Prop.GetETag

		contacts = append(contacts, *contact)
	}

	return contacts, nil
}

package fastmail

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/dav"
)

// Contact represents a Fastmail contact.
type Contact struct {
	ID      string
	Name    string
	Email   string
	Phone   string
	Address string
}

// ContactsClient provides contact operations via CardDAV.
type ContactsClient struct {
	carddav         *dav.CardDAVClient
	addressBookPath string
}

// NewContactsClient creates a new contacts client.
// The endpoint should be the CardDAV URL (e.g., https://carddav.fastmail.com).
func NewContactsClient(endpoint, username, password string) *ContactsClient {
	return &ContactsClient{
		carddav: dav.NewCardDAVClient(endpoint, username, password),
	}
}

// List returns all contacts from the default address book.
func (c *ContactsClient) List(ctx context.Context) ([]Contact, error) {
	// Discover address book if not cached
	if c.addressBookPath == "" {
		books, err := c.carddav.ListAddressBooks(ctx)
		if err != nil {
			return nil, oops.Wrapf(err, "discovering address books")
		}
		if len(books) == 0 {
			return nil, oops.Errorf("no address books found")
		}
		c.addressBookPath = books[0].Path
	}

	davContacts, err := c.carddav.ListContacts(ctx, c.addressBookPath)
	if err != nil {
		return nil, oops.Wrapf(err, "listing contacts")
	}

	return convertDAVContacts(davContacts), nil
}

// Get returns a single contact by ID.
func (c *ContactsClient) Get(ctx context.Context, id string) (*Contact, error) {
	// Discover address book if not cached
	if c.addressBookPath == "" {
		books, err := c.carddav.ListAddressBooks(ctx)
		if err != nil {
			return nil, oops.Wrapf(err, "discovering address books")
		}
		if len(books) == 0 {
			return nil, oops.Errorf("no address books found")
		}
		c.addressBookPath = books[0].Path
	}

	contactPath := c.addressBookPath + id + ".vcf"
	davContact, err := c.carddav.GetContact(ctx, contactPath)
	if err != nil {
		return nil, oops.Wrapf(err, "getting contact")
	}

	return convertDAVContact(davContact), nil
}

// Create adds a new contact to the address book.
func (c *ContactsClient) Create(ctx context.Context, contact *Contact) error {
	// Discover address book if not cached
	if c.addressBookPath == "" {
		books, err := c.carddav.ListAddressBooks(ctx)
		if err != nil {
			return oops.Wrapf(err, "discovering address books")
		}
		if len(books) == 0 {
			return oops.Errorf("no address books found")
		}
		c.addressBookPath = books[0].Path
	}

	// Generate ID if not provided
	if contact.ID == "" {
		contact.ID = uuid.New().String()
	}

	davContact := toDavContact(contact)
	if err := c.carddav.CreateContact(ctx, c.addressBookPath, davContact); err != nil {
		return oops.Wrapf(err, "creating contact")
	}

	return nil
}

// Update modifies an existing contact.
func (c *ContactsClient) Update(ctx context.Context, contact *Contact) error {
	// Discover address book if not cached
	if c.addressBookPath == "" {
		books, err := c.carddav.ListAddressBooks(ctx)
		if err != nil {
			return oops.Wrapf(err, "discovering address books")
		}
		if len(books) == 0 {
			return oops.Errorf("no address books found")
		}
		c.addressBookPath = books[0].Path
	}

	davContact := toDavContact(contact)
	davContact.Path = c.addressBookPath + contact.ID + ".vcf"

	if err := c.carddav.UpdateContact(ctx, davContact); err != nil {
		return oops.Wrapf(err, "updating contact")
	}

	return nil
}

// Delete removes a contact by ID.
func (c *ContactsClient) Delete(ctx context.Context, id string) error {
	// Discover address book if not cached
	if c.addressBookPath == "" {
		books, err := c.carddav.ListAddressBooks(ctx)
		if err != nil {
			return oops.Wrapf(err, "discovering address books")
		}
		if len(books) == 0 {
			return oops.Errorf("no address books found")
		}
		c.addressBookPath = books[0].Path
	}

	contactPath := c.addressBookPath + id + ".vcf"
	if err := c.carddav.DeleteContact(ctx, contactPath); err != nil {
		return oops.Wrapf(err, "deleting contact")
	}

	return nil
}

// Search returns contacts matching a query string.
// The query matches against name and email.
func (c *ContactsClient) Search(ctx context.Context, query string) ([]Contact, error) {
	contacts, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var matches []Contact
	for _, contact := range contacts {
		if strings.Contains(strings.ToLower(contact.Name), query) ||
			strings.Contains(strings.ToLower(contact.Email), query) {
			matches = append(matches, contact)
		}
	}

	return matches, nil
}

func convertDAVContacts(davContacts []dav.Contact) []Contact {
	contacts := make([]Contact, len(davContacts))
	for i, dc := range davContacts {
		contacts[i] = *convertDAVContact(&dc)
	}
	return contacts
}

func convertDAVContact(dc *dav.Contact) *Contact {
	contact := &Contact{
		ID:   dc.UID,
		Name: dc.FormattedName,
	}

	if len(dc.Emails) > 0 {
		contact.Email = dc.Emails[0]
	}

	if len(dc.Phones) > 0 {
		contact.Phone = dc.Phones[0]
	}

	if len(dc.Addresses) > 0 {
		addr := dc.Addresses[0]
		parts := []string{}
		if addr.Street != "" {
			parts = append(parts, addr.Street)
		}
		if addr.City != "" {
			parts = append(parts, addr.City)
		}
		if addr.Region != "" {
			parts = append(parts, addr.Region)
		}
		if addr.PostalCode != "" {
			parts = append(parts, addr.PostalCode)
		}
		if addr.Country != "" {
			parts = append(parts, addr.Country)
		}
		contact.Address = strings.Join(parts, ", ")
	}

	return contact
}

func toDavContact(c *Contact) *dav.Contact {
	dc := &dav.Contact{
		UID:           c.ID,
		FormattedName: c.Name,
	}

	// Split name into given/family (simple heuristic)
	parts := strings.SplitN(c.Name, " ", 2)
	if len(parts) >= 1 {
		dc.GivenName = parts[0]
	}
	if len(parts) >= 2 {
		dc.FamilyName = parts[1]
	}

	if c.Email != "" {
		dc.Emails = []string{c.Email}
	}

	if c.Phone != "" {
		dc.Phones = []string{c.Phone}
	}

	return dc
}

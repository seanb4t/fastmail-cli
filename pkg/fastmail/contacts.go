package fastmail

import (
	"context"
	"strings"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/dav"
)

// ContactsService provides contact management operations via the Fastmail client.
// This wraps ContactsClient functionality for use with the Client API.
type ContactsService struct {
	davClient *dav.CardDAVClient
}

// List returns all contacts from the default address book.
func (s *ContactsService) List(ctx context.Context) ([]Contact, error) {
	if s.davClient == nil {
		return []Contact{}, nil
	}

	// Discover address books
	books, err := s.davClient.ListAddressBooks(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "discovering address books")
	}

	if len(books) == 0 {
		return []Contact{}, nil
	}

	// List contacts from the first address book
	davContacts, err := s.davClient.ListContacts(ctx, books[0].Path)
	if err != nil {
		return nil, oops.Wrapf(err, "listing contacts")
	}

	return convertDAVContactsToFastmail(davContacts), nil
}

// Get returns a single contact by ID.
func (s *ContactsService) Get(ctx context.Context, id string) (*Contact, error) {
	if s.davClient == nil {
		return nil, oops.Errorf("contacts service not configured")
	}

	// Get all contacts and find by ID
	contacts, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, c := range contacts {
		if c.ID == id {
			return &c, nil
		}
	}

	return nil, oops.Errorf("contact not found: %s", id)
}

// Search returns contacts matching a query string.
// Searches across name and email fields.
func (s *ContactsService) Search(ctx context.Context, query string) ([]Contact, error) {
	if s.davClient == nil {
		return []Contact{}, nil
	}

	// Get all contacts first
	allContacts, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	// Filter by query (case-insensitive match on name and email)
	query = strings.ToLower(query)
	var matches []Contact
	for _, c := range allContacts {
		if contactMatchesQuery(c, query) {
			matches = append(matches, c)
		}
	}

	return matches, nil
}

// contactMatchesQuery checks if a contact matches the search query.
func contactMatchesQuery(c Contact, query string) bool {
	if strings.Contains(strings.ToLower(c.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(c.Email), query) {
		return true
	}
	return false
}

// convertDAVContactsToFastmail converts DAV contacts to Fastmail Contact types.
func convertDAVContactsToFastmail(davContacts []dav.Contact) []Contact {
	contacts := make([]Contact, len(davContacts))
	for i, dc := range davContacts {
		contacts[i] = convertSingleDAVContact(&dc)
	}
	return contacts
}

// convertSingleDAVContact converts a single DAV Contact to Fastmail Contact.
func convertSingleDAVContact(dc *dav.Contact) Contact {
	contact := Contact{
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

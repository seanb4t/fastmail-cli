package fastmail

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/emersion/go-vcard"
	"github.com/emersion/go-webdav/carddav"
	"github.com/samber/oops"
	"github.com/seanb4t/fastmail-cli/internal/dav"
)

// Contact represents a contact from the address book.
type Contact struct {
	ID       string   // Unique identifier (vCard UID or path)
	Path     string   // WebDAV path for the contact
	FullName string   // Full name (FN property)
	Emails   []string // Email addresses
	Phones   []string // Phone numbers
}

// ContactsService provides contact management operations.
type ContactsService struct {
	client    *Client
	davClient *dav.CardDAVClient
}

// List returns all contacts from the default address book.
func (s *ContactsService) List(ctx context.Context) ([]Contact, error) {
	if s.davClient == nil {
		return []Contact{}, nil
	}

	// Find address books
	books, err := s.davClient.FindAddressBooks(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "finding address books")
	}

	if len(books) == 0 {
		return []Contact{}, nil
	}

	// Query the first address book for all contacts
	query := &carddav.AddressBookQuery{
		DataRequest: carddav.AddressDataRequest{
			AllProp: true,
		},
	}

	addressObjects, err := s.davClient.Client().QueryAddressBook(ctx, books[0].Path, query)
	if err != nil {
		return nil, oops.Wrapf(err, "querying address book")
	}

	return convertAddressObjects(addressObjects), nil
}

// Get returns a single contact by ID.
func (s *ContactsService) Get(ctx context.Context, id string) (*Contact, error) {
	if s.davClient == nil {
		return nil, oops.Errorf("contacts service not configured")
	}

	// Get all contacts and find by ID
	// (CardDAV doesn't have a direct "get by UID" - would need to query)
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

// Create adds a new contact to the address book.
func (s *ContactsService) Create(ctx context.Context, contact *Contact) error {
	if s.davClient == nil {
		return oops.Errorf("contacts service not configured")
	}

	if strings.TrimSpace(contact.FullName) == "" {
		return oops.Errorf("full name is required")
	}

	// Find address books
	books, err := s.davClient.FindAddressBooks(ctx)
	if err != nil {
		return oops.Wrapf(err, "finding address books")
	}

	if len(books) == 0 {
		return oops.Errorf("no address book found")
	}

	// Generate a unique ID if not provided
	if contact.ID == "" {
		contact.ID = generateUID()
	}

	// Build vCard
	card := buildVCard(contact)

	// Create the contact in the first address book
	path := books[0].Path + contact.ID + ".vcf"
	_, err = s.davClient.Client().PutAddressObject(ctx, path, card)
	if err != nil {
		return oops.Wrapf(err, "creating contact")
	}

	contact.Path = path
	return nil
}

// Update modifies an existing contact.
func (s *ContactsService) Update(ctx context.Context, contact *Contact) error {
	if s.davClient == nil {
		return oops.Errorf("contacts service not configured")
	}

	if strings.TrimSpace(contact.FullName) == "" {
		return oops.Errorf("full name is required")
	}

	// Find existing contact to get path
	existing, err := s.Get(ctx, contact.ID)
	if err != nil {
		return err
	}

	// Build vCard
	card := buildVCard(contact)

	// Update the contact
	_, err = s.davClient.Client().PutAddressObject(ctx, existing.Path, card)
	if err != nil {
		return oops.Wrapf(err, "updating contact")
	}

	contact.Path = existing.Path
	return nil
}

// Delete removes a contact from the address book.
func (s *ContactsService) Delete(ctx context.Context, id string) error {
	if s.davClient == nil {
		return oops.Errorf("contacts service not configured")
	}

	// Find the contact to get its path
	contact, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Delete using the WebDAV client
	err = s.davClient.Client().RemoveAll(ctx, contact.Path)
	if err != nil {
		return oops.Wrapf(err, "deleting contact")
	}

	return nil
}

// Search finds contacts matching the query string.
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
		if matchesQuery(c, query) {
			matches = append(matches, c)
		}
	}

	return matches, nil
}

// matchesQuery checks if a contact matches the search query.
func matchesQuery(c Contact, query string) bool {
	// Match against full name
	if strings.Contains(strings.ToLower(c.FullName), query) {
		return true
	}

	// Match against emails
	for _, email := range c.Emails {
		if strings.Contains(strings.ToLower(email), query) {
			return true
		}
	}

	return false
}

// generateUID creates a unique identifier for a new contact.
func generateUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// buildVCard creates a vCard from a Contact.
func buildVCard(c *Contact) vcard.Card {
	card := make(vcard.Card)

	// Required fields
	card.SetValue(vcard.FieldVersion, "3.0")
	card.SetValue(vcard.FieldUID, c.ID)
	card.SetValue(vcard.FieldFormattedName, c.FullName)

	// Email addresses
	for _, email := range c.Emails {
		card.Add(vcard.FieldEmail, &vcard.Field{Value: email})
	}

	// Phone numbers
	for _, phone := range c.Phones {
		card.Add(vcard.FieldTelephone, &vcard.Field{Value: phone})
	}

	return card
}

// convertAddressObjects converts CardDAV address objects to domain Contact types.
func convertAddressObjects(objects []carddav.AddressObject) []Contact {
	contacts := make([]Contact, 0, len(objects))
	for _, obj := range objects {
		contacts = append(contacts, convertAddressObject(obj))
	}
	return contacts
}

// convertAddressObject converts a single CardDAV address object to a Contact.
func convertAddressObject(obj carddav.AddressObject) Contact {
	contact := Contact{
		Path: obj.Path,
	}

	if obj.Card != nil {
		// Get UID
		if uid := obj.Card.Get("UID"); uid != nil {
			contact.ID = uid.Value
		}

		// Get full name
		if fn := obj.Card.Get("FN"); fn != nil {
			contact.FullName = fn.Value
		}

		// Get all email addresses
		for _, email := range obj.Card["EMAIL"] {
			if email.Value != "" {
				contact.Emails = append(contact.Emails, email.Value)
			}
		}

		// Get all phone numbers
		for _, tel := range obj.Card["TEL"] {
			if tel.Value != "" {
				contact.Phones = append(contact.Phones, tel.Value)
			}
		}
	}

	return contact
}

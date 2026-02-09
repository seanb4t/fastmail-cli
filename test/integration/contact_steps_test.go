//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// contactResult holds captured results from contact When steps.
type contactResult struct {
	Contacts      []fastmail.Contact
	SingleContact *fastmail.Contact
	OperationErr  error
}

func registerContactSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the following contacts exist:$`, theFollowingContactsExist)

	// When steps
	sc.Step(`^I list all contacts$`, iListAllContacts)
	sc.Step(`^I search contacts for "([^"]*)"$`, iSearchContacts)
	sc.Step(`^I get contact "([^"]*)"$`, iGetContact)

	// Then steps
	sc.Step(`^I should receive (\d+) contacts$`, iShouldReceiveNContacts)
	sc.Step(`^contact (\d+) should have name "([^"]*)"$`, contactShouldHaveName)
	sc.Step(`^contact (\d+) should have email "([^"]*)"$`, contactShouldHaveEmail)
	sc.Step(`^the contact should have name "([^"]*)"$`, theContactShouldHaveName)
	sc.Step(`^the contact should have email "([^"]*)"$`, theContactShouldHaveEmail)
	sc.Step(`^the contact should have phone "([^"]*)"$`, theContactShouldHavePhone)
	sc.Step(`^the contact operation should have failed$`, theContactOperationShouldHaveFailed)
}

func getContactResult(w *World) *contactResult {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	result, ok := w.DomainData["contact-result"].(*contactResult)
	if !ok {
		result = &contactResult{}
		w.DomainData["contact-result"] = result
	}
	return result
}

// Given steps

func theFollowingContactsExist(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getContactData(w)
	data.Contacts = parseContactTable(table)
	return ctx, nil
}

// When steps

func iListAllContacts(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getContactData(w)
	result := getContactResult(w)

	server := newMockCardDAVServer(data.Contacts)
	defer server.Close()

	client := fastmail.NewContactsClient(server.URL, "test", "test-token")
	contacts, err := client.List(ctx)
	result.OperationErr = err
	result.Contacts = contacts
	return ctx, nil
}

func iSearchContacts(ctx context.Context, query string) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getContactData(w)
	result := getContactResult(w)

	server := newMockCardDAVServer(data.Contacts)
	defer server.Close()

	client := fastmail.NewContactsClient(server.URL, "test", "test-token")
	contacts, err := client.Search(ctx, query)
	result.OperationErr = err
	result.Contacts = contacts
	return ctx, nil
}

func iGetContact(ctx context.Context, id string) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getContactData(w)
	result := getContactResult(w)

	server := newMockCardDAVServer(data.Contacts)
	defer server.Close()

	client := fastmail.NewContactsClient(server.URL, "test", "test-token")
	contact, err := client.Get(ctx, id)
	result.OperationErr = err
	result.SingleContact = contact
	return ctx, nil
}

// Then steps

func iShouldReceiveNContacts(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	result := getContactResult(w)
	if result.OperationErr != nil {
		return fmt.Errorf("unexpected error: %w", result.OperationErr)
	}
	if len(result.Contacts) != count {
		return fmt.Errorf("expected %d contacts, got %d", count, len(result.Contacts))
	}
	return nil
}

func contactShouldHaveName(ctx context.Context, index int, name string) error {
	w := WorldFromContext(ctx)
	result := getContactResult(w)
	i := index - 1
	if i < 0 || i >= len(result.Contacts) {
		return fmt.Errorf("contact index %d out of range (have %d)", index, len(result.Contacts))
	}
	actual := result.Contacts[i].Name
	if actual != name {
		return fmt.Errorf("expected name %q, got %q", name, actual)
	}
	return nil
}

func contactShouldHaveEmail(ctx context.Context, index int, email string) error {
	w := WorldFromContext(ctx)
	result := getContactResult(w)
	i := index - 1
	if i < 0 || i >= len(result.Contacts) {
		return fmt.Errorf("contact index %d out of range (have %d)", index, len(result.Contacts))
	}
	actual := result.Contacts[i].Email
	if actual != email {
		return fmt.Errorf("expected email %q, got %q", email, actual)
	}
	return nil
}

func theContactShouldHaveName(ctx context.Context, name string) error {
	w := WorldFromContext(ctx)
	result := getContactResult(w)
	if result.OperationErr != nil {
		return fmt.Errorf("unexpected error: %w", result.OperationErr)
	}
	if result.SingleContact == nil {
		return fmt.Errorf("no contact was retrieved")
	}
	if result.SingleContact.Name != name {
		return fmt.Errorf("expected name %q, got %q", name, result.SingleContact.Name)
	}
	return nil
}

func theContactShouldHaveEmail(ctx context.Context, email string) error {
	w := WorldFromContext(ctx)
	result := getContactResult(w)
	if result.OperationErr != nil {
		return fmt.Errorf("unexpected error: %w", result.OperationErr)
	}
	if result.SingleContact == nil {
		return fmt.Errorf("no contact was retrieved")
	}
	if result.SingleContact.Email != email {
		return fmt.Errorf("expected email %q, got %q", email, result.SingleContact.Email)
	}
	return nil
}

func theContactShouldHavePhone(ctx context.Context, phone string) error {
	w := WorldFromContext(ctx)
	result := getContactResult(w)
	if result.OperationErr != nil {
		return fmt.Errorf("unexpected error: %w", result.OperationErr)
	}
	if result.SingleContact == nil {
		return fmt.Errorf("no contact was retrieved")
	}
	if result.SingleContact.Phone != phone {
		return fmt.Errorf("expected phone %q, got %q", phone, result.SingleContact.Phone)
	}
	return nil
}

func theContactOperationShouldHaveFailed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	result := getContactResult(w)
	if result.OperationErr == nil {
		return fmt.Errorf("expected an error but operation succeeded")
	}
	return nil
}

// Helpers

func parseContactTable(table *godog.Table) []MockContact {
	if len(table.Rows) < 2 {
		return nil
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	var contacts []MockContact
	for _, row := range table.Rows[1:] {
		c := MockContact{
			ID:    contactCellValue(row, headers, "id"),
			Name:  contactCellValue(row, headers, "name"),
			Email: contactCellValue(row, headers, "email"),
			Phone: contactCellValue(row, headers, "phone"),
		}
		// Skip rows where all values are empty (header-only tables)
		if c.ID == "" && c.Name == "" && c.Email == "" && c.Phone == "" {
			continue
		}
		contacts = append(contacts, c)
	}
	return contacts
}

func contactCellValue(row *messages.PickleTableRow, headers map[string]int, key string) string {
	idx, ok := headers[key]
	if !ok || idx >= len(row.Cells) {
		return ""
	}
	return row.Cells[idx].Value
}

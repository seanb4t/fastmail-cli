//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// maskedEmailResult holds captured results from masked email When steps.
type maskedEmailResult struct {
	MaskedEmails []fastmail.MaskedEmail
	CreatedEmail *fastmail.MaskedEmail
	OperationErr error
}

func registerMaskedEmailSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the following masked emails exist:$`, theFollowingMaskedEmailsExist)

	// When steps
	sc.Step(`^I list masked emails$`, iListMaskedEmails)
	sc.Step(`^I create a masked email for domain "([^"]*)" with description "([^"]*)"$`, iCreateMaskedEmail)
	sc.Step(`^I disable masked email "([^"]*)"$`, iDisableMaskedEmail)
	sc.Step(`^I enable masked email "([^"]*)"$`, iEnableMaskedEmail)
	sc.Step(`^I delete masked email "([^"]*)"$`, iDeleteMaskedEmail)

	// Then steps
	sc.Step(`^I should receive (\d+) masked emails?$`, iShouldReceiveNMaskedEmails)
	sc.Step(`^masked email (\d+) should have email "([^"]*)"$`, maskedEmailShouldHaveEmail)
	sc.Step(`^masked email (\d+) should have state "([^"]*)"$`, maskedEmailShouldHaveState)
	sc.Step(`^masked email (\d+) should have forDomain "([^"]*)"$`, maskedEmailShouldHaveForDomain)
	sc.Step(`^the masked email creation should succeed$`, theMaskedEmailCreationShouldSucceed)
	sc.Step(`^the created masked email ID should not be empty$`, theCreatedMaskedEmailIDShouldNotBeEmpty)
	sc.Step(`^the masked email operation should succeed$`, theMaskedEmailOperationShouldSucceed)
}

func getMaskedEmailResult(w *World) *maskedEmailResult {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	result, ok := w.DomainData["masked-email-result"].(*maskedEmailResult)
	if !ok {
		result = &maskedEmailResult{}
		w.DomainData["masked-email-result"] = result
	}
	return result
}

// Given steps

func theFollowingMaskedEmailsExist(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getMaskedEmailData(w)
	data.MaskedEmails = parseMaskedEmailTable(table)
	return ctx, nil
}

// When steps

func iListMaskedEmails(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	maskedEmails, err := client.MaskedEmail().List(ctx)
	result.OperationErr = err
	result.MaskedEmails = maskedEmails
	return ctx, nil
}

func iCreateMaskedEmail(ctx context.Context, domain, description string) (context.Context, error) {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	created, err := client.MaskedEmail().Create(ctx, fastmail.CreateMaskedEmailOptions{
		ForDomain:   domain,
		Description: description,
	})
	result.OperationErr = err
	result.CreatedEmail = created
	return ctx, nil
}

func iDisableMaskedEmail(ctx context.Context, id string) (context.Context, error) {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	result.OperationErr = client.MaskedEmail().Disable(ctx, id)
	return ctx, nil
}

func iEnableMaskedEmail(ctx context.Context, id string) (context.Context, error) {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	result.OperationErr = client.MaskedEmail().Enable(ctx, id)
	return ctx, nil
}

func iDeleteMaskedEmail(ctx context.Context, id string) (context.Context, error) {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	result.OperationErr = client.MaskedEmail().Delete(ctx, id)
	return ctx, nil
}

// Then steps

func iShouldReceiveNMaskedEmails(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	if result.OperationErr != nil {
		return fmt.Errorf("unexpected error: %w", result.OperationErr)
	}
	if len(result.MaskedEmails) != count {
		return fmt.Errorf("expected %d masked emails, got %d", count, len(result.MaskedEmails))
	}
	return nil
}

func maskedEmailShouldHaveEmail(ctx context.Context, index int, email string) error {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	i := index - 1
	if i < 0 || i >= len(result.MaskedEmails) {
		return fmt.Errorf("masked email index %d out of range (have %d)", index, len(result.MaskedEmails))
	}
	actual := result.MaskedEmails[i].Email
	if actual != email {
		return fmt.Errorf("expected email %q, got %q", email, actual)
	}
	return nil
}

func maskedEmailShouldHaveState(ctx context.Context, index int, state string) error {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	i := index - 1
	if i < 0 || i >= len(result.MaskedEmails) {
		return fmt.Errorf("masked email index %d out of range (have %d)", index, len(result.MaskedEmails))
	}
	actual := string(result.MaskedEmails[i].State)
	if actual != state {
		return fmt.Errorf("expected state %q, got %q", state, actual)
	}
	return nil
}

func maskedEmailShouldHaveForDomain(ctx context.Context, index int, domain string) error {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	i := index - 1
	if i < 0 || i >= len(result.MaskedEmails) {
		return fmt.Errorf("masked email index %d out of range (have %d)", index, len(result.MaskedEmails))
	}
	actual := result.MaskedEmails[i].ForDomain
	if actual != domain {
		return fmt.Errorf("expected forDomain %q, got %q", domain, actual)
	}
	return nil
}

func theMaskedEmailCreationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	if result.OperationErr != nil {
		return fmt.Errorf("masked email creation failed: %w", result.OperationErr)
	}
	return nil
}

func theCreatedMaskedEmailIDShouldNotBeEmpty(ctx context.Context) error {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	if result.CreatedEmail == nil {
		return fmt.Errorf("created masked email is nil")
	}
	if result.CreatedEmail.ID == "" {
		return fmt.Errorf("created masked email ID is empty")
	}
	return nil
}

func theMaskedEmailOperationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	result := getMaskedEmailResult(w)
	if result.OperationErr != nil {
		return fmt.Errorf("masked email operation failed: %w", result.OperationErr)
	}
	return nil
}

// Helpers

func parseMaskedEmailTable(table *godog.Table) []MockMaskedEmail {
	if len(table.Rows) < 2 {
		return nil
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	var maskedEmails []MockMaskedEmail
	for _, row := range table.Rows[1:] {
		me := MockMaskedEmail{
			ID:          maskedEmailCellValue(row, headers, "id"),
			Email:       maskedEmailCellValue(row, headers, "email"),
			State:       maskedEmailCellValue(row, headers, "state"),
			ForDomain:   maskedEmailCellValue(row, headers, "forDomain"),
			Description: maskedEmailCellValue(row, headers, "description"),
			CreatedBy:   maskedEmailCellValue(row, headers, "createdBy"),
			CreatedAt:   maskedEmailCellValue(row, headers, "createdAt"),
		}
		maskedEmails = append(maskedEmails, me)
	}
	return maskedEmails
}

func maskedEmailCellValue(row *messages.PickleTableRow, headers map[string]int, key string) string {
	idx, ok := headers[key]
	if !ok || idx >= len(row.Cells) {
		return ""
	}
	return row.Cells[idx].Value
}

//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// identityResult stores the results from identity operations.
type identityResult struct {
	Identities []fastmail.Identity
	Error      error
}

func registerIdentitySteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the following identities exist:$`, theFollowingIdentitiesExist)

	// When steps
	sc.Step(`^I list all identities$`, iListAllIdentities)
	sc.Step(`^I update identity "([^"]*)" with name "([^"]*)"$`, iUpdateIdentityName)
	sc.Step(`^I update identity "([^"]*)" with signature "([^"]*)"$`, iUpdateIdentitySignature)

	// Then steps
	sc.Step(`^I should receive (\d+) identit(?:y|ies)$`, iShouldReceiveNIdentities)
	sc.Step(`^identity (\d+) should have name "([^"]*)"$`, identityShouldHaveName)
	sc.Step(`^identity (\d+) should have email "([^"]*)"$`, identityShouldHaveEmail)
	sc.Step(`^identity (\d+) should have signature "([^"]*)"$`, identityShouldHaveSignature)
	sc.Step(`^the identity update should succeed$`, theIdentityUpdateShouldSucceed)
}

// Given steps

func theFollowingIdentitiesExist(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getIdentityData(w)
	data.Identities = parseIdentityTable(table)
	return ctx, nil
}

// When steps

func iListAllIdentities(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	identities, err := client.Identity().List(ctx)
	result := &identityResult{
		Identities: identities,
		Error:      err,
	}
	if w.DomainData == nil {
		w.DomainData = map[string]any{}
	}
	w.DomainData["identityResult"] = result
	return ctx, nil
}

func iUpdateIdentityName(ctx context.Context, identityID, name string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	err := client.Identity().Update(ctx, identityID, fastmail.UpdateIdentityOptions{
		Name: &name,
	})
	result := &identityResult{Error: err}
	if w.DomainData == nil {
		w.DomainData = map[string]any{}
	}
	w.DomainData["identityResult"] = result
	return ctx, nil
}

func iUpdateIdentitySignature(ctx context.Context, identityID, signature string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	err := client.Identity().Update(ctx, identityID, fastmail.UpdateIdentityOptions{
		TextSignature: &signature,
	})
	result := &identityResult{Error: err}
	if w.DomainData == nil {
		w.DomainData = map[string]any{}
	}
	w.DomainData["identityResult"] = result
	return ctx, nil
}

// Then steps

func iShouldReceiveNIdentities(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	result := w.DomainData["identityResult"].(*identityResult)
	if result.Error != nil {
		return fmt.Errorf("unexpected error: %w", result.Error)
	}
	if len(result.Identities) != count {
		return fmt.Errorf("expected %d identities, got %d", count, len(result.Identities))
	}
	return nil
}

func identityShouldHaveName(ctx context.Context, index int, name string) error {
	w := WorldFromContext(ctx)
	result := w.DomainData["identityResult"].(*identityResult)
	i := index - 1
	if i < 0 || i >= len(result.Identities) {
		return fmt.Errorf("identity index %d out of range (have %d)", index, len(result.Identities))
	}
	actual := result.Identities[i].Name
	if actual != name {
		return fmt.Errorf("expected name %q, got %q", name, actual)
	}
	return nil
}

func identityShouldHaveEmail(ctx context.Context, index int, email string) error {
	w := WorldFromContext(ctx)
	result := w.DomainData["identityResult"].(*identityResult)
	i := index - 1
	if i < 0 || i >= len(result.Identities) {
		return fmt.Errorf("identity index %d out of range (have %d)", index, len(result.Identities))
	}
	actual := result.Identities[i].Email
	if actual != email {
		return fmt.Errorf("expected email %q, got %q", email, actual)
	}
	return nil
}

func identityShouldHaveSignature(ctx context.Context, index int, signature string) error {
	w := WorldFromContext(ctx)
	result := w.DomainData["identityResult"].(*identityResult)
	i := index - 1
	if i < 0 || i >= len(result.Identities) {
		return fmt.Errorf("identity index %d out of range (have %d)", index, len(result.Identities))
	}
	actual := result.Identities[i].TextSignature
	if actual != signature {
		return fmt.Errorf("expected signature %q, got %q", signature, actual)
	}
	return nil
}

func theIdentityUpdateShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	result := w.DomainData["identityResult"].(*identityResult)
	if result.Error != nil {
		return fmt.Errorf("identity update failed: %w", result.Error)
	}
	return nil
}

// Helpers

func parseIdentityTable(table *godog.Table) []mockIdentity {
	if len(table.Rows) < 2 {
		return nil
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	var identities []mockIdentity
	for _, row := range table.Rows[1:] {
		identities = append(identities, mockIdentity{
			ID:    identityCellValue(row, headers, "id"),
			Name:  identityCellValue(row, headers, "name"),
			Email: identityCellValue(row, headers, "email"),
		})
	}
	return identities
}

func identityCellValue(row *messages.PickleTableRow, headers map[string]int, key string) string {
	idx, ok := headers[key]
	if !ok || idx >= len(row.Cells) {
		return ""
	}
	return row.Cells[idx].Value
}

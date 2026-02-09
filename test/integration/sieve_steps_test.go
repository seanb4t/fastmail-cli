//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// sieveResult holds captured state from sieve When steps.
type sieveResult struct {
	scripts    []fastmail.SieveScript
	script     *fastmail.SieveScript
	created    *fastmail.SieveScript
	validation *fastmail.SieveValidationResult
	err        error
}

func registerSieveSteps(sc *godog.ScenarioContext) {
	// Given
	sc.Step(`^the following sieve scripts exist:$`, theFollowingSieveScriptsExist)

	// When
	sc.Step(`^I list sieve scripts$`, iListSieveScripts)
	sc.Step(`^I get sieve script "([^"]*)"$`, iGetSieveScript)
	sc.Step(`^I create a sieve script with name "([^"]*)" and script "([^"]*)"$`, iCreateSieveScript)
	sc.Step(`^I activate sieve script "([^"]*)"$`, iActivateSieveScript)
	sc.Step(`^I deactivate sieve script "([^"]*)"$`, iDeactivateSieveScript)
	sc.Step(`^I validate sieve script "([^"]*)"$`, iValidateSieveScript)
	sc.Step(`^I delete sieve script "([^"]*)"$`, iDeleteSieveScript)

	// Then
	sc.Step(`^I should receive (\d+) sieve scripts?$`, iShouldReceiveNSieveScripts)
	sc.Step(`^sieve script (\d+) should have name "([^"]*)"$`, sieveScriptShouldHaveName)
	sc.Step(`^sieve script (\d+) should be active$`, sieveScriptShouldBeActive)
	sc.Step(`^sieve script (\d+) should not be active$`, sieveScriptShouldNotBeActive)
	sc.Step(`^the sieve script should have name "([^"]*)"$`, theSieveScriptShouldHaveName)
	sc.Step(`^the sieve script should be active$`, theSieveScriptShouldBeActive)
	sc.Step(`^the sieve operation should succeed$`, theSieveOperationShouldSucceed)
	sc.Step(`^the created sieve script ID should not be empty$`, theCreatedSieveScriptIDShouldNotBeEmpty)
	sc.Step(`^the sieve validation should be valid$`, theSieveValidationShouldBeValid)
}

// getSieveResult retrieves or initializes the sieve result from DomainData.
func getSieveResult(w *World) *sieveResult {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	r, ok := w.DomainData["sieve"].(*sieveResult)
	if !ok {
		r = &sieveResult{}
		w.DomainData["sieve"] = r
	}
	return r
}

// Given steps

func theFollowingSieveScriptsExist(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	w.DomainData["sieve_scripts"] = parseSieveTable(table)
	return ctx, nil
}

// When steps

func iListSieveScripts(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	r := getSieveResult(w)
	r.scripts, r.err = client.Sieve().List(ctx)
	return ctx, nil
}

func iGetSieveScript(ctx context.Context, id string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	r := getSieveResult(w)
	r.script, r.err = client.Sieve().Get(ctx, id)
	return ctx, nil
}

func iCreateSieveScript(ctx context.Context, name, script string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	r := getSieveResult(w)
	r.created, r.err = client.Sieve().Create(ctx, fastmail.CreateSieveScriptOptions{
		Name:   name,
		Script: script,
	})
	return ctx, nil
}

func iActivateSieveScript(ctx context.Context, id string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	r := getSieveResult(w)
	r.err = client.Sieve().Activate(ctx, id)
	return ctx, nil
}

func iDeactivateSieveScript(ctx context.Context, id string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	r := getSieveResult(w)
	r.err = client.Sieve().Deactivate(ctx, id)
	return ctx, nil
}

func iValidateSieveScript(ctx context.Context, script string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	r := getSieveResult(w)
	r.validation, r.err = client.Sieve().Validate(ctx, script)
	return ctx, nil
}

func iDeleteSieveScript(ctx context.Context, id string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	r := getSieveResult(w)
	r.err = client.Sieve().Delete(ctx, id)
	return ctx, nil
}

// Then steps

func iShouldReceiveNSieveScripts(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	r := getSieveResult(w)
	if r.err != nil {
		return fmt.Errorf("unexpected error: %w", r.err)
	}
	if len(r.scripts) != count {
		return fmt.Errorf("expected %d sieve scripts, got %d", count, len(r.scripts))
	}
	return nil
}

func sieveScriptShouldHaveName(ctx context.Context, index int, name string) error {
	w := WorldFromContext(ctx)
	r := getSieveResult(w)
	i := index - 1
	if i < 0 || i >= len(r.scripts) {
		return fmt.Errorf("sieve script index %d out of range (have %d)", index, len(r.scripts))
	}
	if r.scripts[i].Name != name {
		return fmt.Errorf("expected name %q, got %q", name, r.scripts[i].Name)
	}
	return nil
}

func sieveScriptShouldBeActive(ctx context.Context, index int) error {
	w := WorldFromContext(ctx)
	r := getSieveResult(w)
	i := index - 1
	if i < 0 || i >= len(r.scripts) {
		return fmt.Errorf("sieve script index %d out of range", index)
	}
	if !r.scripts[i].IsActive {
		return fmt.Errorf("expected sieve script %d to be active", index)
	}
	return nil
}

func sieveScriptShouldNotBeActive(ctx context.Context, index int) error {
	w := WorldFromContext(ctx)
	r := getSieveResult(w)
	i := index - 1
	if i < 0 || i >= len(r.scripts) {
		return fmt.Errorf("sieve script index %d out of range", index)
	}
	if r.scripts[i].IsActive {
		return fmt.Errorf("expected sieve script %d to not be active", index)
	}
	return nil
}

func theSieveScriptShouldHaveName(ctx context.Context, name string) error {
	w := WorldFromContext(ctx)
	r := getSieveResult(w)
	if r.err != nil {
		return fmt.Errorf("unexpected error: %w", r.err)
	}
	if r.script == nil {
		return fmt.Errorf("no sieve script result")
	}
	if r.script.Name != name {
		return fmt.Errorf("expected name %q, got %q", name, r.script.Name)
	}
	return nil
}

func theSieveScriptShouldBeActive(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getSieveResult(w)
	if r.err != nil {
		return fmt.Errorf("unexpected error: %w", r.err)
	}
	if r.script == nil {
		return fmt.Errorf("no sieve script result")
	}
	if !r.script.IsActive {
		return fmt.Errorf("expected sieve script to be active")
	}
	return nil
}

func theSieveOperationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getSieveResult(w)
	if r.err != nil {
		return fmt.Errorf("sieve operation failed: %w", r.err)
	}
	return nil
}

func theCreatedSieveScriptIDShouldNotBeEmpty(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getSieveResult(w)
	if r.created == nil {
		return fmt.Errorf("no created sieve script")
	}
	if r.created.ID == "" {
		return fmt.Errorf("created sieve script ID is empty")
	}
	return nil
}

func theSieveValidationShouldBeValid(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getSieveResult(w)
	if r.err != nil {
		return fmt.Errorf("unexpected error: %w", r.err)
	}
	if r.validation == nil {
		return fmt.Errorf("no validation result")
	}
	if !r.validation.IsValid {
		return fmt.Errorf("expected valid script, got error: %s - %s", r.validation.ErrorType, r.validation.Description)
	}
	return nil
}

// Helpers

func parseSieveTable(table *godog.Table) []MockSieveScript {
	if len(table.Rows) < 2 {
		return nil
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	var scripts []MockSieveScript
	for _, row := range table.Rows[1:] {
		s := MockSieveScript{
			ID:       sieveCellValue(row, headers, "id"),
			Name:     sieveCellValue(row, headers, "name"),
			Script:   sieveCellValue(row, headers, "script"),
			IsActive: sieveCellValue(row, headers, "isActive") == "true",
		}
		scripts = append(scripts, s)
	}
	return scripts
}

func sieveCellValue(row *messages.PickleTableRow, headers map[string]int, key string) string {
	idx, ok := headers[key]
	if !ok || idx >= len(row.Cells) {
		return ""
	}
	return row.Cells[idx].Value
}

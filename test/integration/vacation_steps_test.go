//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// vacationResult holds the result of a vacation operation for Then step assertions.
type vacationResult struct {
	Status *fastmail.VacationStatus
	Error  error
}

const vacationResultKey = "vacationResult"

func getVacationResult(w *World) *vacationResult {
	if w.DomainData == nil {
		w.DomainData = map[string]any{}
	}
	if v, ok := w.DomainData[vacationResultKey]; ok {
		return v.(*vacationResult)
	}
	result := &vacationResult{}
	w.DomainData[vacationResultKey] = result
	return result
}

func registerVacationSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the vacation response is disabled$`, theVacationResponseIsDisabled)
	sc.Step(`^the vacation response is enabled with subject "([^"]*)" and body "([^"]*)"$`, theVacationResponseIsEnabled)

	// When steps
	sc.Step(`^I get the vacation status$`, iGetTheVacationStatus)
	sc.Step(`^I enable the vacation response with subject "([^"]*)" and body "([^"]*)"$`, iEnableTheVacationResponse)
	sc.Step(`^I disable the vacation response$`, iDisableTheVacationResponse)

	// Then steps
	sc.Step(`^the vacation response should be disabled$`, theVacationResponseShouldBeDisabled)
	sc.Step(`^the vacation response should be enabled$`, theVacationResponseShouldBeEnabled)
	sc.Step(`^the vacation subject should be "([^"]*)"$`, theVacationSubjectShouldBe)
	sc.Step(`^the vacation body should be "([^"]*)"$`, theVacationBodyShouldBe)
	sc.Step(`^the vacation operation should succeed$`, theVacationOperationShouldSucceed)
}

// Given steps

func theVacationResponseIsDisabled(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	if w.DomainData == nil {
		w.DomainData = map[string]any{}
	}
	w.DomainData[vacationDataKey] = &vacationState{
		IsEnabled: false,
	}
	return ctx, nil
}

func theVacationResponseIsEnabled(ctx context.Context, subject, body string) (context.Context, error) {
	w := WorldFromContext(ctx)
	if w.DomainData == nil {
		w.DomainData = map[string]any{}
	}
	w.DomainData[vacationDataKey] = &vacationState{
		IsEnabled: true,
		Subject:   subject,
		TextBody:  body,
	}
	return ctx, nil
}

// When steps

func iGetTheVacationStatus(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	status, err := client.Vacation().GetStatus(ctx)
	result := getVacationResult(w)
	result.Status = status
	result.Error = err
	return ctx, nil
}

func iEnableTheVacationResponse(ctx context.Context, subject, body string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	err := client.Vacation().Enable(ctx, subject, body, nil, nil)
	result := getVacationResult(w)
	result.Error = err
	return ctx, nil
}

func iDisableTheVacationResponse(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	err := client.Vacation().Disable(ctx)
	result := getVacationResult(w)
	result.Error = err
	return ctx, nil
}

// Then steps

func theVacationResponseShouldBeDisabled(ctx context.Context) error {
	w := WorldFromContext(ctx)
	result := getVacationResult(w)
	if result.Error != nil {
		return fmt.Errorf("unexpected error: %w", result.Error)
	}
	if result.Status == nil {
		return fmt.Errorf("vacation status is nil")
	}
	if result.Status.IsEnabled {
		return fmt.Errorf("expected vacation to be disabled, but it is enabled")
	}
	return nil
}

func theVacationResponseShouldBeEnabled(ctx context.Context) error {
	w := WorldFromContext(ctx)
	result := getVacationResult(w)
	if result.Error != nil {
		return fmt.Errorf("unexpected error: %w", result.Error)
	}
	if result.Status == nil {
		return fmt.Errorf("vacation status is nil")
	}
	if !result.Status.IsEnabled {
		return fmt.Errorf("expected vacation to be enabled, but it is disabled")
	}
	return nil
}

func theVacationSubjectShouldBe(ctx context.Context, expected string) error {
	w := WorldFromContext(ctx)
	result := getVacationResult(w)
	if result.Error != nil {
		return fmt.Errorf("unexpected error: %w", result.Error)
	}
	if result.Status == nil {
		return fmt.Errorf("vacation status is nil")
	}
	if result.Status.Subject != expected {
		return fmt.Errorf("expected subject %q, got %q", expected, result.Status.Subject)
	}
	return nil
}

func theVacationBodyShouldBe(ctx context.Context, expected string) error {
	w := WorldFromContext(ctx)
	result := getVacationResult(w)
	if result.Error != nil {
		return fmt.Errorf("unexpected error: %w", result.Error)
	}
	if result.Status == nil {
		return fmt.Errorf("vacation status is nil")
	}
	if result.Status.TextBody != expected {
		return fmt.Errorf("expected body %q, got %q", expected, result.Status.TextBody)
	}
	return nil
}

func theVacationOperationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	result := getVacationResult(w)
	if result.Error != nil {
		return fmt.Errorf("vacation operation failed: %w", result.Error)
	}
	return nil
}

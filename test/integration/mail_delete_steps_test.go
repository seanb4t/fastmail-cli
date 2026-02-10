//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func registerMailDeleteSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^an email "([^"]*)" exists in trash with subject "([^"]*)"$`, anEmailExistsInTrash)

	// When steps
	sc.Step(`^I delete email "([^"]*)"$`, iDeleteEmail)

	// Then steps
	sc.Step(`^the email delete should succeed$`, theEmailDeleteShouldSucceed)
}

// Given steps

func anEmailExistsInTrash(ctx context.Context, emailID, subject string) (context.Context, error) {
	w := WorldFromContext(ctx)
	e := MockEmail{
		ID:         emailID,
		Subject:    subject,
		Preview:    "preview",
		ReceivedAt: "2024-01-15T10:30:00Z",
		Size:       1234,
		Keywords:   map[string]bool{},
		MailboxIDs: map[string]bool{"mb-trash": true},
	}
	w.Emails = append(w.Emails, e)
	return ctx, nil
}

// When steps

func iDeleteEmail(ctx context.Context, emailID string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	err := client.Mail().Delete(ctx, emailID)
	w.ResultError = err
	return ctx, nil
}

// Then steps

func theEmailDeleteShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("delete operation failed: %w", w.ResultError)
	}
	return nil
}

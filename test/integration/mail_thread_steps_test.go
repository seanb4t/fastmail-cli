//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func init() {
	// Register Thread/get handler for the mock JMAP server.
	RegisterMethodHandler("Thread/get", threadGetHandler)
}

func registerMailThreadSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a thread "([^"]*)" with the following emails:$`, aThreadWithEmails)

	// When steps
	sc.Step(`^I get thread "([^"]*)"$`, iGetThread)

	// Then steps
	sc.Step(`^I should receive (\d+) thread emails?$`, iShouldReceiveNThreadEmails)
	sc.Step(`^thread email (\d+) should have subject "([^"]*)"$`, threadEmailShouldHaveSubject)
}

// Given steps

func aThreadWithEmails(ctx context.Context, threadID string, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)

	if len(table.Rows) < 2 {
		return ctx, fmt.Errorf("email table must have header and at least one data row")
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	var emails []MockEmail
	for _, row := range table.Rows[1:] {
		e := MockEmail{
			ID:         cellValue(row, headers, "id"),
			ThreadID:   threadID,
			Subject:    cellValue(row, headers, "subject"),
			ReceivedAt: cellValue(row, headers, "receivedAt"),
			Size:       1234,
			Keywords:   map[string]bool{},
			MailboxIDs: map[string]bool{"mb-inbox": true},
		}
		if fromEmail := cellValue(row, headers, "from_email"); fromEmail != "" {
			e.From = &MockAddress{
				Name:  cellValue(row, headers, "from_name"),
				Email: fromEmail,
			}
		}
		emails = append(emails, e)
	}

	w.Emails = append(w.Emails, emails...)

	// Store thread mapping in DomainData
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	threads, _ := w.DomainData["threads"].(map[string][]string)
	if threads == nil {
		threads = make(map[string][]string)
	}
	emailIDs := make([]string, len(emails))
	for i, e := range emails {
		emailIDs[i] = e.ID
	}
	threads[threadID] = emailIDs
	w.DomainData["threads"] = threads

	return ctx, nil
}

// When steps

func iGetThread(ctx context.Context, threadID string) (context.Context, error) {
	w := WorldFromContext(ctx)
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	emails, err := client.Mail().GetThread(ctx, threadID)
	w.ResultError = err
	if err == nil {
		w.DomainData["thread_emails"] = emails
	}
	return ctx, nil
}

// Then steps

func iShouldReceiveNThreadEmails(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("unexpected error: %w", w.ResultError)
	}
	emails, _ := w.DomainData["thread_emails"].([]fastmail.Email)
	if len(emails) != count {
		return fmt.Errorf("expected %d thread emails, got %d", count, len(emails))
	}
	return nil
}

func threadEmailShouldHaveSubject(ctx context.Context, index int, subject string) error {
	w := WorldFromContext(ctx)
	emails, _ := w.DomainData["thread_emails"].([]fastmail.Email)
	i := index - 1
	if i < 0 || i >= len(emails) {
		return fmt.Errorf("thread email index %d out of range (have %d)", index, len(emails))
	}
	if emails[i].Subject != subject {
		return fmt.Errorf("expected subject %q, got %q", subject, emails[i].Subject)
	}
	return nil
}

// Thread/get mock handler

func threadGetHandler(w *World, args map[string]any) map[string]any {
	var requestedIDs []string
	if ids, ok := args["ids"].([]any); ok {
		for _, id := range ids {
			requestedIDs = append(requestedIDs, id.(string))
		}
	}

	threads, _ := w.DomainData["threads"].(map[string][]string)

	var list []map[string]any
	var notFound []string
	for _, id := range requestedIDs {
		if emailIDs, ok := threads[id]; ok {
			list = append(list, map[string]any{
				"id":       id,
				"emailIds": emailIDs,
			})
		} else {
			notFound = append(notFound, id)
		}
	}

	if list == nil {
		list = []map[string]any{}
	}
	if notFound == nil {
		notFound = []string{}
	}

	return map[string]any{
		"accountId": "acc1",
		"state":     "t1",
		"list":      list,
		"notFound":  notFound,
	}
}

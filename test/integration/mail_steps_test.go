//go:build integration

package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func registerMailSteps(sc *godog.ScenarioContext) {
	// Background
	sc.Step(`^a connected Fastmail client$`, aConnectedFastmailClient)

	// Given steps
	sc.Step(`^the inbox contains the following emails:$`, theInboxContainsEmails)
	sc.Step(`^the following emails exist:$`, theFollowingEmailsExist)
	sc.Step(`^the folder "([^"]*)" contains no emails$`, theFolderContainsNoEmails)

	// When steps
	sc.Step(`^I list emails from "([^"]*)" with limit (\d+)$`, iListEmailsWithLimit)
	sc.Step(`^I search for "([^"]*)" with limit (\d+)$`, iSearchWithLimit)
	sc.Step(`^I search with filter from "([^"]*)" with limit (\d+)$`, iSearchWithFilterFrom)
	sc.Step(`^I search for "([^"]*)" with snippets and limit (\d+)$`, iSearchWithSnippets)
	sc.Step(`^I send an email to "([^"]*)" with subject "([^"]*)" and body "([^"]*)"$`, iSendAnEmail)
	sc.Step(`^I send a scheduled email to "([^"]*)" with subject "([^"]*)" and body "([^"]*)" at "([^"]*)"$`, iSendScheduledEmail)
	sc.Step(`^I set keyword "([^"]*)" to (true|false) on email "([^"]*)"$`, iSetKeyword)
	sc.Step(`^I move email "([^"]*)" to "([^"]*)"$`, iMoveEmail)

	// Then steps
	sc.Step(`^I should receive (\d+) emails?$`, iShouldReceiveNEmails)
	sc.Step(`^email (\d+) should have subject "([^"]*)"$`, emailShouldHaveSubject)
	sc.Step(`^email (\d+) should be read$`, emailShouldBeRead)
	sc.Step(`^email (\d+) should be flagged$`, emailShouldBeFlagged)
	sc.Step(`^the send should succeed$`, theSendShouldSucceed)
	sc.Step(`^the sent email ID should not be empty$`, theSentEmailIDShouldNotBeEmpty)
	sc.Step(`^the flag operation should succeed$`, theFlagOperationShouldSucceed)
	sc.Step(`^the move operation should succeed$`, theMoveOperationShouldSucceed)
	sc.Step(`^I should receive (\d+) search results? with snippets$`, iShouldReceiveSearchResultsWithSnippets)
	sc.Step(`^search result (\d+) should have a preview snippet$`, searchResultShouldHavePreviewSnippet)
}

// Background steps

func aConnectedFastmailClient(ctx context.Context) (context.Context, error) {
	// World and mock servers are already set up by Before hook
	w := WorldFromContext(ctx)
	if w == nil {
		return ctx, fmt.Errorf("world not initialized")
	}
	return ctx, nil
}

// Given steps

func theInboxContainsEmails(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	w.Emails = parseEmailTable(table, "mb-inbox")
	return ctx, nil
}

func theFollowingEmailsExist(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	w.Emails = parseEmailTable(table, "mb-inbox")
	return ctx, nil
}

func theFolderContainsNoEmails(ctx context.Context, _ string) (context.Context, error) {
	w := WorldFromContext(ctx)
	w.Emails = nil
	return ctx, nil
}

// When steps

func iListEmailsWithLimit(ctx context.Context, folder string, limit int) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	emails, err := client.Mail().List(ctx, folder, uint64(limit))
	w.ResultError = err
	w.ResultEmails = emailsToMaps(emails)
	return ctx, nil
}

func iSearchWithLimit(ctx context.Context, query string, limit int) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	emails, err := client.Mail().Search(ctx, query, uint64(limit))
	w.ResultError = err
	w.ResultEmails = emailsToMaps(emails)
	return ctx, nil
}

func iSearchWithFilterFrom(ctx context.Context, from string, limit int) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	opts := fastmail.SearchOptions{
		From:  from,
		Limit: uint64(limit),
	}
	emails, err := client.Mail().SearchAdvanced(ctx, opts)
	w.ResultError = err
	w.ResultEmails = emailsToMaps(emails)
	return ctx, nil
}

func iSearchWithSnippets(ctx context.Context, query string, limit int) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	results, err := client.Mail().SearchWithSnippets(ctx, query, uint64(limit))
	w.ResultError = err
	if err == nil {
		w.SearchResults = make([]map[string]any, len(results))
		for i, r := range results {
			w.SearchResults[i] = map[string]any{
				"id":              r.Email.ID,
				"subject":         r.Email.Subject,
				"subject_snippet": r.SubjectSnippet,
				"preview_snippet": r.PreviewSnippet,
			}
		}
	}
	return ctx, nil
}

func iSendAnEmail(ctx context.Context, to, subject, body string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	opts := fastmail.SendOptions{
		To:      []fastmail.EmailAddress{{Email: to}},
		Subject: subject,
		Body:    body,
	}
	emailID, err := client.Mail().Send(ctx, opts)
	w.ResultError = err
	w.ResultSendID = emailID
	return ctx, nil
}

func iSendScheduledEmail(ctx context.Context, to, subject, body, scheduleStr string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	scheduleTime, err := time.Parse(time.RFC3339, scheduleStr)
	if err != nil {
		return ctx, fmt.Errorf("invalid schedule time: %w", err)
	}

	opts := fastmail.SendOptions{
		To:       []fastmail.EmailAddress{{Email: to}},
		Subject:  subject,
		Body:     body,
		Schedule: &scheduleTime,
	}
	emailID, err := client.Mail().Send(ctx, opts)
	w.ResultError = err
	w.ResultSendID = emailID
	return ctx, nil
}

func iSetKeyword(ctx context.Context, keyword, valueStr, emailID string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	set := valueStr == "true"
	err := client.Mail().SetKeywords(ctx, emailID, []fastmail.KeywordAction{
		{Keyword: keyword, Set: set},
	})
	w.ResultError = err
	return ctx, nil
}

func iMoveEmail(ctx context.Context, emailID, folder string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	err := client.Mail().Move(ctx, emailID, folder)
	w.ResultError = err
	return ctx, nil
}

// Then steps

func iShouldReceiveNEmails(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("unexpected error: %w", w.ResultError)
	}
	if len(w.ResultEmails) != count {
		return fmt.Errorf("expected %d emails, got %d", count, len(w.ResultEmails))
	}
	return nil
}

func emailShouldHaveSubject(ctx context.Context, index int, subject string) error {
	w := WorldFromContext(ctx)
	// 1-based index
	i := index - 1
	if i < 0 || i >= len(w.ResultEmails) {
		return fmt.Errorf("email index %d out of range (have %d)", index, len(w.ResultEmails))
	}
	actual := w.ResultEmails[i]["subject"]
	if actual != subject {
		return fmt.Errorf("expected subject %q, got %q", subject, actual)
	}
	return nil
}

func emailShouldBeRead(ctx context.Context, index int) error {
	w := WorldFromContext(ctx)
	i := index - 1
	if i < 0 || i >= len(w.ResultEmails) {
		return fmt.Errorf("email index %d out of range", index)
	}
	isRead, _ := w.ResultEmails[i]["is_read"].(bool)
	if !isRead {
		return fmt.Errorf("expected email %d to be read", index)
	}
	return nil
}

func emailShouldBeFlagged(ctx context.Context, index int) error {
	w := WorldFromContext(ctx)
	i := index - 1
	if i < 0 || i >= len(w.ResultEmails) {
		return fmt.Errorf("email index %d out of range", index)
	}
	isFlagged, _ := w.ResultEmails[i]["is_flagged"].(bool)
	if !isFlagged {
		return fmt.Errorf("expected email %d to be flagged", index)
	}
	return nil
}

func theSendShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("send failed: %w", w.ResultError)
	}
	return nil
}

func theSentEmailIDShouldNotBeEmpty(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ResultSendID == "" {
		return fmt.Errorf("sent email ID is empty")
	}
	return nil
}

func theFlagOperationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("flag operation failed: %w", w.ResultError)
	}
	return nil
}

func theMoveOperationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("move operation failed: %w", w.ResultError)
	}
	return nil
}

func iShouldReceiveSearchResultsWithSnippets(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("unexpected error: %w", w.ResultError)
	}
	if len(w.SearchResults) != count {
		return fmt.Errorf("expected %d search results, got %d", count, len(w.SearchResults))
	}
	return nil
}

func searchResultShouldHavePreviewSnippet(ctx context.Context, index int) error {
	w := WorldFromContext(ctx)
	i := index - 1
	if i < 0 || i >= len(w.SearchResults) {
		return fmt.Errorf("search result index %d out of range", index)
	}
	snippet, _ := w.SearchResults[i]["preview_snippet"].(string)
	if snippet == "" {
		return fmt.Errorf("expected preview snippet for result %d, got empty", index)
	}
	return nil
}

// Helpers

func parseEmailTable(table *godog.Table, defaultMailboxID string) []MockEmail {
	if len(table.Rows) < 2 {
		return nil
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	var emails []MockEmail
	for _, row := range table.Rows[1:] {
		e := MockEmail{
			ID:         cellValue(row, headers, "id"),
			Subject:    cellValue(row, headers, "subject"),
			Preview:    cellValue(row, headers, "preview"),
			ReceivedAt: "2024-01-15T10:30:00Z",
			Size:       1234,
			Keywords:   map[string]bool{},
			MailboxIDs: map[string]bool{defaultMailboxID: true},
		}

		if cellValue(row, headers, "read") == "true" {
			e.Keywords["$seen"] = true
		}
		if cellValue(row, headers, "flagged") == "true" {
			e.Keywords["$flagged"] = true
		}

		emails = append(emails, e)
	}
	return emails
}

func cellValue(row *messages.PickleTableRow, headers map[string]int, key string) string {
	idx, ok := headers[key]
	if !ok || idx >= len(row.Cells) {
		return ""
	}
	return row.Cells[idx].Value
}

func emailsToMaps(emails []fastmail.Email) []map[string]any {
	result := make([]map[string]any, len(emails))
	for i, e := range emails {
		result[i] = map[string]any{
			"id":         e.ID,
			"subject":    e.Subject,
			"preview":    e.Preview,
			"is_read":    e.IsRead(),
			"is_flagged": e.IsFlagged(),
		}
	}
	return result
}

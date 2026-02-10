//go:build integration

package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// MockSubmission holds test data for a scheduled email submission.
type MockSubmission struct {
	ID         string
	EmailID    string
	SendAt     string
	UndoStatus string
}

func init() {
	// EmailSubmission/query — returns IDs of mock submissions
	RegisterMethodHandler("EmailSubmission/query", func(w *World, _ map[string]any) map[string]any {
		subs := getMockSubmissions(w)
		ids := make([]string, len(subs))
		for i, s := range subs {
			ids[i] = s.ID
		}
		return map[string]any{
			"accountId":           "acc1",
			"queryState":          "sq1",
			"canCalculateChanges": true,
			"position":            0,
			"ids":                 ids,
			"total":               len(ids),
		}
	})

	// EmailSubmission/get — returns full submission objects
	RegisterMethodHandler("EmailSubmission/get", func(w *World, _ map[string]any) map[string]any {
		subs := getMockSubmissions(w)
		list := make([]map[string]any, len(subs))
		for i, s := range subs {
			list[i] = map[string]any{
				"id":         s.ID,
				"emailId":    s.EmailID,
				"sendAt":     s.SendAt,
				"undoStatus": s.UndoStatus,
			}
		}
		return map[string]any{
			"accountId": "acc1",
			"state":     "sg1",
			"list":      list,
			"notFound":  []any{},
		}
	})

	// EmailSubmission/set — handle destroy for CancelScheduled
	RegisterMethodHandler("EmailSubmission/set", func(_ *World, args map[string]any) map[string]any {
		resp := map[string]any{
			"accountId": "acc1",
			"oldState":  "sub1",
			"newState":  "sub2",
		}

		if destroy, ok := args["destroy"].([]any); ok {
			resp["destroyed"] = destroy
		}

		// Also support creation (for Send, which uses this method too)
		if create, ok := args["create"].(map[string]any); ok {
			created := make(map[string]any)
			for clientID := range create {
				created[clientID] = map[string]any{
					"id": "submission-" + clientID,
				}
			}
			resp["created"] = created
		}

		return resp
	})

	// Email/import — handle import operations
	RegisterMethodHandler("Email/import", func(_ *World, _ map[string]any) map[string]any {
		return map[string]any{
			"accountId": "acc1",
			"oldState":  "ei1",
			"newState":  "ei2",
			"created": map[string]any{
				"imp1": map[string]any{
					"id":       "imported-email-1",
					"blobId":   "blob-imported-1",
					"threadId": "thread-imported-1",
					"size":     float64(1024),
				},
			},
			"notCreated": map[string]any{},
		}
	})
}

func getMockSubmissions(w *World) []MockSubmission {
	if w.DomainData == nil {
		return nil
	}
	subs, _ := w.DomainData["mock_submissions"].([]MockSubmission)
	return subs
}

// scheduledState holds per-scenario state for scheduled email tests.
type scheduledState struct {
	Results []fastmail.ScheduledSend
}

func getScheduledState(w *World) *scheduledState {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	if s, ok := w.DomainData["scheduled_results"].(*scheduledState); ok {
		return s
	}
	s := &scheduledState{}
	w.DomainData["scheduled_results"] = s
	return s
}

// importState holds per-scenario state for import tests.
type importState struct {
	Result *fastmail.ImportResult
}

func getImportState(w *World) *importState {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	if s, ok := w.DomainData["import_result"].(*importState); ok {
		return s
	}
	s := &importState{}
	w.DomainData["import_result"] = s
	return s
}

func registerMailScheduledSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the following scheduled submissions exist:$`, theFollowingScheduledSubmissionsExist)
	sc.Step(`^a blob upload server is configured$`, aBlobUploadServerIsConfigured)

	// When steps
	sc.Step(`^I list scheduled emails with limit (\d+)$`, iListScheduledEmails)
	sc.Step(`^I cancel scheduled email "([^"]*)"$`, iCancelScheduledEmail)
	sc.Step(`^I import an email with content "([^"]*)" to folder "([^"]*)"$`, iImportEmail)

	// Then steps
	sc.Step(`^I should receive (\d+) scheduled emails$`, iShouldReceiveScheduledEmails)
	sc.Step(`^scheduled email (\d+) should have subject "([^"]*)"$`, scheduledEmailShouldHaveSubject)
	sc.Step(`^scheduled email (\d+) should have status "([^"]*)"$`, scheduledEmailShouldHaveStatus)
	sc.Step(`^the cancel operation should succeed$`, theCancelOperationShouldSucceed)
	sc.Step(`^the import should succeed$`, theImportShouldSucceed)
	sc.Step(`^the imported email ID should not be empty$`, theImportedEmailIDShouldNotBeEmpty)
}

// Given steps

func theFollowingScheduledSubmissionsExist(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	subs := make([]MockSubmission, 0, len(table.Rows)-1)
	for _, row := range table.Rows[1:] {
		subs = append(subs, MockSubmission{
			ID:         cellValue(row, headers, "id"),
			EmailID:    cellValue(row, headers, "emailId"),
			SendAt:     cellValue(row, headers, "sendAt"),
			UndoStatus: cellValue(row, headers, "undoStatus"),
		})
	}
	w.DomainData["mock_submissions"] = subs
	return ctx, nil
}

func aBlobUploadServerIsConfigured(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	setupBlobServer(w, nil)
	return ctx, nil
}

// When steps

func iListScheduledEmails(ctx context.Context, limit int) (context.Context, error) {
	w := WorldFromContext(ctx)
	state := getScheduledState(w)

	client := fastmail.NewClient(w.SessionServer.URL, "test-token")
	results, err := client.Mail().ListScheduled(ctx, uint64(limit)) //nolint:gosec // test data, no overflow risk
	w.ResultError = err
	state.Results = results
	return ctx, nil
}

func iCancelScheduledEmail(ctx context.Context, submissionID string) (context.Context, error) {
	w := WorldFromContext(ctx)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")
	w.ResultError = client.Mail().CancelScheduled(ctx, submissionID)
	return ctx, nil
}

func iImportEmail(ctx context.Context, content, folder string) (context.Context, error) {
	w := WorldFromContext(ctx)
	state := getImportState(w)

	client := fastmail.NewClient(w.SessionServer.URL, "test-token")
	result, err := client.Mail().Import(ctx, strings.NewReader(content), fastmail.ImportOptions{
		Folder: folder,
	})
	w.ResultError = err
	state.Result = result
	return ctx, nil
}

// Then steps

func iShouldReceiveScheduledEmails(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("unexpected error: %w", w.ResultError)
	}
	state := getScheduledState(w)
	if len(state.Results) != count {
		return fmt.Errorf("expected %d scheduled emails, got %d", count, len(state.Results))
	}
	return nil
}

func scheduledEmailShouldHaveSubject(ctx context.Context, index int, subject string) error {
	w := WorldFromContext(ctx)
	state := getScheduledState(w)
	i := index - 1
	if i < 0 || i >= len(state.Results) {
		return fmt.Errorf("scheduled email index %d out of range (have %d)", index, len(state.Results))
	}
	if state.Results[i].Subject != subject {
		return fmt.Errorf("expected subject %q, got %q", subject, state.Results[i].Subject)
	}
	return nil
}

func scheduledEmailShouldHaveStatus(ctx context.Context, index int, status string) error {
	w := WorldFromContext(ctx)
	state := getScheduledState(w)
	i := index - 1
	if i < 0 || i >= len(state.Results) {
		return fmt.Errorf("scheduled email index %d out of range (have %d)", index, len(state.Results))
	}
	if state.Results[i].UndoStatus != status {
		return fmt.Errorf("expected status %q, got %q", status, state.Results[i].UndoStatus)
	}
	return nil
}

func theCancelOperationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("cancel failed: %w", w.ResultError)
	}
	return nil
}

func theImportShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("import failed: %w", w.ResultError)
	}
	return nil
}

func theImportedEmailIDShouldNotBeEmpty(ctx context.Context) error {
	w := WorldFromContext(ctx)
	state := getImportState(w)
	if state.Result == nil {
		return fmt.Errorf("import result is nil")
	}
	if state.Result.ID == "" {
		return fmt.Errorf("imported email ID is empty")
	}
	return nil
}

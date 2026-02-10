//go:build integration

package integration

import (
	"context"

	"github.com/cucumber/godog"
)

func registerMCPMailScheduledSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the following scheduled emails exist for MCP:$`, theFollowingScheduledEmailsExistForMCP)
}

// theFollowingScheduledEmailsExistForMCP populates both the mock submission data
// (for EmailSubmission/query and EmailSubmission/get handlers) and the mock email
// data (for the Email/get handler that hydrates subjects and recipients).
func theFollowingScheduledEmailsExistForMCP(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	subs := make([]MockSubmission, 0, len(table.Rows)-1)
	emails := make([]MockEmail, 0, len(table.Rows)-1)

	for _, row := range table.Rows[1:] {
		sub := MockSubmission{
			ID:         cellValue(row, headers, "id"),
			EmailID:    cellValue(row, headers, "emailId"),
			SendAt:     cellValue(row, headers, "sendAt"),
			UndoStatus: cellValue(row, headers, "undoStatus"),
		}
		subs = append(subs, sub)

		subject := cellValue(row, headers, "subject")
		toAddr := cellValue(row, headers, "to")
		email := MockEmail{
			ID:      sub.EmailID,
			Subject: subject,
		}
		if toAddr != "" {
			email.To = []MockAddress{{Email: toAddr}}
		}
		emails = append(emails, email)
	}

	w.DomainData["mock_submissions"] = subs
	w.Emails = append(w.Emails, emails...)

	return ctx, nil
}

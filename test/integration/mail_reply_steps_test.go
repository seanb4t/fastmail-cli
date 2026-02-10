//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func registerMailReplySteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^an email exists for reply:$`, anEmailExistsForReply)
	sc.Step(`^an email exists for reply with cc:$`, anEmailExistsForReplyWithCc)

	// When steps
	sc.Step(`^I reply to email "([^"]*)" with body "([^"]*)"$`, iReplyToEmail)
	sc.Step(`^I reply all to email "([^"]*)" with body "([^"]*)"$`, iReplyAllToEmail)

	// Then steps
	sc.Step(`^the reply should succeed$`, theReplyShouldSucceed)
	sc.Step(`^the reply email ID should not be empty$`, theReplyEmailIDShouldNotBeEmpty)
}

// Given steps

func anEmailExistsForReply(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)

	if len(table.Rows) < 2 {
		return ctx, fmt.Errorf("email table must have header and at least one data row")
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	for _, row := range table.Rows[1:] {
		e := MockEmail{
			ID:         cellValue(row, headers, "id"),
			Subject:    cellValue(row, headers, "subject"),
			ReceivedAt: "2024-01-15T10:30:00Z",
			Size:       1234,
			Keywords:   map[string]bool{},
			MailboxIDs: map[string]bool{"mb-inbox": true},
			Body:       "Original message body",
		}
		if fromEmail := cellValue(row, headers, "from_email"); fromEmail != "" {
			e.From = &MockAddress{
				Name:  cellValue(row, headers, "from_name"),
				Email: fromEmail,
			}
		}
		if msgID := cellValue(row, headers, "messageId"); msgID != "" {
			e.MessageID = []string{msgID}
		}
		w.Emails = append(w.Emails, e)
	}

	return ctx, nil
}

func anEmailExistsForReplyWithCc(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)

	if len(table.Rows) < 2 {
		return ctx, fmt.Errorf("email table must have header and at least one data row")
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	for _, row := range table.Rows[1:] {
		e := MockEmail{
			ID:         cellValue(row, headers, "id"),
			Subject:    cellValue(row, headers, "subject"),
			ReceivedAt: "2024-01-15T10:30:00Z",
			Size:       1234,
			Keywords:   map[string]bool{},
			MailboxIDs: map[string]bool{"mb-inbox": true},
			Body:       "Original message body",
		}
		if fromEmail := cellValue(row, headers, "from_email"); fromEmail != "" {
			e.From = &MockAddress{
				Name:  cellValue(row, headers, "from_name"),
				Email: fromEmail,
			}
		}
		if toEmail := cellValue(row, headers, "to_email"); toEmail != "" {
			e.To = []MockAddress{{Email: toEmail}}
		}
		if ccEmail := cellValue(row, headers, "cc_email"); ccEmail != "" {
			e.Cc = []MockAddress{{Email: ccEmail}}
		}
		if msgID := cellValue(row, headers, "messageId"); msgID != "" {
			e.MessageID = []string{msgID}
		}
		w.Emails = append(w.Emails, e)
	}

	return ctx, nil
}

// When steps

func iReplyToEmail(ctx context.Context, emailID, body string) (context.Context, error) {
	w := WorldFromContext(ctx)
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	replyID, err := client.Mail().Reply(ctx, fastmail.ReplyOptions{
		EmailID:  emailID,
		Body:     body,
		ReplyAll: false,
	})
	w.ResultError = err
	w.DomainData["reply_email_id"] = replyID
	return ctx, nil
}

func iReplyAllToEmail(ctx context.Context, emailID, body string) (context.Context, error) {
	w := WorldFromContext(ctx)
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	replyID, err := client.Mail().Reply(ctx, fastmail.ReplyOptions{
		EmailID:  emailID,
		Body:     body,
		ReplyAll: true,
	})
	w.ResultError = err
	w.DomainData["reply_email_id"] = replyID
	return ctx, nil
}

// Then steps

func theReplyShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	if w.ResultError != nil {
		return fmt.Errorf("reply failed: %w", w.ResultError)
	}
	return nil
}

func theReplyEmailIDShouldNotBeEmpty(ctx context.Context) error {
	w := WorldFromContext(ctx)
	replyID, _ := w.DomainData["reply_email_id"].(string)
	if replyID == "" {
		return fmt.Errorf("reply email ID is empty")
	}
	return nil
}

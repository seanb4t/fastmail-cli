//go:build integration

package integration

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// mailboxResult holds the captured results from mailbox operations.
type mailboxResult struct {
	mailboxes []fastmail.Mailbox
	created   *fastmail.Mailbox
	err       error
}

func getMailboxResult(w *World) *mailboxResult {
	if w.DomainData == nil {
		w.DomainData = map[string]any{}
	}
	r, ok := w.DomainData["mailbox"].(*mailboxResult)
	if !ok {
		r = &mailboxResult{}
		w.DomainData["mailbox"] = r
	}
	return r
}

func registerMailboxSteps(sc *godog.ScenarioContext) {
	// When steps
	sc.Step(`^I list all mailboxes$`, iListAllMailboxes)
	sc.Step(`^I create a mailbox named "([^"]*)"$`, iCreateMailbox)
	sc.Step(`^I create a mailbox named "([^"]*)" under parent "([^"]*)"$`, iCreateMailboxWithParent)
	sc.Step(`^I rename mailbox "([^"]*)" to "([^"]*)"$`, iRenameMailbox)
	sc.Step(`^I delete mailbox "([^"]*)"$`, iDeleteMailbox)

	// Then steps
	sc.Step(`^I should receive (\d+) mailboxes$`, iShouldReceiveNMailboxes)
	sc.Step(`^the mailbox list should contain "([^"]*)"$`, theMailboxListShouldContain)
	sc.Step(`^the mailbox creation should succeed$`, theMailboxCreationShouldSucceed)
	sc.Step(`^the created mailbox ID should not be empty$`, theCreatedMailboxIDShouldNotBeEmpty)
	sc.Step(`^the rename operation should succeed$`, theRenameOperationShouldSucceed)
	sc.Step(`^the delete operation should succeed$`, theDeleteOperationShouldSucceed)
}

// When steps

func iListAllMailboxes(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	r := getMailboxResult(w)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	mailboxes, err := client.Mailbox().List(ctx)
	r.mailboxes = mailboxes
	r.err = err
	return ctx, nil
}

func iCreateMailbox(ctx context.Context, name string) (context.Context, error) {
	w := WorldFromContext(ctx)
	r := getMailboxResult(w)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	mb, err := client.Mailbox().Create(ctx, name, "")
	r.created = mb
	r.err = err
	return ctx, nil
}

func iCreateMailboxWithParent(ctx context.Context, name, parentID string) (context.Context, error) {
	w := WorldFromContext(ctx)
	r := getMailboxResult(w)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	mb, err := client.Mailbox().Create(ctx, name, parentID)
	r.created = mb
	r.err = err
	return ctx, nil
}

func iRenameMailbox(ctx context.Context, id, newName string) (context.Context, error) {
	w := WorldFromContext(ctx)
	r := getMailboxResult(w)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	r.err = client.Mailbox().Rename(ctx, id, newName)
	return ctx, nil
}

func iDeleteMailbox(ctx context.Context, id string) (context.Context, error) {
	w := WorldFromContext(ctx)
	r := getMailboxResult(w)
	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	r.err = client.Mailbox().Delete(ctx, id)
	return ctx, nil
}

// Then steps

func iShouldReceiveNMailboxes(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	r := getMailboxResult(w)
	if r.err != nil {
		return fmt.Errorf("unexpected error: %w", r.err)
	}
	if len(r.mailboxes) != count {
		return fmt.Errorf("expected %d mailboxes, got %d", count, len(r.mailboxes))
	}
	return nil
}

func theMailboxListShouldContain(ctx context.Context, name string) error {
	w := WorldFromContext(ctx)
	r := getMailboxResult(w)
	for _, mb := range r.mailboxes {
		if mb.Name == name {
			return nil
		}
	}
	return fmt.Errorf("mailbox %q not found in list", name)
}

func theMailboxCreationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getMailboxResult(w)
	if r.err != nil {
		return fmt.Errorf("mailbox creation failed: %w", r.err)
	}
	return nil
}

func theCreatedMailboxIDShouldNotBeEmpty(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getMailboxResult(w)
	if r.created == nil {
		return fmt.Errorf("no created mailbox in result")
	}
	if r.created.ID == "" {
		return fmt.Errorf("created mailbox ID is empty")
	}
	return nil
}

func theRenameOperationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getMailboxResult(w)
	if r.err != nil {
		return fmt.Errorf("rename operation failed: %w", r.err)
	}
	return nil
}

func theDeleteOperationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	r := getMailboxResult(w)
	if r.err != nil {
		return fmt.Errorf("delete operation failed: %w", r.err)
	}
	return nil
}

package fastmail

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// MailboxService provides mailbox management operations.
type MailboxService struct {
	client *Client
}

// List returns all mailboxes for the account.
//
//nolint:dupl // JMAP service pattern - structural similarity with other services is intentional
func (s *MailboxService) List(ctx context.Context) ([]Mailbox, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	getBuilder := jmap.NewMailboxGet(accountID)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Mailbox/get", getBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("mailbox get failed: %s", result.Error())
	}

	var getResp jmap.MailboxGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding mailbox response")
	}

	return convertMailboxes(getResp.List), nil
}

// Create creates a new mailbox with the given name.
// If parentID is non-empty, the mailbox is created as a child of the specified parent.
func (s *MailboxService) Create(ctx context.Context, name string, parentID string) (*Mailbox, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	mbData := map[string]any{
		"name": name,
	}
	if parentID != "" {
		mbData["parentId"] = parentID
	}

	setBuilder := jmap.NewMailboxSet(accountID).Create("new-mailbox", mbData)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Mailbox/set", setBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("mailbox set failed: %s", result.Error())
	}

	var setResp jmap.MailboxSetResponse
	if err := result.Decode(&setResp); err != nil {
		return nil, oops.Wrapf(err, "decoding mailbox set response")
	}

	if errInfo, ok := setResp.NotCreated["new-mailbox"]; ok {
		return nil, oops.Errorf("failed to create mailbox: %s", errInfo.Error())
	}

	created, ok := setResp.Created["new-mailbox"]
	if !ok {
		return nil, oops.Errorf("mailbox not returned in created map")
	}

	mb := convertMailbox(created)
	return &mb, nil
}

// Rename renames an existing mailbox.
func (s *MailboxService) Rename(ctx context.Context, id string, name string) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setBuilder := jmap.NewMailboxSet(accountID).Update(id, map[string]any{
		"name": name,
	})

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Mailbox/set", setBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return oops.Errorf("mailbox set failed: %s", result.Error())
	}

	var setResp jmap.MailboxSetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding mailbox set response")
	}

	if errInfo, ok := setResp.NotUpdated[id]; ok {
		return oops.Errorf("failed to rename mailbox: %s", errInfo.Error())
	}

	return nil
}

// Delete destroys a mailbox by ID.
func (s *MailboxService) Delete(ctx context.Context, id string) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setBuilder := jmap.NewMailboxSet(accountID).Destroy(id)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Mailbox/set", setBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return oops.Errorf("mailbox set failed: %s", result.Error())
	}

	var setResp jmap.MailboxSetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding mailbox set response")
	}

	if errInfo, ok := setResp.NotDestroyed[id]; ok {
		return oops.Errorf("failed to delete mailbox: %s", errInfo.Error())
	}

	return nil
}

// convertMailboxes converts JMAP mailboxes to domain mailboxes.
func convertMailboxes(jmapMailboxes []jmap.Mailbox) []Mailbox {
	mailboxes := make([]Mailbox, len(jmapMailboxes))
	for i, jmb := range jmapMailboxes {
		mailboxes[i] = convertMailbox(jmb)
	}
	return mailboxes
}

// convertMailbox converts a single JMAP mailbox to a domain mailbox.
func convertMailbox(jmb jmap.Mailbox) Mailbox {
	return Mailbox{
		ID:            jmb.ID,
		Name:          jmb.Name,
		Role:          MailboxRole(jmb.Role),
		ParentID:      jmb.ParentID,
		SortOrder:     jmb.SortOrder,
		TotalEmails:   jmb.TotalEmails,
		UnreadEmails:  jmb.UnreadEmails,
		TotalThreads:  jmb.TotalThreads,
		UnreadThreads: jmb.UnreadThreads,
	}
}

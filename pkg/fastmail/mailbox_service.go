package fastmail

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// MailboxService provides mailbox (folder) operations.
type MailboxService struct {
	client *Client
}

// List returns all mailboxes.
func (s *MailboxService) List(ctx context.Context) ([]Mailbox, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "listing mailboxes")
	}

	getBuilder := jmap.NewMailboxGet(accountID).
		Properties("id", "name", "role", "parentId", "totalEmails", "unreadEmails", "totalThreads", "unreadThreads")
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
		return nil, oops.Wrapf(err, "decoding response")
	}

	return convertMailboxes(getResp.List), nil
}

// Create creates a new mailbox with the given name. If parentID is non-empty,
// the mailbox is created as a child of that parent. Returns the created mailbox.
func (s *MailboxService) Create(ctx context.Context, name, parentID string) (*Mailbox, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "creating mailbox")
	}

	data := map[string]any{"name": name}
	if parentID != "" {
		data["parentId"] = parentID
	}

	setBuilder := jmap.NewMailboxSet(accountID).Create("mb1", data)
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
		return nil, oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotCreated["mb1"]; ok {
		return nil, oops.Errorf("failed to create mailbox: %s", errInfo.Error())
	}

	created, ok := setResp.Created["mb1"]
	if !ok {
		return nil, oops.Errorf("mailbox not returned in created map")
	}

	mb := convertMailbox(created)
	mb.Name = name

	return &mb, nil
}

// Rename changes the name of a mailbox.
func (s *MailboxService) Rename(ctx context.Context, id, newName string) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return oops.Wrapf(err, "renaming mailbox")
	}

	setBuilder := jmap.NewMailboxSet(accountID).Update(id, map[string]any{"name": newName})
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
		return oops.Wrapf(err, "decoding response")
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
		return oops.Wrapf(err, "deleting mailbox")
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
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotDestroyed[id]; ok {
		return oops.Errorf("failed to delete mailbox: %s", errInfo.Error())
	}

	return nil
}

func convertMailboxes(jmapMailboxes []jmap.Mailbox) []Mailbox {
	result := make([]Mailbox, len(jmapMailboxes))
	for i, mb := range jmapMailboxes {
		result[i] = convertMailbox(mb)
	}
	return result
}

func convertMailbox(mb jmap.Mailbox) Mailbox {
	return Mailbox{
		ID:            mb.ID,
		Name:          mb.Name,
		Role:          MailboxRole(mb.Role),
		ParentID:      mb.ParentID,
		TotalEmails:   mb.TotalEmails,
		UnreadEmails:  mb.UnreadEmails,
		TotalThreads:  mb.TotalThreads,
		UnreadThreads: mb.UnreadThreads,
	}
}

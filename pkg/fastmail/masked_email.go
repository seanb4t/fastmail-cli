package fastmail

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// MaskedEmailState represents the state of a masked email address.
type MaskedEmailState string

// MaskedEmailState values for masked email lifecycle.
const (
	MaskedEmailStateEnabled  MaskedEmailState = "enabled"
	MaskedEmailStateDisabled MaskedEmailState = "disabled"
	MaskedEmailStatePending  MaskedEmailState = "pending"
	MaskedEmailStateDeleted  MaskedEmailState = "deleted"
)

// MaskedEmail represents a Fastmail masked email address.
type MaskedEmail struct {
	ID            string           `json:"id"`
	Email         string           `json:"email"`
	State         MaskedEmailState `json:"state"`
	ForDomain     string           `json:"forDomain,omitempty"`
	Description   string           `json:"description,omitempty"`
	URL           string           `json:"url,omitempty"`
	CreatedBy     string           `json:"createdBy,omitempty"`
	CreatedAt     string           `json:"createdAt,omitempty"`
	LastMessageAt string           `json:"lastMessageAt,omitempty"`
}

// MaskedEmailService provides masked email operations.
type MaskedEmailService struct {
	client *Client
}

// CreateMaskedEmailOptions specifies options for creating a masked email.
type CreateMaskedEmailOptions struct {
	ForDomain   string
	Description string
}

// List returns all masked emails.
//
//nolint:dupl // JMAP service pattern - structural similarity with other services is intentional
func (s *MaskedEmailService) List(ctx context.Context) ([]MaskedEmail, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	getBuilder := jmap.NewMaskedEmailGet(accountID)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.MaskedEmailCapability)
	callID := req.Invoke("MaskedEmail/get", getBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("get failed: %s", result.Error())
	}

	var getResp jmap.MaskedEmailGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	return convertMaskedEmails(getResp.List), nil
}

// Get returns a single masked email by ID.
func (s *MaskedEmailService) Get(ctx context.Context, id string) (*MaskedEmail, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	getBuilder := jmap.NewMaskedEmailGet(accountID).IDs(id)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.MaskedEmailCapability)
	callID := req.Invoke("MaskedEmail/get", getBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("get failed: %s", result.Error())
	}

	var getResp jmap.MaskedEmailGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if len(getResp.List) == 0 {
		return nil, oops.Errorf("masked email not found: %s", id)
	}

	return convertMaskedEmail(&getResp.List[0]), nil
}

// Create creates a new masked email address.
func (s *MaskedEmailService) Create(ctx context.Context, opts CreateMaskedEmailOptions) (*MaskedEmail, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	setBuilder := jmap.NewMaskedEmailSet(accountID).
		Create("new", opts.ForDomain, opts.Description, jmap.MaskedEmailStateEnabled)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.MaskedEmailCapability)
	callID := req.Invoke("MaskedEmail/set", setBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("set failed: %s", result.Error())
	}

	var setResp jmap.MaskedEmailSetResponse
	if err := result.Decode(&setResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotCreated["new"]; ok {
		return nil, oops.Errorf("failed to create masked email: %s - %s", errInfo.Type, errInfo.Description)
	}

	created, ok := setResp.Created["new"]
	if !ok || created == nil {
		return nil, oops.Errorf("masked email not returned in created map")
	}

	return convertMaskedEmail(created), nil
}

// Enable enables a masked email address.
func (s *MaskedEmailService) Enable(ctx context.Context, id string) error {
	return s.updateState(ctx, id, jmap.MaskedEmailStateEnabled)
}

// Disable disables a masked email address.
func (s *MaskedEmailService) Disable(ctx context.Context, id string) error {
	return s.updateState(ctx, id, jmap.MaskedEmailStateDisabled)
}

// updateState updates the state of a masked email.
func (s *MaskedEmailService) updateState(ctx context.Context, id string, state jmap.MaskedEmailState) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setBuilder := jmap.NewMaskedEmailSet(accountID).
		Update(id, map[string]any{"state": state})

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.MaskedEmailCapability)
	callID := req.Invoke("MaskedEmail/set", setBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return oops.Errorf("set failed: %s", result.Error())
	}

	var setResp jmap.MaskedEmailSetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotUpdated[id]; ok {
		return oops.Errorf("failed to update masked email: %s - %s", errInfo.Type, errInfo.Description)
	}

	return nil
}

// Delete permanently deletes a masked email address.
//
//nolint:dupl // JMAP service pattern - structural similarity with other services is intentional
func (s *MaskedEmailService) Delete(ctx context.Context, id string) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setBuilder := jmap.NewMaskedEmailSet(accountID).Destroy(id)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.MaskedEmailCapability)
	callID := req.Invoke("MaskedEmail/set", setBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return oops.Errorf("set failed: %s", result.Error())
	}

	var setResp jmap.MaskedEmailSetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotDestroyed[id]; ok {
		return oops.Errorf("failed to delete masked email: %s - %s", errInfo.Type, errInfo.Description)
	}

	return nil
}

// convertMaskedEmails converts JMAP masked emails to domain masked emails.
func convertMaskedEmails(jmapMEs []jmap.MaskedEmail) []MaskedEmail {
	result := make([]MaskedEmail, len(jmapMEs))
	for i, me := range jmapMEs {
		result[i] = MaskedEmail{
			ID:            me.ID,
			Email:         me.Email,
			State:         MaskedEmailState(me.State),
			ForDomain:     me.ForDomain,
			Description:   me.Description,
			URL:           me.URL,
			CreatedBy:     me.CreatedBy,
			CreatedAt:     me.CreatedAt,
			LastMessageAt: me.LastMessageAt,
		}
	}
	return result
}

// convertMaskedEmail converts a single JMAP masked email to domain type.
func convertMaskedEmail(me *jmap.MaskedEmail) *MaskedEmail {
	if me == nil {
		return nil
	}
	return &MaskedEmail{
		ID:            me.ID,
		Email:         me.Email,
		State:         MaskedEmailState(me.State),
		ForDomain:     me.ForDomain,
		Description:   me.Description,
		URL:           me.URL,
		CreatedBy:     me.CreatedBy,
		CreatedAt:     me.CreatedAt,
		LastMessageAt: me.LastMessageAt,
	}
}

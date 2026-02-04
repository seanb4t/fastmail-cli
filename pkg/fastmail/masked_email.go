package fastmail

import (
	"context"
	"time"

	"github.com/samber/oops"
	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// CapMaskedEmail is the JMAP capability URI for Fastmail Masked Email.
const CapMaskedEmail = "https://www.fastmail.com/dev/maskedemail"

// MaskedEmailState represents the state of a masked email address.
type MaskedEmailState string

const (
	// MaskedEmailStateEnabled indicates the masked email is active and receiving mail.
	MaskedEmailStateEnabled MaskedEmailState = "enabled"

	// MaskedEmailStateDisabled indicates the masked email is blocked from receiving mail.
	MaskedEmailStateDisabled MaskedEmailState = "disabled"

	// MaskedEmailStateDeleted indicates the masked email has been deleted.
	MaskedEmailStateDeleted MaskedEmailState = "deleted"
)

// MaskedEmail represents a Fastmail masked email address.
//
// Masked emails are unique, auto-generated email addresses that forward
// to your real inbox. They help protect your privacy by hiding your
// real email address from websites and services.
type MaskedEmail struct {
	// ID is the unique identifier for this masked email.
	ID string

	// Email is the generated masked email address.
	Email string

	// State indicates whether the masked email is enabled, disabled, or deleted.
	State MaskedEmailState

	// ForDomain is the domain this masked email was created for.
	ForDomain string

	// Description is a user-provided description for this masked email.
	Description string

	// URL is an optional URL associated with this masked email.
	URL string

	// CreatedAt is when this masked email was created.
	CreatedAt time.Time

	// CreatedBy indicates what created this masked email (e.g., "user", "1password", "api").
	CreatedBy string

	// LastMessageAt is when the last email was received at this address.
	// Zero value indicates no messages have been received.
	LastMessageAt time.Time
}

// IsEnabled reports whether this masked email is enabled.
func (m *MaskedEmail) IsEnabled() bool {
	return m.State == MaskedEmailStateEnabled
}

// HasReceivedMail reports whether this masked email has received any messages.
func (m *MaskedEmail) HasReceivedMail() bool {
	return !m.LastMessageAt.IsZero()
}

// MaskedEmailService provides masked email operations.
type MaskedEmailService struct {
	client *Client
}

// MaskedEmail returns the masked email service for managing masked email addresses.
func (c *Client) MaskedEmail() *MaskedEmailService {
	return &MaskedEmailService{
		client: c,
	}
}

// List returns all masked emails for the account.
func (s *MaskedEmailService) List(ctx context.Context) ([]MaskedEmail, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	getArgs := map[string]any{
		"accountId": accountID,
		"ids":       nil, // nil returns all
	}

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, CapMaskedEmail)
	callID := req.Invoke("MaskedEmail/get", getArgs)

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

	var getResp maskedEmailGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	return convertMaskedEmails(getResp.List), nil
}

// Create creates a new masked email for the given domain.
func (s *MaskedEmailService) Create(ctx context.Context, forDomain string) (*MaskedEmail, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	setArgs := map[string]any{
		"accountId": accountID,
		"create": map[string]any{
			"new-masked": map[string]any{
				"state":     "enabled",
				"forDomain": forDomain,
			},
		},
	}

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, CapMaskedEmail)
	callID := req.Invoke("MaskedEmail/set", setArgs)

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

	var setResp maskedEmailSetResponse
	if err := result.Decode(&setResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotCreated["new-masked"]; ok {
		return nil, oops.Errorf("failed to create masked email: %s", errInfo.Error())
	}

	created, ok := setResp.Created["new-masked"]
	if !ok {
		return nil, oops.Errorf("masked email not returned in created map")
	}

	maskedEmail := convertJMAPMaskedEmail(created)
	return &maskedEmail, nil
}

// Enable sets the masked email state to enabled.
func (s *MaskedEmailService) Enable(ctx context.Context, id string) error {
	return s.setState(ctx, id, MaskedEmailStateEnabled)
}

// Disable sets the masked email state to disabled.
func (s *MaskedEmailService) Disable(ctx context.Context, id string) error {
	return s.setState(ctx, id, MaskedEmailStateDisabled)
}

// setState updates the state of a masked email.
func (s *MaskedEmailService) setState(ctx context.Context, id string, state MaskedEmailState) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setArgs := map[string]any{
		"accountId": accountID,
		"update": map[string]any{
			id: map[string]any{
				"state": string(state),
			},
		},
	}

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, CapMaskedEmail)
	callID := req.Invoke("MaskedEmail/set", setArgs)

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

	var setResp maskedEmailSetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotUpdated[id]; ok {
		return oops.Errorf("failed to update masked email: %s", errInfo.Error())
	}

	return nil
}

// Delete permanently removes a masked email.
func (s *MaskedEmailService) Delete(ctx context.Context, id string) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setArgs := map[string]any{
		"accountId": accountID,
		"destroy":   []string{id},
	}

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, CapMaskedEmail)
	callID := req.Invoke("MaskedEmail/set", setArgs)

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return oops.Errorf("destroy failed: %s", result.Error())
	}

	var setResp maskedEmailSetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotDestroyed[id]; ok {
		return oops.Errorf("failed to delete masked email: %s", errInfo.Error())
	}

	return nil
}

// jmapMaskedEmail represents the JMAP protocol structure for a masked email.
type jmapMaskedEmail struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	State         string `json:"state"`
	ForDomain     string `json:"forDomain"`
	Description   string `json:"description"`
	URL           string `json:"url"`
	CreatedAt     string `json:"createdAt"`
	CreatedBy     string `json:"createdBy"`
	LastMessageAt string `json:"lastMessageAt"`
}

// maskedEmailGetResponse represents the response from MaskedEmail/get.
type maskedEmailGetResponse struct {
	AccountID string            `json:"accountId"`
	State     string            `json:"state"`
	List      []jmapMaskedEmail `json:"list"`
	NotFound  []string          `json:"notFound"`
}

// maskedEmailSetResponse represents the response from MaskedEmail/set.
type maskedEmailSetResponse struct {
	AccountID    string                      `json:"accountId"`
	OldState     string                      `json:"oldState"`
	NewState     string                      `json:"newState"`
	Created      map[string]jmapMaskedEmail  `json:"created"`
	NotCreated   map[string]jmap.MethodError `json:"notCreated"`
	Updated      map[string]any              `json:"updated"`
	NotUpdated   map[string]jmap.MethodError `json:"notUpdated"`
	Destroyed    []string                    `json:"destroyed"`
	NotDestroyed map[string]jmap.MethodError `json:"notDestroyed"`
}

// convertMaskedEmails converts JMAP masked emails to domain types.
func convertMaskedEmails(jmapEmails []jmapMaskedEmail) []MaskedEmail {
	emails := make([]MaskedEmail, len(jmapEmails))
	for i, je := range jmapEmails {
		emails[i] = convertJMAPMaskedEmail(je)
	}
	return emails
}

// convertJMAPMaskedEmail converts a single JMAP masked email to the domain type.
func convertJMAPMaskedEmail(je jmapMaskedEmail) MaskedEmail {
	var createdAt time.Time
	if je.CreatedAt != "" {
		createdAt, _ = time.Parse(time.RFC3339, je.CreatedAt)
	}

	var lastMessageAt time.Time
	if je.LastMessageAt != "" {
		lastMessageAt, _ = time.Parse(time.RFC3339, je.LastMessageAt)
	}

	return MaskedEmail{
		ID:            je.ID,
		Email:         je.Email,
		State:         MaskedEmailState(je.State),
		ForDomain:     je.ForDomain,
		Description:   je.Description,
		URL:           je.URL,
		CreatedAt:     createdAt,
		CreatedBy:     je.CreatedBy,
		LastMessageAt: lastMessageAt,
	}
}

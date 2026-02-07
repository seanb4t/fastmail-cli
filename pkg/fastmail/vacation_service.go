package fastmail

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// Vacation represents the vacation/auto-reply status.
type Vacation struct {
	IsEnabled bool   `json:"isEnabled"`
	FromDate  string `json:"fromDate,omitempty"`
	ToDate    string `json:"toDate,omitempty"`
	Subject   string `json:"subject,omitempty"`
	TextBody  string `json:"textBody,omitempty"`
	HTMLBody  string `json:"htmlBody,omitempty"`
}

// SetVacationOptions are the options for setting vacation response.
type SetVacationOptions struct {
	IsEnabled *bool
	FromDate  string
	ToDate    string
	Subject   string
	TextBody  string
	HTMLBody  string
}

// VacationService provides vacation response operations.
type VacationService struct {
	client *Client
}

// Get returns the current vacation response settings.
func (s *VacationService) Get(ctx context.Context) (*Vacation, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	getBuilder := jmap.NewVacationGet(accountID)
	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail, jmap.CapVacationResponse)
	callID := req.Invoke("VacationResponse/get", getBuilder.Build())

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

	var getResp jmap.VacationGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if len(getResp.List) == 0 {
		return &Vacation{}, nil
	}

	v := getResp.List[0]
	return &Vacation{
		IsEnabled: v.IsEnabled,
		FromDate:  v.FromDate,
		ToDate:    v.ToDate,
		Subject:   v.Subject,
		TextBody:  v.TextBody,
		HTMLBody:  v.HTMLBody,
	}, nil
}

// Set updates the vacation response settings.
func (s *VacationService) Set(ctx context.Context, opts SetVacationOptions) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	patch := make(map[string]any)
	if opts.IsEnabled != nil {
		patch["isEnabled"] = *opts.IsEnabled
	}
	if opts.FromDate != "" {
		patch["fromDate"] = opts.FromDate
	}
	if opts.ToDate != "" {
		patch["toDate"] = opts.ToDate
	}
	if opts.Subject != "" {
		patch["subject"] = opts.Subject
	}
	if opts.TextBody != "" {
		patch["textBody"] = opts.TextBody
	}
	if opts.HTMLBody != "" {
		patch["htmlBody"] = opts.HTMLBody
	}

	setBuilder := jmap.NewVacationSet(accountID).Update(patch)
	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail, jmap.CapVacationResponse)
	callID := req.Invoke("VacationResponse/set", setBuilder.Build())

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

	return nil
}

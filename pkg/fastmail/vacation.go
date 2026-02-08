package fastmail

import (
	"context"
	"time"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// VacationStatus represents the current vacation/out-of-office configuration.
type VacationStatus struct {
	IsEnabled bool
	FromDate  *time.Time
	ToDate    *time.Time
	Subject   string
	TextBody  string
}

// VacationService provides vacation response operations.
type VacationService struct {
	client *Client
}

// GetStatus returns the current vacation response status.
//
//nolint:dupl // JMAP service pattern - structural similarity with other services is intentional
func (s *VacationService) GetStatus(ctx context.Context) (*VacationStatus, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	getBuilder := jmap.NewVacationResponseGet(accountID)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
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

	var getResp jmap.VacationResponseGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if len(getResp.List) == 0 {
		return &VacationStatus{}, nil
	}

	return convertVacationResponse(&getResp.List[0])
}

// Enable enables the vacation response with the given subject and body.
// Optional from and to dates restrict when the auto-reply is active.
func (s *VacationService) Enable(ctx context.Context, subject, body string, from, to *time.Time) error {
	return s.update(ctx, func() map[string]any {
		patch := map[string]any{
			"isEnabled": true,
			"subject":   subject,
			"textBody":  body,
		}
		if from != nil {
			patch["fromDate"] = from.Format(time.RFC3339)
		}
		if to != nil {
			patch["toDate"] = to.Format(time.RFC3339)
		}
		return patch
	})
}

// Disable disables the vacation response.
func (s *VacationService) Disable(ctx context.Context) error {
	return s.update(ctx, func() map[string]any {
		return map[string]any{
			"isEnabled": false,
		}
	})
}

// update performs a VacationResponse/set with the given patch builder.
func (s *VacationService) update(ctx context.Context, patchFn func() map[string]any) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setBuilder := jmap.NewVacationResponseSet(accountID).
		Update("singleton", patchFn())

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
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

	var setResp jmap.VacationResponseSetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotUpdated["singleton"]; ok {
		return oops.Errorf("failed to update vacation response: %s - %s", errInfo.Type, errInfo.Description)
	}

	return nil
}

// convertVacationResponse converts a JMAP VacationResponse to a domain VacationStatus.
func convertVacationResponse(vr *jmap.VacationResponse) (*VacationStatus, error) {
	status := &VacationStatus{
		IsEnabled: vr.IsEnabled,
		Subject:   vr.Subject,
		TextBody:  vr.TextBody,
	}

	if vr.FromDate != "" {
		t, err := time.Parse(time.RFC3339, vr.FromDate)
		if err != nil {
			return nil, oops.Wrapf(err, "parsing fromDate")
		}
		status.FromDate = &t
	}

	if vr.ToDate != "" {
		t, err := time.Parse(time.RFC3339, vr.ToDate)
		if err != nil {
			return nil, oops.Wrapf(err, "parsing toDate")
		}
		status.ToDate = &t
	}

	return status, nil
}

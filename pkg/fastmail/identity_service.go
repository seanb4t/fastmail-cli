package fastmail

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// Identity represents a sender identity.
type Identity struct {
	ID            string
	Name          string
	Email         string
	ReplyTo       []EmailAddress
	BCC           []EmailAddress
	TextSignature string
	HTMLSignature string
	MayDelete     bool
}

// IdentityService provides identity operations.
type IdentityService struct {
	client *Client
}

// List returns all sender identities.
//
//nolint:dupl // JMAP service List methods follow a shared call/decode pattern by design.
func (s *IdentityService) List(ctx context.Context) ([]Identity, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "listing identities")
	}

	getBuilder := jmap.NewIdentityGet(accountID)
	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapSubmission)
	callID := req.Invoke("Identity/get", getBuilder.Build())

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

	var getResp jmap.IdentityGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	return convertIdentities(getResp.List), nil
}

func convertIdentities(jmapIdentities []jmap.Identity) []Identity {
	result := make([]Identity, 0, len(jmapIdentities))
	for _, id := range jmapIdentities {
		result = append(result, Identity{
			ID:            id.ID,
			Name:          id.Name,
			Email:         id.Email,
			ReplyTo:       convertJMAPToEmailAddresses(id.ReplyTo),
			BCC:           convertJMAPToEmailAddresses(id.BCC),
			TextSignature: id.TextSignature,
			HTMLSignature: id.HTMLSignature,
			MayDelete:     id.MayDelete,
		})
	}
	return result
}

func convertJMAPToEmailAddresses(addrs []jmap.EmailAddress) []EmailAddress {
	if len(addrs) == 0 {
		return nil
	}
	result := make([]EmailAddress, len(addrs))
	for i, a := range addrs {
		result[i] = EmailAddress{Name: a.Name, Email: a.Email}
	}
	return result
}

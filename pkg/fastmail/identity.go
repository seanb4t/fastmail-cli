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
	Bcc           []EmailAddress
	TextSignature string
	HTMLSignature string
	MayDelete     bool
}

// UpdateIdentityOptions specifies which fields to update on an identity.
// Only non-nil pointer fields and non-nil slices are included in the update.
type UpdateIdentityOptions struct {
	Name          *string
	ReplyTo       []EmailAddress
	TextSignature *string
	HTMLSignature *string
}

// IdentityService provides identity management operations.
type IdentityService struct {
	client *Client
}

// List returns all identities for the account.
//
//nolint:dupl // JMAP service pattern - structural similarity with other services is intentional
func (s *IdentityService) List(ctx context.Context) ([]Identity, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
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

// Update modifies an identity with the specified options.
func (s *IdentityService) Update(ctx context.Context, id string, opts UpdateIdentityOptions) error {
	patch := buildIdentityPatch(opts)
	if len(patch) == 0 {
		return oops.Errorf("no fields specified to update")
	}

	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setBuilder := jmap.NewIdentitySet(accountID).Update(id, patch)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapSubmission)
	callID := req.Invoke("Identity/set", setBuilder.Build())

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

	var setResp jmap.IdentitySetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotUpdated[id]; ok {
		return oops.Errorf("failed to update identity: %s - %s", errInfo.Type, errInfo.Description)
	}

	return nil
}

// buildIdentityPatch creates a JMAP patch map from UpdateIdentityOptions.
func buildIdentityPatch(opts UpdateIdentityOptions) map[string]any {
	patch := make(map[string]any)

	if opts.Name != nil {
		patch["name"] = *opts.Name
	}
	if opts.ReplyTo != nil {
		addrs := make([]map[string]string, len(opts.ReplyTo))
		for i, a := range opts.ReplyTo {
			addrs[i] = map[string]string{"email": a.Email}
			if a.Name != "" {
				addrs[i]["name"] = a.Name
			}
		}
		patch["replyTo"] = addrs
	}
	if opts.TextSignature != nil {
		patch["textSignature"] = *opts.TextSignature
	}
	if opts.HTMLSignature != nil {
		patch["htmlSignature"] = *opts.HTMLSignature
	}

	return patch
}

// convertIdentities converts JMAP identities to domain identities.
func convertIdentities(jmapIdentities []jmap.Identity) []Identity {
	identities := make([]Identity, len(jmapIdentities))
	for i, ji := range jmapIdentities {
		identities[i] = Identity{
			ID:            ji.ID,
			Name:          ji.Name,
			Email:         ji.Email,
			ReplyTo:       convertJMAPEmailAddresses(ji.ReplyTo),
			Bcc:           convertJMAPEmailAddresses(ji.Bcc),
			TextSignature: ji.TextSignature,
			HTMLSignature: ji.HTMLSignature,
			MayDelete:     ji.MayDelete,
		}
	}
	return identities
}

// convertJMAPEmailAddresses converts JMAP email addresses to domain email addresses.
func convertJMAPEmailAddresses(addrs []jmap.EmailAddress) []EmailAddress {
	if len(addrs) == 0 {
		return nil
	}
	result := make([]EmailAddress, len(addrs))
	for i, a := range addrs {
		result[i] = EmailAddress{
			Name:  a.Name,
			Email: a.Email,
		}
	}
	return result
}

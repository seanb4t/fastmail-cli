package fastmail

import (
	"context"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// SieveScript represents a Fastmail sieve filter script.
type SieveScript struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BlobID    string `json:"blobId,omitempty"`
	Script    string `json:"script,omitempty"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// SieveService provides sieve script operations.
type SieveService struct {
	client *Client
}

// CreateSieveScriptOptions specifies options for creating a sieve script.
type CreateSieveScriptOptions struct {
	Name     string
	Script   string
	Activate bool
}

// SieveValidationResult represents the result of a script validation.
type SieveValidationResult struct {
	IsValid     bool
	ErrorType   string
	Description string
}

// List returns all sieve scripts.
//
//nolint:dupl // JMAP service pattern - structural similarity with other services is intentional
func (s *SieveService) List(ctx context.Context) ([]SieveScript, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	getBuilder := jmap.NewSieveScriptGet(accountID)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.SieveScriptCapability)
	callID := req.Invoke("SieveScript/get", getBuilder.Build())

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

	var getResp jmap.SieveScriptGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	return convertSieveScripts(getResp.List), nil
}

// Get returns a single sieve script by ID.
func (s *SieveService) Get(ctx context.Context, id string) (*SieveScript, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	getBuilder := jmap.NewSieveScriptGet(accountID).IDs(id)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.SieveScriptCapability)
	callID := req.Invoke("SieveScript/get", getBuilder.Build())

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

	var getResp jmap.SieveScriptGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if len(getResp.List) == 0 {
		return nil, oops.Errorf("sieve script %s not found", id)
	}

	script := convertSieveScript(&getResp.List[0])
	return script, nil
}

// Create creates a new sieve script.
func (s *SieveService) Create(ctx context.Context, opts CreateSieveScriptOptions) (*SieveScript, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	setBuilder := jmap.NewSieveScriptSet(accountID)
	if opts.Activate {
		setBuilder.CreateActive("new", opts.Name, opts.Script)
	} else {
		setBuilder.Create("new", opts.Name, opts.Script)
	}

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.SieveScriptCapability)
	callID := req.Invoke("SieveScript/set", setBuilder.Build())

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

	var setResp jmap.SieveScriptSetResponse
	if err := result.Decode(&setResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotCreated["new"]; ok {
		return nil, oops.Errorf("failed to create sieve script: %s - %s", errInfo.Type, errInfo.Description)
	}

	created, ok := setResp.Created["new"]
	if !ok || created == nil {
		return nil, oops.Errorf("sieve script not returned in created map")
	}

	return convertSieveScript(created), nil
}

// Activate activates a sieve script.
func (s *SieveService) Activate(ctx context.Context, id string) error {
	return s.updateActive(ctx, id, true)
}

// Deactivate deactivates a sieve script.
func (s *SieveService) Deactivate(ctx context.Context, id string) error {
	return s.updateActive(ctx, id, false)
}

// updateActive updates the isActive state of a sieve script.
func (s *SieveService) updateActive(ctx context.Context, id string, active bool) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setBuilder := jmap.NewSieveScriptSet(accountID)
	if active {
		setBuilder.Activate(id)
	} else {
		setBuilder.Deactivate(id)
	}

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.SieveScriptCapability)
	callID := req.Invoke("SieveScript/set", setBuilder.Build())

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

	var setResp jmap.SieveScriptSetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotUpdated[id]; ok {
		return oops.Errorf("failed to update sieve script: %s - %s", errInfo.Type, errInfo.Description)
	}

	return nil
}

// Delete permanently deletes a sieve script.
//
//nolint:dupl // JMAP service pattern - structural similarity with other services is intentional
func (s *SieveService) Delete(ctx context.Context, id string) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	setBuilder := jmap.NewSieveScriptSet(accountID).Destroy(id)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.SieveScriptCapability)
	callID := req.Invoke("SieveScript/set", setBuilder.Build())

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

	var setResp jmap.SieveScriptSetResponse
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotDestroyed[id]; ok {
		return oops.Errorf("failed to delete sieve script: %s - %s", errInfo.Type, errInfo.Description)
	}

	return nil
}

// Validate validates a sieve script without storing it.
func (s *SieveService) Validate(ctx context.Context, script string) (*SieveValidationResult, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	validateBuilder := jmap.NewSieveScriptValidate(accountID, script)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.SieveScriptCapability)
	callID := req.Invoke("SieveScript/validate", validateBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("validate failed: %s", result.Error())
	}

	var validateResp jmap.SieveScriptValidateResponse
	if err := result.Decode(&validateResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if validateResp.Error != nil {
		return &SieveValidationResult{
			IsValid:     false,
			ErrorType:   validateResp.Error.Type,
			Description: validateResp.Error.Description,
		}, nil
	}

	return &SieveValidationResult{IsValid: true}, nil
}

// convertSieveScripts converts JMAP sieve scripts to domain sieve scripts.
func convertSieveScripts(jmapScripts []jmap.SieveScript) []SieveScript {
	result := make([]SieveScript, len(jmapScripts))
	for i, s := range jmapScripts {
		result[i] = SieveScript{
			ID:        s.ID,
			Name:      s.Name,
			BlobID:    s.BlobID,
			Script:    s.Script,
			IsActive:  s.IsActive,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		}
	}
	return result
}

// convertSieveScript converts a single JMAP sieve script to domain type.
func convertSieveScript(s *jmap.SieveScript) *SieveScript {
	if s == nil {
		return nil
	}
	return &SieveScript{
		ID:        s.ID,
		Name:      s.Name,
		BlobID:    s.BlobID,
		Script:    s.Script,
		IsActive:  s.IsActive,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

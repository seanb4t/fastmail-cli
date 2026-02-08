package jmap

// VacationResponse represents a JMAP VacationResponse object.
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-8
type VacationResponse struct {
	ID        string `json:"id"`
	IsEnabled bool   `json:"isEnabled"`
	FromDate  string `json:"fromDate,omitempty"`
	ToDate    string `json:"toDate,omitempty"`
	Subject   string `json:"subject,omitempty"`
	TextBody  string `json:"textBody,omitempty"`
	HTMLBody  string `json:"htmlBody,omitempty"`
}

// VacationResponseGetBuilder builds arguments for VacationResponse/get.
type VacationResponseGetBuilder struct {
	accountID string
}

// NewVacationResponseGet creates a new VacationResponse/get builder.
func NewVacationResponseGet(accountID string) *VacationResponseGetBuilder {
	return &VacationResponseGetBuilder{
		accountID: accountID,
	}
}

// Build returns the arguments map for Request.Invoke.
func (b *VacationResponseGetBuilder) Build() map[string]any {
	return map[string]any{
		"accountId": b.accountID,
		"ids":       []string{"singleton"},
	}
}

// VacationResponseGetResponse represents the response from VacationResponse/get.
type VacationResponseGetResponse struct {
	AccountID string             `json:"accountId"`
	State     string             `json:"state"`
	List      []VacationResponse `json:"list"`
	NotFound  []string           `json:"notFound"`
}

// VacationResponseSetBuilder builds arguments for VacationResponse/set.
type VacationResponseSetBuilder struct {
	accountID string
	update    map[string]map[string]any
}

// NewVacationResponseSet creates a new VacationResponse/set builder.
func NewVacationResponseSet(accountID string) *VacationResponseSetBuilder {
	return &VacationResponseSetBuilder{
		accountID: accountID,
		update:    make(map[string]map[string]any),
	}
}

// Update adds a VacationResponse to update.
func (b *VacationResponseSetBuilder) Update(id string, patch map[string]any) *VacationResponseSetBuilder {
	b.update[id] = patch
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *VacationResponseSetBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}

	if len(b.update) > 0 {
		args["update"] = b.update
	}

	return args
}

// VacationResponseSetResponse represents the response from VacationResponse/set.
type VacationResponseSetResponse struct {
	AccountID  string              `json:"accountId"`
	OldState   string              `json:"oldState"`
	NewState   string              `json:"newState"`
	Updated    map[string]any      `json:"updated"`
	NotUpdated map[string]SetError `json:"notUpdated"`
}

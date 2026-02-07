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

// VacationGetBuilder builds arguments for VacationResponse/get.
type VacationGetBuilder struct {
	accountID string
}

// NewVacationGet creates a new VacationResponse/get builder.
func NewVacationGet(accountID string) *VacationGetBuilder {
	return &VacationGetBuilder{accountID: accountID}
}

// Build returns the arguments map for Request.Invoke.
func (b *VacationGetBuilder) Build() map[string]any {
	return map[string]any{
		"accountId": b.accountID,
	}
}

// VacationGetResponse represents the response from VacationResponse/get.
type VacationGetResponse struct {
	AccountID string             `json:"accountId"`
	State     string             `json:"state"`
	List      []VacationResponse `json:"list"`
	NotFound  []string           `json:"notFound"`
}

// VacationSetBuilder builds arguments for VacationResponse/set.
type VacationSetBuilder struct {
	accountID string
	update    map[string]map[string]any
}

// NewVacationSet creates a new VacationResponse/set builder.
func NewVacationSet(accountID string) *VacationSetBuilder {
	return &VacationSetBuilder{
		accountID: accountID,
		update:    make(map[string]map[string]any),
	}
}

// Update adds a vacation response to be updated.
// VacationResponse is a singleton — the ID is always "singleton".
func (b *VacationSetBuilder) Update(patch map[string]any) *VacationSetBuilder {
	b.update["singleton"] = patch
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *VacationSetBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}
	if len(b.update) > 0 {
		args["update"] = b.update
	}
	return args
}

// VacationSetResponse represents the response from VacationResponse/set.
type VacationSetResponse struct {
	AccountID  string                 `json:"accountId"`
	OldState   string                 `json:"oldState"`
	NewState   string                 `json:"newState"`
	Updated    map[string]any         `json:"updated"`
	NotUpdated map[string]MethodError `json:"notUpdated"`
}

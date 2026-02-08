package jmap

// CapQuota is the JMAP capability URI for quota (RFC 9245).
const CapQuota = "urn:ietf:params:jmap:quota"

// Quota represents a JMAP Quota object.
// See: https://datatracker.ietf.org/doc/html/rfc9245
type Quota struct {
	ID           string   `json:"id"`
	ResourceType string   `json:"resourceType"`
	Used         uint64   `json:"used"`
	HardLimit    uint64   `json:"hardLimit"`
	Scope        string   `json:"scope"`
	Name         string   `json:"name"`
	Types        []string `json:"types,omitempty"`
	WarnLimit    uint64   `json:"warnLimit,omitempty"`
	SoftLimit    uint64   `json:"softLimit,omitempty"`
	Description  string   `json:"description,omitempty"`
}

// QuotaGetBuilder builds arguments for Quota/get.
type QuotaGetBuilder struct {
	accountID string
	ids       []string
}

// NewQuotaGet creates a new Quota/get builder.
func NewQuotaGet(accountID string) *QuotaGetBuilder {
	return &QuotaGetBuilder{
		accountID: accountID,
	}
}

// IDs sets the quota IDs to fetch.
func (b *QuotaGetBuilder) IDs(ids ...string) *QuotaGetBuilder {
	b.ids = ids
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *QuotaGetBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}

	if len(b.ids) > 0 {
		args["ids"] = b.ids
	}

	return args
}

// QuotaGetResponse represents the response from Quota/get.
type QuotaGetResponse struct {
	AccountID string   `json:"accountId"`
	State     string   `json:"state"`
	List      []Quota  `json:"list"`
	NotFound  []string `json:"notFound"`
}

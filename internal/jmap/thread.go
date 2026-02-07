package jmap

// Thread represents a JMAP Thread object.
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-3
type Thread struct {
	ID       string   `json:"id"`
	EmailIDs []string `json:"emailIds"`
}

// ThreadGetBuilder builds arguments for Thread/get.
type ThreadGetBuilder struct {
	accountID string
	ids       []string
}

// NewThreadGet creates a new Thread/get builder.
func NewThreadGet(accountID string) *ThreadGetBuilder {
	return &ThreadGetBuilder{accountID: accountID}
}

// IDs sets the thread IDs to fetch.
func (b *ThreadGetBuilder) IDs(ids ...string) *ThreadGetBuilder {
	b.ids = ids
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *ThreadGetBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}
	if len(b.ids) > 0 {
		args["ids"] = b.ids
	}
	return args
}

// ThreadGetResponse represents the response from Thread/get.
type ThreadGetResponse struct {
	AccountID string   `json:"accountId"`
	State     string   `json:"state"`
	List      []Thread `json:"list"`
	NotFound  []string `json:"notFound"`
}

package jmap

// SearchSnippet represents a JMAP SearchSnippet object.
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-5
type SearchSnippet struct {
	EmailID string  `json:"emailId"`
	Subject *string `json:"subject"` // nullable — highlighted subject with <mark> tags, or null
	Preview *string `json:"preview"` // nullable — highlighted preview with <mark> tags, or null
}

// SearchSnippetGetBuilder builds arguments for SearchSnippet/get.
type SearchSnippetGetBuilder struct {
	accountID string
	filter    map[string]any
	emailIDs  []string
}

// NewSearchSnippetGet creates a new SearchSnippet/get builder.
func NewSearchSnippetGet(accountID string) *SearchSnippetGetBuilder {
	return &SearchSnippetGetBuilder{
		accountID: accountID,
	}
}

// Filter sets the filter for SearchSnippet/get.
// This must match the filter used in the corresponding Email/query.
func (b *SearchSnippetGetBuilder) Filter(filter map[string]any) *SearchSnippetGetBuilder {
	b.filter = filter
	return b
}

// FilterFromQuery copies the filter from an EmailQueryBuilder.
// This ensures the SearchSnippet/get filter matches the Email/query filter exactly.
func (b *SearchSnippetGetBuilder) FilterFromQuery(query *EmailQueryBuilder) *SearchSnippetGetBuilder {
	if len(query.filter) > 0 {
		// Copy the filter map to avoid aliasing
		b.filter = make(map[string]any, len(query.filter))
		for k, v := range query.filter {
			b.filter[k] = v
		}
	}
	return b
}

// EmailIDs sets the email IDs to get snippets for.
func (b *SearchSnippetGetBuilder) EmailIDs(ids ...string) *SearchSnippetGetBuilder {
	b.emailIDs = ids
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *SearchSnippetGetBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}

	if len(b.filter) > 0 {
		args["filter"] = b.filter
	}
	if len(b.emailIDs) > 0 {
		args["emailIds"] = b.emailIDs
	}

	return args
}

// SearchSnippetGetResponse represents the response from SearchSnippet/get.
type SearchSnippetGetResponse struct {
	AccountID string          `json:"accountId"`
	List      []SearchSnippet `json:"list"`
	NotFound  []string        `json:"notFound"`
}

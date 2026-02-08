package jmap

// Email represents a JMAP Email object.
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-4
type Email struct {
	ID         string          `json:"id"`
	BlobID     string          `json:"blobId,omitempty"`
	ThreadID   string          `json:"threadId,omitempty"`
	MailboxIDs map[string]bool `json:"mailboxIds,omitempty"`
	Keywords   map[string]bool `json:"keywords,omitempty"`
	Size       uint64          `json:"size,omitempty"`
	ReceivedAt string          `json:"receivedAt,omitempty"`
	Subject    string          `json:"subject,omitempty"`
	Preview    string          `json:"preview,omitempty"`

	// Header fields for reading/creating emails
	From       []EmailAddress `json:"from,omitempty"`
	To         []EmailAddress `json:"to,omitempty"`
	Cc         []EmailAddress `json:"cc,omitempty"`
	Bcc        []EmailAddress `json:"bcc,omitempty"`
	ReplyTo    []EmailAddress `json:"replyTo,omitempty"`
	MessageID  []string       `json:"messageId,omitempty"`
	InReplyTo  []string       `json:"inReplyTo,omitempty"`
	References []string       `json:"references,omitempty"`

	// Body content
	BodyValues  map[string]BodyValue `json:"bodyValues,omitempty"`
	TextBody    []BodyPart           `json:"textBody,omitempty"`
	HTMLBody    []BodyPart           `json:"htmlBody,omitempty"`
	Attachments []Attachment         `json:"attachments,omitempty"`
}

// EmailAddress represents a JMAP email address.
type EmailAddress struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

// BodyValue contains the actual content of an email body part.
type BodyValue struct {
	Value             string `json:"value"`
	IsEncodingProblem bool   `json:"isEncodingProblem,omitempty"`
	IsTruncated       bool   `json:"isTruncated,omitempty"`
}

// BodyPart describes a part of the email body structure.
type BodyPart struct {
	PartID string `json:"partId,omitempty"`
	Type   string `json:"type,omitempty"`
}

// Attachment represents a JMAP email attachment.
type Attachment struct {
	BlobID string `json:"blobId"`
	Type   string `json:"type"`
	Name   string `json:"name,omitempty"`
	Size   uint64 `json:"size"`
}

// Comparator specifies sort order for queries.
// See: https://datatracker.ietf.org/doc/html/rfc8620#section-5.1
type Comparator struct {
	Property    string `json:"property"`
	IsAscending bool   `json:"isAscending"`
}

// EmailQueryBuilder builds arguments for Email/query.
type EmailQueryBuilder struct {
	accountID string
	filter    map[string]any
	sort      []Comparator
	limit     uint64
	position  uint64
}

// NewEmailQuery creates a new Email/query builder.
func NewEmailQuery(accountID string) *EmailQueryBuilder {
	return &EmailQueryBuilder{
		accountID: accountID,
		filter:    make(map[string]any),
	}
}

// InMailbox filters emails in a specific mailbox.
func (b *EmailQueryBuilder) InMailbox(mailboxID string) *EmailQueryBuilder {
	b.filter["inMailbox"] = mailboxID
	return b
}

// From filters emails by sender.
func (b *EmailQueryBuilder) From(email string) *EmailQueryBuilder {
	b.filter["from"] = email
	return b
}

// HasKeyword filters emails with a specific keyword.
func (b *EmailQueryBuilder) HasKeyword(keyword string) *EmailQueryBuilder {
	b.filter["hasKeyword"] = keyword
	return b
}

// Limit sets the maximum number of results.
func (b *EmailQueryBuilder) Limit(n uint64) *EmailQueryBuilder {
	b.limit = n
	return b
}

// SortBy adds a sort comparator. Set descending=true for newest first.
func (b *EmailQueryBuilder) SortBy(property string, descending bool) *EmailQueryBuilder {
	b.sort = append(b.sort, Comparator{
		Property:    property,
		IsAscending: !descending,
	})
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *EmailQueryBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}

	if len(b.filter) > 0 {
		args["filter"] = b.filter
	}
	if len(b.sort) > 0 {
		args["sort"] = b.sort
	}
	if b.limit > 0 {
		args["limit"] = b.limit
	}
	if b.position > 0 {
		args["position"] = b.position
	}

	return args
}

// EmailQueryResponse represents the response from Email/query.
type EmailQueryResponse struct {
	AccountID           string   `json:"accountId"`
	QueryState          string   `json:"queryState"`
	CanCalculateChanges bool     `json:"canCalculateChanges"`
	Position            uint64   `json:"position"`
	IDs                 []string `json:"ids"`
	Total               uint64   `json:"total"`
}

// EmailGetBuilder builds arguments for Email/get.
type EmailGetBuilder struct {
	accountID  string
	ids        []string
	properties []string
}

// NewEmailGet creates a new Email/get builder.
func NewEmailGet(accountID string) *EmailGetBuilder {
	return &EmailGetBuilder{
		accountID: accountID,
	}
}

// IDs sets the email IDs to fetch.
func (b *EmailGetBuilder) IDs(ids ...string) *EmailGetBuilder {
	b.ids = ids
	return b
}

// Properties sets which properties to fetch (optimization).
func (b *EmailGetBuilder) Properties(props ...string) *EmailGetBuilder {
	b.properties = props
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *EmailGetBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}

	if len(b.ids) > 0 {
		args["ids"] = b.ids
	}
	if len(b.properties) > 0 {
		args["properties"] = b.properties
	}

	return args
}

// EmailGetResponse represents the response from Email/get.
type EmailGetResponse struct {
	AccountID string   `json:"accountId"`
	State     string   `json:"state"`
	List      []Email  `json:"list"`
	NotFound  []string `json:"notFound"`
}

// EmailChangesBuilder builds arguments for Email/changes.
type EmailChangesBuilder struct {
	accountID  string
	sinceState string
	maxChanges uint64
}

// NewEmailChanges creates a new Email/changes builder.
func NewEmailChanges(accountID, sinceState string) *EmailChangesBuilder {
	return &EmailChangesBuilder{
		accountID:  accountID,
		sinceState: sinceState,
	}
}

// MaxChanges sets the maximum number of changes to return.
func (b *EmailChangesBuilder) MaxChanges(n uint64) *EmailChangesBuilder {
	b.maxChanges = n
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *EmailChangesBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId":  b.accountID,
		"sinceState": b.sinceState,
	}

	if b.maxChanges > 0 {
		args["maxChanges"] = b.maxChanges
	}

	return args
}

// EmailChangesResponse represents the response from Email/changes.
type EmailChangesResponse struct {
	AccountID      string   `json:"accountId"`
	OldState       string   `json:"oldState"`
	NewState       string   `json:"newState"`
	HasMoreChanges bool     `json:"hasMoreChanges"`
	Created        []string `json:"created"`
	Updated        []string `json:"updated"`
	Destroyed      []string `json:"destroyed"`
}

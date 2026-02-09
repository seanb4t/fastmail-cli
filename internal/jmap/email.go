package jmap

import "time"

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
	Attachments []BodyPart           `json:"attachments,omitempty"`
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
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-4.1.4
type BodyPart struct {
	PartID      string `json:"partId,omitempty"`
	BlobID      string `json:"blobId,omitempty"`
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Disposition string `json:"disposition,omitempty"`
	Size        uint64 `json:"size,omitempty"`
	Cid         string `json:"cid,omitempty"`
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

// To filters emails by recipient.
func (b *EmailQueryBuilder) To(email string) *EmailQueryBuilder {
	b.filter["to"] = email
	return b
}

// Cc filters emails by CC recipient.
func (b *EmailQueryBuilder) Cc(email string) *EmailQueryBuilder {
	b.filter["cc"] = email
	return b
}

// Bcc filters emails by BCC recipient.
func (b *EmailQueryBuilder) Bcc(email string) *EmailQueryBuilder {
	b.filter["bcc"] = email
	return b
}

// Subject filters emails by subject text.
func (b *EmailQueryBuilder) Subject(text string) *EmailQueryBuilder {
	b.filter["subject"] = text
	return b
}

// Body filters emails by body content.
func (b *EmailQueryBuilder) Body(text string) *EmailQueryBuilder {
	b.filter["body"] = text
	return b
}

// Text filters emails by full-text search across all fields.
func (b *EmailQueryBuilder) Text(text string) *EmailQueryBuilder {
	b.filter["text"] = text
	return b
}

// Before filters emails received before the given time.
func (b *EmailQueryBuilder) Before(t time.Time) *EmailQueryBuilder {
	b.filter["before"] = t.UTC().Format(time.RFC3339)
	return b
}

// After filters emails received after the given time.
func (b *EmailQueryBuilder) After(t time.Time) *EmailQueryBuilder {
	b.filter["after"] = t.UTC().Format(time.RFC3339)
	return b
}

// MinSize filters emails with size >= minSize bytes.
func (b *EmailQueryBuilder) MinSize(minSize uint64) *EmailQueryBuilder {
	b.filter["minSize"] = minSize
	return b
}

// MaxSize filters emails with size <= maxSize bytes.
func (b *EmailQueryBuilder) MaxSize(maxSize uint64) *EmailQueryBuilder {
	b.filter["maxSize"] = maxSize
	return b
}

// HasAttachment filters emails that have (or don't have) attachments.
func (b *EmailQueryBuilder) HasAttachment(has bool) *EmailQueryBuilder {
	b.filter["hasAttachment"] = has
	return b
}

// HasKeyword filters emails with a specific keyword.
func (b *EmailQueryBuilder) HasKeyword(keyword string) *EmailQueryBuilder {
	b.filter["hasKeyword"] = keyword
	return b
}

// NotKeyword filters emails that do NOT have a specific keyword.
func (b *EmailQueryBuilder) NotKeyword(keyword string) *EmailQueryBuilder {
	b.filter["notKeyword"] = keyword
	return b
}

// InMailboxOtherThan filters emails NOT in the specified mailboxes.
func (b *EmailQueryBuilder) InMailboxOtherThan(mailboxIDs ...string) *EmailQueryBuilder {
	b.filter["inMailboxOtherThan"] = mailboxIDs
	return b
}

// Header filters emails by a specific header field name and value.
func (b *EmailQueryBuilder) Header(name, value string) *EmailQueryBuilder {
	b.filter["header"] = []string{name, value}
	return b
}

// WithFilter replaces the entire filter with a pre-built filter map.
// This is used for compound filters (AND/OR/NOT) from the search parser.
func (b *EmailQueryBuilder) WithFilter(filter map[string]any) *EmailQueryBuilder {
	b.filter = filter
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

// ThreadGetBuilder builds arguments for Thread/get.
type ThreadGetBuilder struct {
	accountID string
	ids       []string
}

// NewThreadGet creates a new Thread/get builder.
func NewThreadGet(accountID string) *ThreadGetBuilder {
	return &ThreadGetBuilder{
		accountID: accountID,
	}
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

// Thread represents a JMAP Thread object.
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-3
type Thread struct {
	ID       string   `json:"id"`
	EmailIDs []string `json:"emailIds"`
}

// ThreadGetResponse represents the response from Thread/get.
type ThreadGetResponse struct {
	AccountID string   `json:"accountId"`
	State     string   `json:"state"`
	List      []Thread `json:"list"`
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

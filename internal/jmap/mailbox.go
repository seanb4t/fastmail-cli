package jmap

// Mailbox represents a JMAP Mailbox object.
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-2
type Mailbox struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ParentID      string `json:"parentId,omitempty"`
	Role          string `json:"role,omitempty"`
	SortOrder     uint32 `json:"sortOrder"`
	TotalEmails   uint64 `json:"totalEmails"`
	UnreadEmails  uint64 `json:"unreadEmails"`
	TotalThreads  uint64 `json:"totalThreads"`
	UnreadThreads uint64 `json:"unreadThreads"`
}

// MailboxGetBuilder builds arguments for Mailbox/get.
type MailboxGetBuilder struct {
	accountID  string
	ids        []string
	properties []string
}

// NewMailboxGet creates a new Mailbox/get builder.
func NewMailboxGet(accountID string) *MailboxGetBuilder {
	return &MailboxGetBuilder{
		accountID: accountID,
	}
}

// IDs sets specific mailbox IDs to fetch. If not called, fetches all mailboxes.
func (b *MailboxGetBuilder) IDs(ids ...string) *MailboxGetBuilder {
	b.ids = ids
	return b
}

// Properties sets which properties to fetch (optimization).
func (b *MailboxGetBuilder) Properties(props ...string) *MailboxGetBuilder {
	b.properties = props
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *MailboxGetBuilder) Build() map[string]any {
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

// MailboxGetResponse represents the response from Mailbox/get.
type MailboxGetResponse struct {
	AccountID string    `json:"accountId"`
	State     string    `json:"state"`
	List      []Mailbox `json:"list"`
	NotFound  []string  `json:"notFound"`
}

// MailboxSetBuilder builds arguments for Mailbox/set.
type MailboxSetBuilder struct {
	accountID string
	create    map[string]map[string]any
	update    map[string]map[string]any
	destroy   []string
}

// NewMailboxSet creates a new Mailbox/set builder.
func NewMailboxSet(accountID string) *MailboxSetBuilder {
	return &MailboxSetBuilder{
		accountID: accountID,
		create:    make(map[string]map[string]any),
		update:    make(map[string]map[string]any),
	}
}

// Create adds a mailbox to be created.
func (b *MailboxSetBuilder) Create(clientID string, mailbox map[string]any) *MailboxSetBuilder {
	b.create[clientID] = mailbox
	return b
}

// Update adds a mailbox to be updated.
func (b *MailboxSetBuilder) Update(mailboxID string, patch map[string]any) *MailboxSetBuilder {
	b.update[mailboxID] = patch
	return b
}

// Destroy adds mailbox IDs to be destroyed.
func (b *MailboxSetBuilder) Destroy(ids ...string) *MailboxSetBuilder {
	b.destroy = append(b.destroy, ids...)
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *MailboxSetBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}
	if len(b.create) > 0 {
		args["create"] = b.create
	}
	if len(b.update) > 0 {
		args["update"] = b.update
	}
	if len(b.destroy) > 0 {
		args["destroy"] = b.destroy
	}
	return args
}

// MailboxSetResponse represents the response from Mailbox/set.
type MailboxSetResponse struct {
	AccountID    string                 `json:"accountId"`
	OldState     string                 `json:"oldState"`
	NewState     string                 `json:"newState"`
	Created      map[string]Mailbox     `json:"created"`
	Updated      map[string]any         `json:"updated"`
	Destroyed    []string               `json:"destroyed"`
	NotCreated   map[string]MethodError `json:"notCreated"`
	NotUpdated   map[string]MethodError `json:"notUpdated"`
	NotDestroyed map[string]MethodError `json:"notDestroyed"`
}

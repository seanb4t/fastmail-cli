package jmap

// CapMaskedEmail is the Fastmail capability URI for masked email.
const CapMaskedEmail = "https://www.fastmail.com/dev/maskedemail"

// MaskedEmailState represents the state of a masked email address.
type MaskedEmailState string

// Masked email states.
const (
	MaskedEmailStateEnabled  MaskedEmailState = "enabled"
	MaskedEmailStateDisabled MaskedEmailState = "disabled"
	MaskedEmailStateDeleted  MaskedEmailState = "deleted"
	MaskedEmailStatePending  MaskedEmailState = "pending"
)

// MaskedEmail represents a Fastmail masked email address.
// See: https://www.fastmail.com/dev/maskedemail
type MaskedEmail struct {
	ID            string           `json:"id,omitempty"`
	Email         string           `json:"email,omitempty"`
	State         MaskedEmailState `json:"state,omitempty"`
	ForDomain     string           `json:"forDomain,omitempty"`
	Description   string           `json:"description,omitempty"`
	URL           string           `json:"url,omitempty"`
	CreatedBy     string           `json:"createdBy,omitempty"`
	CreatedAt     string           `json:"createdAt,omitempty"`
	LastMessageAt string           `json:"lastMessageAt,omitempty"`
}

// MaskedEmailGetBuilder builds arguments for MaskedEmail/get.
type MaskedEmailGetBuilder struct {
	accountID  string
	ids        []string
	properties []string
}

// NewMaskedEmailGet creates a new MaskedEmail/get builder.
func NewMaskedEmailGet(accountID string) *MaskedEmailGetBuilder {
	return &MaskedEmailGetBuilder{
		accountID: accountID,
	}
}

// IDs sets the masked email IDs to fetch.
func (b *MaskedEmailGetBuilder) IDs(ids ...string) *MaskedEmailGetBuilder {
	b.ids = ids
	return b
}

// Properties sets which properties to fetch.
func (b *MaskedEmailGetBuilder) Properties(props ...string) *MaskedEmailGetBuilder {
	b.properties = props
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *MaskedEmailGetBuilder) Build() map[string]any {
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

// MaskedEmailGetResponse represents the response from MaskedEmail/get.
type MaskedEmailGetResponse struct {
	AccountID string        `json:"accountId"`
	State     string        `json:"state"`
	List      []MaskedEmail `json:"list"`
	NotFound  []string      `json:"notFound"`
}

// MaskedEmailSetBuilder builds arguments for MaskedEmail/set.
type MaskedEmailSetBuilder struct {
	accountID string
	create    map[string]map[string]any
	update    map[string]map[string]any
	destroy   []string
}

// NewMaskedEmailSet creates a new MaskedEmail/set builder.
func NewMaskedEmailSet(accountID string) *MaskedEmailSetBuilder {
	return &MaskedEmailSetBuilder{
		accountID: accountID,
		create:    make(map[string]map[string]any),
		update:    make(map[string]map[string]any),
	}
}

// Create adds a masked email to be created.
// The clientID is a temporary ID used to reference this masked email in result references.
func (b *MaskedEmailSetBuilder) Create(clientID string, maskedEmail map[string]any) *MaskedEmailSetBuilder {
	b.create[clientID] = maskedEmail
	return b
}

// Update adds a masked email to be updated.
func (b *MaskedEmailSetBuilder) Update(maskedEmailID string, patch map[string]any) *MaskedEmailSetBuilder {
	b.update[maskedEmailID] = patch
	return b
}

// Destroy adds masked email IDs to be destroyed.
func (b *MaskedEmailSetBuilder) Destroy(ids ...string) *MaskedEmailSetBuilder {
	b.destroy = append(b.destroy, ids...)
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *MaskedEmailSetBuilder) Build() map[string]any {
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

// MaskedEmailSetResponse represents the response from MaskedEmail/set.
type MaskedEmailSetResponse struct {
	AccountID    string                 `json:"accountId"`
	OldState     string                 `json:"oldState"`
	NewState     string                 `json:"newState"`
	Created      map[string]MaskedEmail `json:"created"`
	Updated      map[string]any         `json:"updated"`
	Destroyed    []string               `json:"destroyed"`
	NotCreated   map[string]MethodError `json:"notCreated"`
	NotUpdated   map[string]MethodError `json:"notUpdated"`
	NotDestroyed map[string]MethodError `json:"notDestroyed"`
}

// MaskedEmailChangesBuilder builds arguments for MaskedEmail/changes.
type MaskedEmailChangesBuilder struct {
	accountID  string
	sinceState string
	maxChanges uint64
}

// NewMaskedEmailChanges creates a new MaskedEmail/changes builder.
func NewMaskedEmailChanges(accountID, sinceState string) *MaskedEmailChangesBuilder {
	return &MaskedEmailChangesBuilder{
		accountID:  accountID,
		sinceState: sinceState,
	}
}

// MaxChanges sets the maximum number of changes to return.
func (b *MaskedEmailChangesBuilder) MaxChanges(n uint64) *MaskedEmailChangesBuilder {
	b.maxChanges = n
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *MaskedEmailChangesBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId":  b.accountID,
		"sinceState": b.sinceState,
	}

	if b.maxChanges > 0 {
		args["maxChanges"] = b.maxChanges
	}

	return args
}

// MaskedEmailChangesResponse represents the response from MaskedEmail/changes.
type MaskedEmailChangesResponse struct {
	AccountID      string   `json:"accountId"`
	OldState       string   `json:"oldState"`
	NewState       string   `json:"newState"`
	HasMoreChanges bool     `json:"hasMoreChanges"`
	Created        []string `json:"created"`
	Updated        []string `json:"updated"`
	Destroyed      []string `json:"destroyed"`
}

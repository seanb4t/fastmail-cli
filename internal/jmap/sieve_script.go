package jmap

// SieveScriptCapability is the JMAP capability URI for SieveScript (RFC 9661).
const SieveScriptCapability = "urn:ietf:params:jmap:sieve"

// SieveScript represents a JMAP SieveScript object.
type SieveScript struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	BlobID    string `json:"blobId,omitempty"`
	Script    string `json:"script,omitempty"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// SieveScriptGetBuilder builds arguments for SieveScript/get.
type SieveScriptGetBuilder struct {
	accountID  string
	ids        []string
	properties []string
}

// NewSieveScriptGet creates a new SieveScript/get builder.
func NewSieveScriptGet(accountID string) *SieveScriptGetBuilder {
	return &SieveScriptGetBuilder{
		accountID: accountID,
	}
}

// IDs sets the sieve script IDs to fetch.
func (b *SieveScriptGetBuilder) IDs(ids ...string) *SieveScriptGetBuilder {
	b.ids = ids
	return b
}

// Properties sets which properties to fetch.
func (b *SieveScriptGetBuilder) Properties(props ...string) *SieveScriptGetBuilder {
	b.properties = props
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *SieveScriptGetBuilder) Build() map[string]any {
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

// SieveScriptGetResponse represents the response from SieveScript/get.
type SieveScriptGetResponse struct {
	AccountID string        `json:"accountId"`
	State     string        `json:"state"`
	List      []SieveScript `json:"list"`
	NotFound  []string      `json:"notFound"`
}

// SieveScriptSetBuilder builds arguments for SieveScript/set.
type SieveScriptSetBuilder struct {
	accountID string
	ifInState string
	create    map[string]map[string]any
	update    map[string]map[string]any
	destroy   []string
}

// NewSieveScriptSet creates a new SieveScript/set builder.
func NewSieveScriptSet(accountID string) *SieveScriptSetBuilder {
	return &SieveScriptSetBuilder{
		accountID: accountID,
		create:    make(map[string]map[string]any),
		update:    make(map[string]map[string]any),
	}
}

// Create adds a sieve script to create.
func (b *SieveScriptSetBuilder) Create(clientID, name, script string) *SieveScriptSetBuilder {
	b.create[clientID] = map[string]any{
		"name":   name,
		"script": script,
	}
	return b
}

// CreateActive adds a sieve script to create with isActive set.
func (b *SieveScriptSetBuilder) CreateActive(clientID, name, script string) *SieveScriptSetBuilder {
	b.create[clientID] = map[string]any{
		"name":     name,
		"script":   script,
		"isActive": true,
	}
	return b
}

// Update adds a sieve script to update.
func (b *SieveScriptSetBuilder) Update(id string, updates map[string]any) *SieveScriptSetBuilder {
	b.update[id] = updates
	return b
}

// Activate updates a sieve script to be active.
func (b *SieveScriptSetBuilder) Activate(id string) *SieveScriptSetBuilder {
	return b.Update(id, map[string]any{"isActive": true})
}

// Deactivate updates a sieve script to be inactive.
func (b *SieveScriptSetBuilder) Deactivate(id string) *SieveScriptSetBuilder {
	return b.Update(id, map[string]any{"isActive": false})
}

// Destroy adds sieve script IDs to delete.
func (b *SieveScriptSetBuilder) Destroy(ids ...string) *SieveScriptSetBuilder {
	b.destroy = append(b.destroy, ids...)
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *SieveScriptSetBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}

	if b.ifInState != "" {
		args["ifInState"] = b.ifInState
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

// SieveScriptSetResponse represents the response from SieveScript/set.
type SieveScriptSetResponse struct {
	AccountID    string                  `json:"accountId"`
	OldState     string                  `json:"oldState"`
	NewState     string                  `json:"newState"`
	Created      map[string]*SieveScript `json:"created"`
	Updated      map[string]*SieveScript `json:"updated"`
	Destroyed    []string                `json:"destroyed"`
	NotCreated   map[string]SetError     `json:"notCreated"`
	NotUpdated   map[string]SetError     `json:"notUpdated"`
	NotDestroyed map[string]SetError     `json:"notDestroyed"`
}

// SieveScriptValidateBuilder builds arguments for SieveScript/validate.
type SieveScriptValidateBuilder struct {
	accountID string
	script    string
}

// NewSieveScriptValidate creates a new SieveScript/validate builder.
func NewSieveScriptValidate(accountID, script string) *SieveScriptValidateBuilder {
	return &SieveScriptValidateBuilder{
		accountID: accountID,
		script:    script,
	}
}

// Build returns the arguments map for Request.Invoke.
func (b *SieveScriptValidateBuilder) Build() map[string]any {
	return map[string]any{
		"accountId": b.accountID,
		"script":    b.script,
	}
}

// SieveScriptValidateResponse represents the response from SieveScript/validate.
type SieveScriptValidateResponse struct {
	AccountID string            `json:"accountId"`
	Error     *SieveScriptError `json:"error,omitempty"`
}

// SieveScriptError represents a validation error from SieveScript/validate.
type SieveScriptError struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

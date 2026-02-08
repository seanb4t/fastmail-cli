package jmap

// CapSubmission is the JMAP capability URI for email submission.
const CapSubmission = "urn:ietf:params:jmap:submission"

// EmailSubmission represents a JMAP EmailSubmission object.
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-7
type EmailSubmission struct {
	ID             string            `json:"id,omitempty"`
	IdentityID     string            `json:"identityId"`
	EmailID        string            `json:"emailId"`
	ThreadID       string            `json:"threadId,omitempty"`
	Envelope       *Envelope         `json:"envelope,omitempty"`
	SendAt         string            `json:"sendAt,omitempty"`
	UndoStatus     string            `json:"undoStatus,omitempty"`
	DeliveryStatus map[string]Status `json:"deliveryStatus,omitempty"`
	DSNBlobIDs     []string          `json:"dsnBlobIds,omitempty"`
	MDNBlobIDs     []string          `json:"mdnBlobIds,omitempty"`
}

// Envelope specifies the SMTP envelope for email submission.
type Envelope struct {
	MailFrom Address   `json:"mailFrom"`
	RcptTo   []Address `json:"rcptTo"`
}

// Address represents an email address for submission.
type Address struct {
	Email      string            `json:"email"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// Status represents delivery status for a recipient.
type Status struct {
	SMTPReply string `json:"smtpReply,omitempty"`
	Delivered string `json:"delivered"`
	Displayed string `json:"displayed"`
}

// EmailSetBuilder builds arguments for Email/set (create/update/destroy).
type EmailSetBuilder struct {
	accountID string
	create    map[string]map[string]any
	update    map[string]map[string]any
	destroy   []string
}

// NewEmailSet creates a new Email/set builder.
func NewEmailSet(accountID string) *EmailSetBuilder {
	return &EmailSetBuilder{
		accountID: accountID,
		create:    make(map[string]map[string]any),
		update:    make(map[string]map[string]any),
	}
}

// Create adds an email to be created.
// The clientID is a temporary ID used to reference this email in result references.
func (b *EmailSetBuilder) Create(clientID string, email map[string]any) *EmailSetBuilder {
	b.create[clientID] = email
	return b
}

// Update adds an email to be updated.
func (b *EmailSetBuilder) Update(emailID string, patch map[string]any) *EmailSetBuilder {
	b.update[emailID] = patch
	return b
}

// Destroy adds email IDs to be destroyed.
func (b *EmailSetBuilder) Destroy(ids ...string) *EmailSetBuilder {
	b.destroy = append(b.destroy, ids...)
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *EmailSetBuilder) Build() map[string]any {
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

// EmailSetResponse represents the response from Email/set.
type EmailSetResponse struct {
	AccountID    string                 `json:"accountId"`
	OldState     string                 `json:"oldState"`
	NewState     string                 `json:"newState"`
	Created      map[string]Email       `json:"created"`
	Updated      map[string]any         `json:"updated"`
	Destroyed    []string               `json:"destroyed"`
	NotCreated   map[string]MethodError `json:"notCreated"`
	NotUpdated   map[string]MethodError `json:"notUpdated"`
	NotDestroyed map[string]MethodError `json:"notDestroyed"`
}

// EmailSubmissionSetBuilder builds arguments for EmailSubmission/set.
type EmailSubmissionSetBuilder struct {
	accountID        string
	create           map[string]map[string]any
	onSuccessUpdate  map[string]map[string]any
	onSuccessDestroy []string
}

// NewEmailSubmissionSet creates a new EmailSubmission/set builder.
func NewEmailSubmissionSet(accountID string) *EmailSubmissionSetBuilder {
	return &EmailSubmissionSetBuilder{
		accountID:       accountID,
		create:          make(map[string]map[string]any),
		onSuccessUpdate: make(map[string]map[string]any),
	}
}

// Create adds an email submission to be created.
func (b *EmailSubmissionSetBuilder) Create(clientID string, submission map[string]any) *EmailSubmissionSetBuilder {
	b.create[clientID] = submission
	return b
}

// OnSuccessUpdateEmail sets properties to update on the email after successful submission.
// The emailRef should be a reference like "#emailClientId".
func (b *EmailSubmissionSetBuilder) OnSuccessUpdateEmail(emailRef string, patch map[string]any) *EmailSubmissionSetBuilder {
	b.onSuccessUpdate[emailRef] = patch
	return b
}

// OnSuccessDestroyEmail adds email IDs to destroy after successful submission.
func (b *EmailSubmissionSetBuilder) OnSuccessDestroyEmail(emailRefs ...string) *EmailSubmissionSetBuilder {
	b.onSuccessDestroy = append(b.onSuccessDestroy, emailRefs...)
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *EmailSubmissionSetBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}

	if len(b.create) > 0 {
		args["create"] = b.create
	}
	if len(b.onSuccessUpdate) > 0 {
		args["onSuccessUpdateEmail"] = b.onSuccessUpdate
	}
	if len(b.onSuccessDestroy) > 0 {
		args["onSuccessDestroyEmail"] = b.onSuccessDestroy
	}

	return args
}

// EmailSubmissionSetResponse represents the response from EmailSubmission/set.
type EmailSubmissionSetResponse struct {
	AccountID  string                     `json:"accountId"`
	OldState   string                     `json:"oldState"`
	NewState   string                     `json:"newState"`
	Created    map[string]EmailSubmission `json:"created"`
	NotCreated map[string]MethodError     `json:"notCreated"`
}

// Identity represents a JMAP Identity object (sender identity).
// See: https://datatracker.ietf.org/doc/html/rfc8620#section-6 (via submission).
type Identity struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	ReplyTo       []EmailAddress `json:"replyTo,omitempty"`
	Bcc           []EmailAddress `json:"bcc,omitempty"`
	TextSignature string         `json:"textSignature,omitempty"`
	HTMLSignature string         `json:"htmlSignature,omitempty"`
	MayDelete     bool           `json:"mayDelete"`
}

// IdentityGetBuilder builds arguments for Identity/get.
type IdentityGetBuilder struct {
	accountID  string
	ids        []string
	properties []string
}

// NewIdentityGet creates a new Identity/get builder.
func NewIdentityGet(accountID string) *IdentityGetBuilder {
	return &IdentityGetBuilder{
		accountID: accountID,
	}
}

// IDs sets the identity IDs to fetch.
func (b *IdentityGetBuilder) IDs(ids ...string) *IdentityGetBuilder {
	b.ids = ids
	return b
}

// Properties sets which properties to fetch.
func (b *IdentityGetBuilder) Properties(props ...string) *IdentityGetBuilder {
	b.properties = props
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *IdentityGetBuilder) Build() map[string]any {
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

// IdentityGetResponse represents the response from Identity/get.
type IdentityGetResponse struct {
	AccountID string     `json:"accountId"`
	State     string     `json:"state"`
	List      []Identity `json:"list"`
	NotFound  []string   `json:"notFound"`
}

// IdentitySetBuilder builds arguments for Identity/set.
// Per JMAP spec, Identity only supports get and set (not create/destroy);
// the server manages identity creation and deletion.
type IdentitySetBuilder struct {
	accountID string
	update    map[string]map[string]any
}

// NewIdentitySet creates a new Identity/set builder.
func NewIdentitySet(accountID string) *IdentitySetBuilder {
	return &IdentitySetBuilder{
		accountID: accountID,
		update:    make(map[string]map[string]any),
	}
}

// Update adds an identity to be updated with the given patch.
func (b *IdentitySetBuilder) Update(identityID string, patch map[string]any) *IdentitySetBuilder {
	b.update[identityID] = patch
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *IdentitySetBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}

	if len(b.update) > 0 {
		args["update"] = b.update
	}

	return args
}

// IdentitySetResponse represents the response from Identity/set.
type IdentitySetResponse struct {
	AccountID  string                 `json:"accountId"`
	OldState   string                 `json:"oldState"`
	NewState   string                 `json:"newState"`
	Updated    map[string]any         `json:"updated"`
	NotUpdated map[string]MethodError `json:"notUpdated"`
}

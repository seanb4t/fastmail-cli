package jmap

// EmailImportBuilder builds arguments for Email/import.
// See: https://datatracker.ietf.org/doc/html/rfc8621#section-4.8
type EmailImportBuilder struct {
	accountID string
	emails    map[string]map[string]any
}

// NewEmailImport creates a new Email/import builder.
func NewEmailImport(accountID string) *EmailImportBuilder {
	return &EmailImportBuilder{
		accountID: accountID,
		emails:    make(map[string]map[string]any),
	}
}

// Import adds an email to be imported.
// The clientID is a temporary ID used to reference this email in the response.
// blobID is the blob containing the RFC 5322 message.
// mailboxIDs maps mailbox IDs to true.
// keywords maps keyword names (e.g., "$seen") to true.
func (b *EmailImportBuilder) Import(clientID, blobID string, mailboxIDs map[string]bool, keywords map[string]bool) *EmailImportBuilder {
	entry := map[string]any{
		"blobId":     blobID,
		"mailboxIds": mailboxIDs,
	}
	if len(keywords) > 0 {
		entry["keywords"] = keywords
	}
	b.emails[clientID] = entry
	return b
}

// Build returns the arguments map for Request.Invoke.
func (b *EmailImportBuilder) Build() map[string]any {
	args := map[string]any{
		"accountId": b.accountID,
	}

	if len(b.emails) > 0 {
		args["emails"] = b.emails
	}

	return args
}

// EmailImportResponse represents the response from Email/import.
type EmailImportResponse struct {
	AccountID  string                       `json:"accountId"`
	Created    map[string]EmailImportResult `json:"created,omitempty"`
	NotCreated map[string]MethodError       `json:"notCreated,omitempty"`
}

// EmailImportResult represents a successfully imported email.
type EmailImportResult struct {
	ID       string `json:"id"`
	BlobID   string `json:"blobId"`
	ThreadID string `json:"threadId"`
	Size     uint64 `json:"size"`
}

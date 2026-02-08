package fastmail

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// MailService provides email operations.
type MailService struct {
	client *Client
}

// List returns emails from a folder.
// The folder can be a name (e.g., "Inbox", "Sent") or a mailbox ID.
// Use limit=0 for server default.
func (s *MailService) List(ctx context.Context, folder string, limit uint64) ([]Email, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	mailboxID, err := s.resolveMailbox(ctx, accountID, folder)
	if err != nil {
		return nil, oops.Wrapf(err, "resolving folder %q", folder)
	}

	// Build query to get email IDs
	queryBuilder := jmap.NewEmailQuery(accountID).
		InMailbox(mailboxID).
		SortBy("receivedAt", true) // newest first

	if limit > 0 {
		queryBuilder.Limit(limit)
	}

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	queryCallID := req.Invoke("Email/query", queryBuilder.Build())

	// Chain Email/get to fetch the actual emails
	getBuilder := jmap.NewEmailGet(accountID).
		Properties("id", "threadId", "subject", "preview", "receivedAt", "size", "keywords", "mailboxIds")
	getArgs := getBuilder.Build()
	getArgs["#ids"] = jmap.ResultReference(queryCallID, "Email/query", "/ids")
	req.Invoke("Email/get", getArgs)

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	// Check query result
	queryResult, err := resp.GetResult(queryCallID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting query result")
	}
	if queryResult.IsError() {
		return nil, oops.Errorf("query failed: %s", queryResult.Error())
	}

	// Get the email list from the second response
	getResult, err := resp.GetResult("1")
	if err != nil {
		return nil, oops.Wrapf(err, "getting email list result")
	}
	if getResult.IsError() {
		return nil, oops.Errorf("get failed: %s", getResult.Error())
	}

	var getResp jmap.EmailGetResponse
	if err := getResult.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding email response")
	}

	return convertEmails(getResp.List), nil
}

// Get returns a single email by ID.
func (s *MailService) Get(ctx context.Context, id string) (*Email, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	getBuilder := jmap.NewEmailGet(accountID).
		IDs(id).
		Properties("id", "threadId", "subject", "preview", "receivedAt", "size", "keywords", "mailboxIds")

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Email/get", getBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("get failed: %s", result.Error())
	}

	var getResp jmap.EmailGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if len(getResp.NotFound) > 0 {
		return nil, oops.Errorf("email not found: %s", id)
	}

	if len(getResp.List) == 0 {
		return nil, oops.Errorf("email not found: %s", id)
	}

	emails := convertEmails(getResp.List)
	return &emails[0], nil
}

// SearchResult pairs an email with its search snippet highlights.
type SearchResult struct {
	Email          Email
	SubjectSnippet string
	PreviewSnippet string
}

// Search returns emails matching a query string.
// The query uses JMAP filter syntax (e.g., "from:alice subject:meeting").
func (s *MailService) Search(ctx context.Context, query string, limit uint64) ([]Email, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	// Build query with text filter
	queryArgs := map[string]any{
		"accountId": accountID,
		"filter": map[string]any{
			"text": query,
		},
		"sort": []jmap.Comparator{
			{Property: "receivedAt", IsAscending: false},
		},
	}
	if limit > 0 {
		queryArgs["limit"] = limit
	}

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	queryCallID := req.Invoke("Email/query", queryArgs)

	// Chain Email/get
	getBuilder := jmap.NewEmailGet(accountID).
		Properties("id", "threadId", "subject", "preview", "receivedAt", "size", "keywords", "mailboxIds")
	getArgs := getBuilder.Build()
	getArgs["#ids"] = jmap.ResultReference(queryCallID, "Email/query", "/ids")
	req.Invoke("Email/get", getArgs)

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	// Check query result
	queryResult, err := resp.GetResult(queryCallID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting query result")
	}
	if queryResult.IsError() {
		return nil, oops.Errorf("query failed: %s", queryResult.Error())
	}

	// Get email list
	getResult, err := resp.GetResult("1")
	if err != nil {
		return nil, oops.Wrapf(err, "getting email list result")
	}
	if getResult.IsError() {
		return nil, oops.Errorf("get failed: %s", getResult.Error())
	}

	var getResp jmap.EmailGetResponse
	if err := getResult.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding email response")
	}

	return convertEmails(getResp.List), nil
}

// SearchWithSnippets returns emails matching a query string along with highlighted snippets.
// It chains Email/query, Email/get, and SearchSnippet/get in a single JMAP request.
func (s *MailService) SearchWithSnippets(ctx context.Context, query string, limit uint64) ([]SearchResult, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	// Build the filter (shared between Email/query and SearchSnippet/get)
	filter := map[string]any{
		"text": query,
	}

	// 1. Email/query
	queryArgs := map[string]any{
		"accountId": accountID,
		"filter":    filter,
		"sort": []jmap.Comparator{
			{Property: "receivedAt", IsAscending: false},
		},
	}
	if limit > 0 {
		queryArgs["limit"] = limit
	}

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	queryCallID := req.Invoke("Email/query", queryArgs)

	// 2. Email/get with back-reference to query result IDs
	getBuilder := jmap.NewEmailGet(accountID).
		Properties("id", "threadId", "subject", "preview", "receivedAt", "size", "keywords", "mailboxIds")
	getArgs := getBuilder.Build()
	getArgs["#ids"] = jmap.ResultReference(queryCallID, "Email/query", "/ids")
	req.Invoke("Email/get", getArgs)

	// 3. SearchSnippet/get with back-reference to query result IDs and same filter
	snippetArgs := jmap.NewSearchSnippetGet(accountID).
		Filter(filter).
		Build()
	snippetArgs["#emailIds"] = jmap.ResultReference(queryCallID, "Email/query", "/ids")
	req.Invoke("SearchSnippet/get", snippetArgs)

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	// Check query result
	queryResult, err := resp.GetResult(queryCallID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting query result")
	}
	if queryResult.IsError() {
		return nil, oops.Errorf("query failed: %s", queryResult.Error())
	}

	// Get email list
	getResult, err := resp.GetResult("1")
	if err != nil {
		return nil, oops.Wrapf(err, "getting email list result")
	}
	if getResult.IsError() {
		return nil, oops.Errorf("get failed: %s", getResult.Error())
	}

	var getResp jmap.EmailGetResponse
	if err := getResult.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding email response")
	}

	// Get snippet list
	snippetResult, err := resp.GetResult("2")
	if err != nil {
		return nil, oops.Wrapf(err, "getting snippet result")
	}
	if snippetResult.IsError() {
		return nil, oops.Errorf("snippet get failed: %s", snippetResult.Error())
	}

	var snippetResp jmap.SearchSnippetGetResponse
	if err := snippetResult.Decode(&snippetResp); err != nil {
		return nil, oops.Wrapf(err, "decoding snippet response")
	}

	// Build a map of emailID -> snippet for fast lookup
	snippetMap := make(map[string]jmap.SearchSnippet, len(snippetResp.List))
	for _, s := range snippetResp.List {
		snippetMap[s.EmailID] = s
	}

	// Combine emails with their snippets
	emails := convertEmails(getResp.List)
	results := make([]SearchResult, len(emails))
	for i, email := range emails {
		result := SearchResult{Email: email}
		if snippet, ok := snippetMap[email.ID]; ok {
			if snippet.Subject != nil {
				result.SubjectSnippet = *snippet.Subject
			}
			if snippet.Preview != nil {
				result.PreviewSnippet = *snippet.Preview
			}
		}
		results[i] = result
	}

	return results, nil
}

// Move moves an email to a different folder.
// The folder can be a name (e.g., "Archive") or a mailbox ID.
func (s *MailService) Move(ctx context.Context, id string, folder string) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	// First get the current email to know its mailboxes
	email, err := s.Get(ctx, id)
	if err != nil {
		return oops.Wrapf(err, "getting email")
	}

	targetMailboxID, err := s.resolveMailbox(ctx, accountID, folder)
	if err != nil {
		return oops.Wrapf(err, "resolving folder %q", folder)
	}

	// Build new mailboxIds map - remove all current, add target
	newMailboxIDs := map[string]bool{
		targetMailboxID: true,
	}

	// Build Email/set request
	setArgs := map[string]any{
		"accountId": accountID,
		"update": map[string]any{
			id: map[string]any{
				"mailboxIds": newMailboxIDs,
			},
		},
	}

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Email/set", setArgs)

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return oops.Errorf("set failed: %s", result.Error())
	}

	// Check for update errors
	var setResp struct {
		NotUpdated map[string]jmap.MethodError `json:"notUpdated"`
	}
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotUpdated[id]; ok {
		return oops.Errorf("failed to move email: %s", errInfo.Error())
	}

	// Suppress unused variable warning - email was used to validate existence
	_ = email

	return nil
}

// Delete moves an email to Trash, or permanently destroys it if already in Trash.
func (s *MailService) Delete(ctx context.Context, id string) error {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	// Get the email to check if it's already in Trash
	email, err := s.Get(ctx, id)
	if err != nil {
		return oops.Wrapf(err, "getting email")
	}

	// Find the Trash mailbox
	trashID, err := s.resolveMailbox(ctx, accountID, "Trash")
	if err != nil {
		// If no Trash folder, permanently delete
		return s.destroy(ctx, accountID, id)
	}

	// Check if already in Trash
	if slices.Contains(email.MailboxIDs, trashID) {
		// Already in Trash, permanently delete
		return s.destroy(ctx, accountID, id)
	}

	// Move to Trash
	return s.Move(ctx, id, "Trash")
}

// destroy permanently deletes an email.
func (s *MailService) destroy(ctx context.Context, accountID, id string) error {
	setArgs := map[string]any{
		"accountId": accountID,
		"destroy":   []string{id},
	}

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Email/set", setArgs)

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return oops.Errorf("destroy failed: %s", result.Error())
	}

	// Check for destroy errors
	var setResp struct {
		NotDestroyed map[string]jmap.MethodError `json:"notDestroyed"`
	}
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotDestroyed[id]; ok {
		return oops.Errorf("failed to delete email: %s", errInfo.Error())
	}

	return nil
}

// resolveMailbox converts a folder name to a mailbox ID.
// If the input looks like an ID (contains no common folder names), it's returned as-is.
func (s *MailService) resolveMailbox(ctx context.Context, accountID, folder string) (string, error) {
	// Common folder role mappings
	roleMap := map[string]string{
		"inbox":   "inbox",
		"drafts":  "drafts",
		"sent":    "sent",
		"trash":   "trash",
		"junk":    "junk",
		"spam":    "junk",
		"archive": "archive",
	}

	lowerFolder := strings.ToLower(folder)
	role, hasRole := roleMap[lowerFolder]

	// Fetch all mailboxes
	getBuilder := jmap.NewMailboxGet(accountID).
		Properties("id", "name", "role")

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Mailbox/get", getBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return "", oops.Wrapf(err, "fetching mailboxes")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return "", oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return "", oops.Errorf("mailbox get failed: %s", result.Error())
	}

	var mbResp jmap.MailboxGetResponse
	if err := result.Decode(&mbResp); err != nil {
		return "", oops.Wrapf(err, "decoding mailbox response")
	}

	// First, try to match by role
	if hasRole {
		for _, mb := range mbResp.List {
			if strings.EqualFold(mb.Role, role) {
				return mb.ID, nil
			}
		}
	}

	// Then try to match by exact name
	for _, mb := range mbResp.List {
		if strings.EqualFold(mb.Name, folder) {
			return mb.ID, nil
		}
	}

	// Finally, check if the folder looks like an ID (already resolved)
	for _, mb := range mbResp.List {
		if mb.ID == folder {
			return folder, nil
		}
	}

	return "", oops.Errorf("mailbox not found: %s", folder)
}

// convertEmails converts JMAP emails to domain emails.
func convertEmails(jmapEmails []jmap.Email) []Email {
	emails := make([]Email, len(jmapEmails))
	for i, je := range jmapEmails {
		var receivedAt time.Time
		if je.ReceivedAt != "" {
			receivedAt, _ = time.Parse(time.RFC3339, je.ReceivedAt)
		}

		mailboxIDs := make([]string, 0, len(je.MailboxIDs))
		for id := range je.MailboxIDs {
			mailboxIDs = append(mailboxIDs, id)
		}

		// Convert JMAP keywords map to keyword slice
		keywords := make([]string, 0, len(je.Keywords))
		for keyword, set := range je.Keywords {
			if set {
				keywords = append(keywords, keyword)
			}
		}

		emails[i] = Email{
			ID:         je.ID,
			ThreadID:   je.ThreadID,
			Subject:    je.Subject,
			Preview:    je.Preview,
			ReceivedAt: receivedAt,
			Size:       je.Size,
			Keywords:   keywords,
			MailboxIDs: mailboxIDs,
		}
	}
	return emails
}

// GetThread returns all emails in a thread, sorted chronologically (oldest first).
func (s *MailService) GetThread(ctx context.Context, threadID string) ([]Email, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	// Step 1: Thread/get to get the email IDs in the thread
	threadBuilder := jmap.NewThreadGet(accountID).IDs(threadID)

	threadReq := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	threadCallID := threadReq.Invoke("Thread/get", threadBuilder.Build())

	threadResp, err := s.client.jmap.Call(ctx, threadReq)
	if err != nil {
		return nil, oops.Wrapf(err, "executing Thread/get request")
	}

	threadResult, err := threadResp.GetResult(threadCallID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting thread result")
	}
	if threadResult.IsError() {
		return nil, oops.Errorf("thread get failed: %s", threadResult.Error())
	}

	var threadGetResp jmap.ThreadGetResponse
	if err := threadResult.Decode(&threadGetResp); err != nil {
		return nil, oops.Wrapf(err, "decoding thread response")
	}

	if len(threadGetResp.NotFound) > 0 {
		return nil, oops.Errorf("thread not found: %s", threadID)
	}

	if len(threadGetResp.List) == 0 {
		return nil, oops.Errorf("thread not found: %s", threadID)
	}

	emailIDs := threadGetResp.List[0].EmailIDs
	if len(emailIDs) == 0 {
		return []Email{}, nil
	}

	// Step 2: Email/get to fetch the actual emails
	getBuilder := jmap.NewEmailGet(accountID).
		IDs(emailIDs...).
		Properties("id", "threadId", "subject", "preview", "receivedAt", "size", "keywords", "mailboxIds")

	emailReq := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	emailCallID := emailReq.Invoke("Email/get", getBuilder.Build())

	emailResp, err := s.client.jmap.Call(ctx, emailReq)
	if err != nil {
		return nil, oops.Wrapf(err, "executing Email/get request")
	}

	emailResult, err := emailResp.GetResult(emailCallID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting email result")
	}
	if emailResult.IsError() {
		return nil, oops.Errorf("email get failed: %s", emailResult.Error())
	}

	var getResp jmap.EmailGetResponse
	if err := emailResult.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding email response")
	}

	emails := convertEmails(getResp.List)

	// Sort chronologically (oldest first) by ReceivedAt
	slices.SortFunc(emails, func(a, b Email) int {
		return a.ReceivedAt.Compare(b.ReceivedAt)
	})

	return emails, nil
}

// KeywordAction represents a keyword change to apply to an email.
type KeywordAction struct {
	// Keyword is the JMAP keyword (e.g., "$seen", "$flagged").
	Keyword string
	// Set is true to add the keyword, false to remove it.
	Set bool
}

// SetKeywords sets or removes keywords on an email.
func (s *MailService) SetKeywords(ctx context.Context, id string, actions []KeywordAction) error {
	if len(actions) == 0 {
		return oops.Errorf("no keyword actions provided")
	}

	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return err
	}

	// Build keyword patches
	patch := make(map[string]any, len(actions))
	for _, a := range actions {
		if a.Set {
			patch["keywords/"+a.Keyword] = true
		} else {
			patch["keywords/"+a.Keyword] = nil
		}
	}

	// Build Email/set request
	setBuilder := jmap.NewEmailSet(accountID).Update(id, patch)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Email/set", setBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return oops.Errorf("set failed: %s", result.Error())
	}

	// Check for update errors
	var setResp struct {
		NotUpdated map[string]jmap.MethodError `json:"notUpdated"`
	}
	if err := result.Decode(&setResp); err != nil {
		return oops.Wrapf(err, "decoding response")
	}

	if errInfo, ok := setResp.NotUpdated[id]; ok {
		return oops.Errorf("failed to update keywords: %s", errInfo.Error())
	}

	return nil
}

// SendOptions specifies options for sending an email.
type SendOptions struct {
	To      []EmailAddress
	Cc      []EmailAddress
	Bcc     []EmailAddress
	Subject string
	Body    string
}

// Send creates and sends a new email.
func (s *MailService) Send(ctx context.Context, opts SendOptions) (string, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return "", err
	}

	// Get the default identity for sending
	identity, err := s.getDefaultIdentity(ctx, accountID)
	if err != nil {
		return "", oops.Wrapf(err, "getting identity")
	}

	// Find the Drafts mailbox
	draftsID, err := s.resolveMailbox(ctx, accountID, "Drafts")
	if err != nil {
		return "", oops.Wrapf(err, "resolving Drafts folder")
	}

	// Build the email object
	emailData := buildEmailForSend(identity, opts, draftsID)

	// Create email and submit in one request
	return s.createAndSubmit(ctx, accountID, identity.ID, "draft", emailData)
}

// ReplyOptions specifies options for replying to an email.
type ReplyOptions struct {
	EmailID  string
	Body     string
	ReplyAll bool
}

// Reply creates and sends a reply to an existing email.
func (s *MailService) Reply(ctx context.Context, opts ReplyOptions) (string, error) {
	accountID, err := s.client.getAccountID(ctx)
	if err != nil {
		return "", err
	}

	// Get the default identity
	identity, err := s.getDefaultIdentity(ctx, accountID)
	if err != nil {
		return "", oops.Wrapf(err, "getting identity")
	}

	// Get the original email with full details
	original, err := s.getEmailForReply(ctx, accountID, opts.EmailID)
	if err != nil {
		return "", oops.Wrapf(err, "getting original email")
	}

	// Find the Drafts mailbox
	draftsID, err := s.resolveMailbox(ctx, accountID, "Drafts")
	if err != nil {
		return "", oops.Wrapf(err, "resolving Drafts folder")
	}

	// Build the reply email
	emailData := buildEmailForReply(identity, original, opts, draftsID)

	// Create and submit
	return s.createAndSubmit(ctx, accountID, identity.ID, "reply", emailData)
}

// getDefaultIdentity retrieves the first available identity for the account.
func (s *MailService) getDefaultIdentity(ctx context.Context, accountID string) (*jmap.Identity, error) {
	getBuilder := jmap.NewIdentityGet(accountID)

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapSubmission)
	callID := req.Invoke("Identity/get", getBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("identity get failed: %s", result.Error())
	}

	var identityResp jmap.IdentityGetResponse
	if err := result.Decode(&identityResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if len(identityResp.List) == 0 {
		return nil, oops.Errorf("no identities found")
	}

	return &identityResp.List[0], nil
}

// getEmailForReply retrieves an email with fields needed for reply.
func (s *MailService) getEmailForReply(ctx context.Context, accountID, emailID string) (*jmap.Email, error) {
	getBuilder := jmap.NewEmailGet(accountID).
		IDs(emailID).
		Properties(
			"id", "threadId", "subject", "from", "to", "cc", "replyTo",
			"messageId", "inReplyTo", "references", "bodyValues", "textBody",
		)

	// Request body values for quoted text
	args := getBuilder.Build()
	args["fetchTextBodyValues"] = true

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail)
	callID := req.Invoke("Email/get", args)

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return nil, oops.Wrapf(err, "executing JMAP request")
	}

	result, err := resp.GetResult(callID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting result")
	}
	if result.IsError() {
		return nil, oops.Errorf("get failed: %s", result.Error())
	}

	var getResp jmap.EmailGetResponse
	if err := result.Decode(&getResp); err != nil {
		return nil, oops.Wrapf(err, "decoding response")
	}

	if len(getResp.NotFound) > 0 || len(getResp.List) == 0 {
		return nil, oops.Errorf("email not found: %s", emailID)
	}

	return &getResp.List[0], nil
}

// createAndSubmit creates an email and submits it for delivery.
func (s *MailService) createAndSubmit(ctx context.Context, accountID, identityID, clientID string, emailData map[string]any) (string, error) {
	// Build Email/set to create the email
	emailSetBuilder := jmap.NewEmailSet(accountID).Create(clientID, emailData)

	// Get the drafts mailbox ID for the update patch
	draftsID := getDraftsKey(emailData)

	// Build EmailSubmission/set to submit the email
	// After successful send, remove draft keyword
	submissionSetBuilder := jmap.NewEmailSubmissionSet(accountID).
		Create("sub1", map[string]any{
			"identityId": identityID,
			"emailId":    "#" + clientID,
		}).
		// Remove draft keyword after successful send
		OnSuccessUpdateEmail("#"+clientID, map[string]any{
			"mailboxIds/" + draftsID: nil,
			"keywords/$draft":        nil,
		})

	req := jmap.NewRequest().WithCapabilities(jmap.CapCore, jmap.CapMail, jmap.CapSubmission)
	emailCallID := req.Invoke("Email/set", emailSetBuilder.Build())
	subCallID := req.Invoke("EmailSubmission/set", submissionSetBuilder.Build())

	resp, err := s.client.jmap.Call(ctx, req)
	if err != nil {
		return "", oops.Wrapf(err, "executing JMAP request")
	}

	// Check Email/set result
	emailResult, err := resp.GetResult(emailCallID)
	if err != nil {
		return "", oops.Wrapf(err, "getting email set result")
	}
	if emailResult.IsError() {
		return "", oops.Errorf("email create failed: %s", emailResult.Error())
	}

	var emailSetResp jmap.EmailSetResponse
	if err := emailResult.Decode(&emailSetResp); err != nil {
		return "", oops.Wrapf(err, "decoding email set response")
	}

	if errInfo, ok := emailSetResp.NotCreated[clientID]; ok {
		return "", oops.Errorf("failed to create email: %s", errInfo.Error())
	}

	createdEmail, ok := emailSetResp.Created[clientID]
	if !ok {
		return "", oops.Errorf("email not returned in created map")
	}

	// Check EmailSubmission/set result
	subResult, err := resp.GetResult(subCallID)
	if err != nil {
		return "", oops.Wrapf(err, "getting submission result")
	}
	if subResult.IsError() {
		return "", oops.Errorf("submission failed: %s", subResult.Error())
	}

	var subResp jmap.EmailSubmissionSetResponse
	if err := subResult.Decode(&subResp); err != nil {
		return "", oops.Wrapf(err, "decoding submission response")
	}

	if errInfo, ok := subResp.NotCreated["sub1"]; ok {
		return "", oops.Errorf("failed to submit email: %s", errInfo.Error())
	}

	return createdEmail.ID, nil
}

// getDraftsKey extracts the drafts mailbox ID from emailData.
func getDraftsKey(emailData map[string]any) string {
	if mailboxIDs, ok := emailData["mailboxIds"].(map[string]bool); ok {
		for id := range mailboxIDs {
			return id
		}
	}
	return ""
}

// buildEmailForSend constructs the email data map for a new email.
func buildEmailForSend(identity *jmap.Identity, opts SendOptions, draftsMailboxID string) map[string]any {
	// Convert domain addresses to JMAP addresses
	to := make([]map[string]string, len(opts.To))
	for i, addr := range opts.To {
		to[i] = map[string]string{"email": addr.Email}
		if addr.Name != "" {
			to[i]["name"] = addr.Name
		}
	}

	cc := make([]map[string]string, len(opts.Cc))
	for i, addr := range opts.Cc {
		cc[i] = map[string]string{"email": addr.Email}
		if addr.Name != "" {
			cc[i]["name"] = addr.Name
		}
	}

	bcc := make([]map[string]string, len(opts.Bcc))
	for i, addr := range opts.Bcc {
		bcc[i] = map[string]string{"email": addr.Email}
		if addr.Name != "" {
			bcc[i]["name"] = addr.Name
		}
	}

	email := map[string]any{
		"mailboxIds": map[string]bool{draftsMailboxID: true},
		"keywords":   map[string]bool{"$draft": true},
		"from":       []map[string]string{{"email": identity.Email, "name": identity.Name}},
		"to":         to,
		"subject":    opts.Subject,
		"bodyValues": map[string]any{
			"body": map[string]any{
				"value":       opts.Body,
				"charset":     "utf-8",
				"disposition": nil,
			},
		},
		"textBody": []map[string]any{
			{"partId": "body", "type": "text/plain"},
		},
	}

	if len(cc) > 0 {
		email["cc"] = cc
	}
	if len(bcc) > 0 {
		email["bcc"] = bcc
	}

	return email
}

// buildEmailForReply constructs the email data map for a reply.
func buildEmailForReply(identity *jmap.Identity, original *jmap.Email, opts ReplyOptions, draftsMailboxID string) map[string]any {
	// Determine recipients
	var to []map[string]string

	// Reply-To header takes precedence, then From
	if len(original.ReplyTo) > 0 {
		to = convertJMAPAddresses(original.ReplyTo)
	} else if len(original.From) > 0 {
		to = convertJMAPAddresses(original.From)
	}

	var cc []map[string]string
	if opts.ReplyAll {
		// Add original To (excluding ourselves) to CC
		for _, addr := range original.To {
			if !strings.EqualFold(addr.Email, identity.Email) {
				cc = append(cc, map[string]string{"email": addr.Email, "name": addr.Name})
			}
		}
		// Add original CC (excluding ourselves) to CC
		for _, addr := range original.Cc {
			if !strings.EqualFold(addr.Email, identity.Email) {
				cc = append(cc, map[string]string{"email": addr.Email, "name": addr.Name})
			}
		}
	}

	// Build subject with Re: prefix
	subject := original.Subject
	if !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}

	// Build references header for threading
	var references []string
	references = append(references, original.References...)
	if len(original.MessageID) > 0 {
		references = append(references, original.MessageID...)
	}

	// Quote the original message
	quotedBody := buildQuotedReply(original, opts.Body)

	email := map[string]any{
		"mailboxIds": map[string]bool{draftsMailboxID: true},
		"keywords":   map[string]bool{"$draft": true},
		"from":       []map[string]string{{"email": identity.Email, "name": identity.Name}},
		"to":         to,
		"subject":    subject,
		"inReplyTo":  original.MessageID,
		"references": references,
		"bodyValues": map[string]any{
			"body": map[string]any{
				"value":   quotedBody,
				"charset": "utf-8",
			},
		},
		"textBody": []map[string]any{
			{"partId": "body", "type": "text/plain"},
		},
	}

	if len(cc) > 0 {
		email["cc"] = cc
	}

	return email
}

// convertJMAPAddresses converts JMAP addresses to the map format for creation.
func convertJMAPAddresses(addrs []jmap.EmailAddress) []map[string]string {
	result := make([]map[string]string, len(addrs))
	for i, addr := range addrs {
		result[i] = map[string]string{"email": addr.Email}
		if addr.Name != "" {
			result[i]["name"] = addr.Name
		}
	}
	return result
}

// buildQuotedReply creates the reply body with quoted original.
func buildQuotedReply(original *jmap.Email, replyBody string) string {
	var originalText string

	// Extract original body text
	if len(original.TextBody) > 0 && len(original.BodyValues) > 0 {
		partID := original.TextBody[0].PartID
		if bodyValue, ok := original.BodyValues[partID]; ok {
			originalText = bodyValue.Value
		}
	}

	if originalText == "" {
		return replyBody
	}

	// Quote each line
	lines := strings.Split(originalText, "\n")
	quoted := make([]string, len(lines))
	for i, line := range lines {
		quoted[i] = "> " + line
	}

	// Build attribution
	var from string
	if len(original.From) > 0 {
		if original.From[0].Name != "" {
			from = original.From[0].Name
		} else {
			from = original.From[0].Email
		}
	}

	attribution := ""
	if from != "" {
		attribution = "On " + original.ReceivedAt + ", " + from + " wrote:\n"
	}

	return replyBody + "\n\n" + attribution + strings.Join(quoted, "\n")
}

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
		var date time.Time
		if je.ReceivedAt != "" {
			date, _ = time.Parse(time.RFC3339, je.ReceivedAt)
		}

		mailboxIDs := make([]string, 0, len(je.MailboxIDs))
		for id := range je.MailboxIDs {
			mailboxIDs = append(mailboxIDs, id)
		}

		// Convert keywords map to slice
		keywords := make([]string, 0, len(je.Keywords))
		for keyword, present := range je.Keywords {
			if present {
				keywords = append(keywords, keyword)
			}
		}

		emails[i] = Email{
			ID:         je.ID,
			ThreadID:   je.ThreadID,
			Subject:    je.Subject,
			Preview:    je.Preview,
			Date:       date,
			Size:       je.Size,
			Keywords:   keywords,
			MailboxIDs: mailboxIDs,
		}
	}
	return emails
}

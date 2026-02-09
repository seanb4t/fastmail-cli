//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

// World holds shared test state across scenario steps.
type World struct {
	mu sync.Mutex

	// Servers
	SessionServer *httptest.Server
	APIServer     *httptest.Server

	// Test data
	Mailboxes []MockMailbox
	Emails    []MockEmail

	// Internal mock state
	lastQueryIDs []string

	// Captured results from When steps
	ResultEmails  []map[string]any
	ResultError   error
	ResultSendID  string
	ToolResult    any
	ToolError     error
	SearchResults []map[string]any
}

// MockMailbox holds mailbox test data.
type MockMailbox struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// MockEmail holds email test data for mock server responses.
type MockEmail struct {
	ID         string          `json:"id"`
	ThreadID   string          `json:"threadId"`
	Subject    string          `json:"subject"`
	Preview    string          `json:"preview"`
	ReceivedAt string          `json:"receivedAt"`
	Size       uint64          `json:"size"`
	Keywords   map[string]bool `json:"keywords"`
	MailboxIDs map[string]bool `json:"mailboxIds"`
}

// worldKey is the context key for World.
type worldKey struct{}

// ContextWithWorld stores the World in the context.
func ContextWithWorld(ctx context.Context, w *World) context.Context {
	return context.WithValue(ctx, worldKey{}, w)
}

// WorldFromContext retrieves the World from the context.
func WorldFromContext(ctx context.Context) *World {
	return ctx.Value(worldKey{}).(*World)
}

// DefaultMailboxes returns the standard set of test mailboxes.
func DefaultMailboxes() []MockMailbox {
	return []MockMailbox{
		{ID: "mb-inbox", Name: "Inbox", Role: "inbox"},
		{ID: "mb-sent", Name: "Sent", Role: "sent"},
		{ID: "mb-trash", Name: "Trash", Role: "trash"},
		{ID: "mb-archive", Name: "Archive", Role: "archive"},
		{ID: "mb-drafts", Name: "Drafts", Role: "drafts"},
	}
}

// SessionResponse returns a valid JMAP session JSON for the given API URL.
func SessionResponse(apiURL string) string {
	return `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {},
			"urn:ietf:params:jmap:mail": {},
			"urn:ietf:params:jmap:submission": {}
		},
		"accounts": {
			"acc1": {
				"name": "test@example.com",
				"isPersonal": true,
				"isReadOnly": false,
				"accountCapabilities": {"urn:ietf:params:jmap:mail": {}}
			}
		},
		"primaryAccounts": {"urn:ietf:params:jmap:mail": "acc1"},
		"username": "test@example.com",
		"apiUrl": "` + apiURL + `",
		"downloadUrl": "https://example.com/download",
		"uploadUrl": "https://example.com/upload/{accountId}/",
		"eventSourceUrl": "https://example.com/events",
		"state": "s1"
	}`
}

// NewMockServers creates and starts the session and API mock servers.
func NewMockServers(w *World) {
	w.APIServer = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		handleJMAPRequest(w, rw, r)
	}))

	w.SessionServer = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(SessionResponse(w.APIServer.URL)))
	}))
}

// CloseServers shuts down the mock servers.
func CloseServers(w *World) {
	if w.APIServer != nil {
		w.APIServer.Close()
	}
	if w.SessionServer != nil {
		w.SessionServer.Close()
	}
}

// handleJMAPRequest dispatches JMAP method calls against the World's test data.
func handleJMAPRequest(w *World, rw http.ResponseWriter, r *http.Request) {
	w.mu.Lock()
	defer w.mu.Unlock()

	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	methodCalls, ok := req["methodCalls"].([]any)
	if !ok || len(methodCalls) == 0 {
		http.Error(rw, "no method calls", http.StatusBadRequest)
		return
	}

	var responses []any
	for _, rawCall := range methodCalls {
		call := rawCall.([]any)
		method := call[0].(string)
		callID := call[2].(string)

		resp := dispatchMethod(w, method, call[1].(map[string]any))
		responses = append(responses, []any{method, resp, callID})
	}

	rw.Header().Set("Content-Type", "application/json")
	result := map[string]any{
		"sessionState":    "s1",
		"methodResponses": responses,
	}
	_ = json.NewEncoder(rw).Encode(result)
}

// dispatchMethod handles a single JMAP method call.
//
//nolint:gocognit,gocyclo // test helper with method dispatch
func dispatchMethod(w *World, method string, args map[string]any) map[string]any {
	switch method {
	case "Mailbox/get":
		return mailboxGetResponse(w)
	case "Email/query":
		return emailQueryResponse(w, args)
	case "Email/get":
		return emailGetResponse(w, args)
	case "Email/set":
		return emailSetResponse(args)
	case "Identity/get":
		return identityGetResponse()
	case "EmailSubmission/set":
		return emailSubmissionSetResponse()
	case "SearchSnippet/get":
		return searchSnippetGetResponse(w)
	default:
		return map[string]any{
			"type":        "unknownMethod",
			"description": fmt.Sprintf("unknown method: %s", method),
		}
	}
}

func mailboxGetResponse(w *World) map[string]any {
	mailboxes := w.Mailboxes
	if len(mailboxes) == 0 {
		mailboxes = DefaultMailboxes()
	}
	list := make([]map[string]any, len(mailboxes))
	for i, mb := range mailboxes {
		list[i] = map[string]any{
			"id":   mb.ID,
			"name": mb.Name,
			"role": mb.Role,
		}
	}
	return map[string]any{
		"accountId": "acc1",
		"state":     "m1",
		"list":      list,
		"notFound":  []any{},
	}
}

func emailQueryResponse(w *World, args map[string]any) map[string]any {
	emails := w.Emails

	// Apply limit if specified
	limit := 0
	if l, ok := args["limit"]; ok {
		switch v := l.(type) {
		case float64:
			limit = int(v)
		}
	}

	ids := make([]string, len(emails))
	for i, e := range emails {
		ids[i] = e.ID
	}

	if limit > 0 && limit < len(ids) {
		ids = ids[:limit]
	}

	// Store query result IDs for back-reference resolution in Email/get
	w.lastQueryIDs = ids

	return map[string]any{
		"accountId":           "acc1",
		"queryState":          "q1",
		"canCalculateChanges": true,
		"position":            0,
		"ids":                 ids,
		"total":               len(ids),
	}
}

func emailGetResponse(w *World, args map[string]any) map[string]any {
	var requestedIDs []string

	if ids, ok := args["ids"].([]any); ok {
		for _, id := range ids {
			requestedIDs = append(requestedIDs, id.(string))
		}
	}

	// If no explicit IDs, use the last query result IDs (simulates back-reference)
	if len(requestedIDs) == 0 && len(w.lastQueryIDs) > 0 {
		requestedIDs = w.lastQueryIDs
	}

	var list []map[string]any
	emails := w.Emails

	if len(requestedIDs) > 0 {
		idSet := make(map[string]bool, len(requestedIDs))
		for _, id := range requestedIDs {
			idSet[id] = true
		}
		for _, e := range emails {
			if idSet[e.ID] {
				list = append(list, emailToMap(e))
			}
		}
	} else {
		for _, e := range emails {
			list = append(list, emailToMap(e))
		}
	}

	if list == nil {
		list = []map[string]any{}
	}

	return map[string]any{
		"accountId": "acc1",
		"state":     "e1",
		"list":      list,
		"notFound":  []any{},
	}
}

func emailToMap(e MockEmail) map[string]any {
	keywords := e.Keywords
	if keywords == nil {
		keywords = map[string]bool{}
	}
	mailboxIDs := e.MailboxIDs
	if mailboxIDs == nil {
		mailboxIDs = map[string]bool{"mb-inbox": true}
	}
	receivedAt := e.ReceivedAt
	if receivedAt == "" {
		receivedAt = "2024-01-15T10:30:00Z"
	}
	size := e.Size
	if size == 0 {
		size = 1234
	}
	threadID := e.ThreadID
	if threadID == "" {
		threadID = "t-" + e.ID
	}

	return map[string]any{
		"id":         e.ID,
		"threadId":   threadID,
		"subject":    e.Subject,
		"preview":    e.Preview,
		"receivedAt": receivedAt,
		"size":       size,
		"keywords":   keywords,
		"mailboxIds": mailboxIDs,
	}
}

func emailSetResponse(args map[string]any) map[string]any {
	resp := map[string]any{
		"accountId": "acc1",
		"oldState":  "e1",
		"newState":  "e2",
	}

	// Handle create (send)
	if create, ok := args["create"].(map[string]any); ok {
		created := make(map[string]any)
		for clientID := range create {
			created[clientID] = map[string]any{
				"id":       "email-created-" + clientID,
				"blobId":   "blob-" + clientID,
				"threadId": "thread-" + clientID,
			}
		}
		resp["created"] = created
	}

	// Handle update (move, flag)
	if update, ok := args["update"].(map[string]any); ok {
		updated := make(map[string]any)
		for id := range update {
			updated[id] = nil
		}
		resp["updated"] = updated
	}

	// Handle destroy (delete)
	if destroy, ok := args["destroy"].([]any); ok {
		resp["destroyed"] = destroy
	}

	return resp
}

func identityGetResponse() map[string]any {
	return map[string]any{
		"accountId": "acc1",
		"state":     "i1",
		"list": []map[string]any{
			{
				"id":    "identity1",
				"name":  "Test User",
				"email": "test@example.com",
			},
		},
		"notFound": []any{},
	}
}

func emailSubmissionSetResponse() map[string]any {
	return map[string]any{
		"accountId": "acc1",
		"oldState":  "sub1",
		"newState":  "sub2",
		"created": map[string]any{
			"sub1": map[string]any{
				"id": "submission-1",
			},
		},
	}
}

func searchSnippetGetResponse(w *World) map[string]any {
	var list []map[string]any
	for _, e := range w.Emails {
		list = append(list, map[string]any{
			"emailId": e.ID,
			"subject": "<mark>" + e.Subject + "</mark>",
			"preview": "<mark>" + e.Preview + "</mark>",
		})
	}
	if list == nil {
		list = []map[string]any{}
	}
	return map[string]any{
		"accountId": "acc1",
		"list":      list,
		"notFound":  []any{},
	}
}

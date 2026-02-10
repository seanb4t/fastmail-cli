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

// MethodHandlerFunc handles a single JMAP method call in the mock server.
type MethodHandlerFunc func(w *World, args map[string]any) map[string]any

var extensionHandlers = map[string]MethodHandlerFunc{}

// RegisterMethodHandler registers an extension JMAP method handler for integration tests.
func RegisterMethodHandler(method string, handler MethodHandlerFunc) {
	extensionHandlers[method] = handler
}

// World holds shared test state across scenario steps.
type World struct {
	mu sync.Mutex

	// Servers
	SessionServer *httptest.Server
	APIServer     *httptest.Server
	BlobServer    *httptest.Server

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

	// DomainData holds per-scenario state for domain-specific test extensions.
	// Each domain stores/retrieves typed state using its own key.
	DomainData map[string]any
}

// MockMailbox holds mailbox test data.
type MockMailbox struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// MockEmail holds email test data for mock server responses.
type MockEmail struct {
	ID          string           `json:"id"`
	ThreadID    string           `json:"threadId"`
	Subject     string           `json:"subject"`
	Preview     string           `json:"preview"`
	ReceivedAt  string           `json:"receivedAt"`
	Size        uint64           `json:"size"`
	Keywords    map[string]bool  `json:"keywords"`
	MailboxIDs  map[string]bool  `json:"mailboxIds"`
	From        *MockAddress     `json:"from,omitempty"`
	To          []MockAddress    `json:"to,omitempty"`
	Cc          []MockAddress    `json:"cc,omitempty"`
	ReplyTo     []MockAddress    `json:"replyTo,omitempty"`
	MessageID   []string         `json:"messageId,omitempty"`
	InReplyTo   []string         `json:"inReplyTo,omitempty"`
	References  []string         `json:"references,omitempty"`
	Body        string           `json:"body,omitempty"`
	Attachments []MockAttachment `json:"attachments,omitempty"`
}

// MockAttachment holds attachment test data for mock email responses.
type MockAttachment struct {
	BlobID      string `json:"blobId"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Size        uint64 `json:"size"`
	Disposition string `json:"disposition"`
}

// MockAddress holds a name/email pair for mock email addresses.
type MockAddress struct {
	Name  string `json:"name"`
	Email string `json:"email"`
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
	return SessionResponseWithBlob(apiURL, "https://example.com/download/{accountId}/{blobId}/{name}", "https://example.com/upload/{accountId}/")
}

// SessionResponseWithBlob returns a valid JMAP session JSON with configurable download/upload URLs.
func SessionResponseWithBlob(apiURL, downloadURL, uploadURL string) string {
	return `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {},
			"urn:ietf:params:jmap:mail": {},
			"urn:ietf:params:jmap:submission": {},
			"urn:ietf:params:jmap:quota": {},
			"urn:ietf:params:jmap:sieve": {}
		},
		"accounts": {
			"acc1": {
				"name": "test@example.com",
				"isPersonal": true,
				"isReadOnly": false,
				"accountCapabilities": {
					"urn:ietf:params:jmap:mail": {},
					"urn:ietf:params:jmap:sieve": {}
				}
			}
		},
		"primaryAccounts": {
			"urn:ietf:params:jmap:mail": "acc1",
			"urn:ietf:params:jmap:sieve": "acc1"
		},
		"username": "test@example.com",
		"apiUrl": "` + apiURL + `",
		"downloadUrl": "` + downloadURL + `",
		"uploadUrl": "` + uploadURL + `",
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
	if w.BlobServer != nil {
		w.BlobServer.Close()
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
	// Check extension handlers first so domain tests can override built-in responses.
	if handler, ok := extensionHandlers[method]; ok {
		return handler(w, args)
	}

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

	// Extract requested properties (JMAP RFC 8620 §5.1 property filtering)
	var properties []string
	if props, ok := args["properties"].([]any); ok {
		for _, p := range props {
			if s, ok := p.(string); ok {
				properties = append(properties, s)
			}
		}
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
				list = append(list, filterByProperties(emailToFullMap(e), properties))
			}
		}
	} else {
		for _, e := range emails {
			list = append(list, filterByProperties(emailToFullMap(e), properties))
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

// emailToFullMap returns all fields for a MockEmail. The caller filters by requested properties.
func emailToFullMap(e MockEmail) map[string]any {
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

	result := map[string]any{
		"id":         e.ID,
		"threadId":   threadID,
		"subject":    e.Subject,
		"preview":    e.Preview,
		"receivedAt": receivedAt,
		"size":       size,
		"keywords":   keywords,
		"mailboxIds": mailboxIDs,
	}

	// Address fields — only present when the mock data provides them
	if e.From != nil {
		result["from"] = []map[string]any{{"name": e.From.Name, "email": e.From.Email}}
	}
	if len(e.To) > 0 {
		to := make([]map[string]any, len(e.To))
		for i, a := range e.To {
			to[i] = map[string]any{"name": a.Name, "email": a.Email}
		}
		result["to"] = to
	}

	// Cc addresses — only present when the mock data provides them
	if len(e.Cc) > 0 {
		cc := make([]map[string]any, len(e.Cc))
		for i, a := range e.Cc {
			cc[i] = map[string]any{"name": a.Name, "email": a.Email}
		}
		result["cc"] = cc
	}

	// ReplyTo addresses — only present when the mock data provides them
	if len(e.ReplyTo) > 0 {
		replyTo := make([]map[string]any, len(e.ReplyTo))
		for i, a := range e.ReplyTo {
			replyTo[i] = map[string]any{"name": a.Name, "email": a.Email}
		}
		result["replyTo"] = replyTo
	}

	// Message threading headers — only present when mock data provides them
	if len(e.MessageID) > 0 {
		result["messageId"] = e.MessageID
	}
	if len(e.InReplyTo) > 0 {
		result["inReplyTo"] = e.InReplyTo
	}
	if len(e.References) > 0 {
		result["references"] = e.References
	}

	// Body content — only present when mock data provides it
	if e.Body != "" {
		result["bodyValues"] = map[string]any{
			"1": map[string]any{"value": e.Body, "isEncodingProblem": false, "isTruncated": false},
		}
		result["textBody"] = []map[string]any{{"partId": "1", "type": "text/plain"}}
	}

	// Attachments — only present when mock data provides them
	if len(e.Attachments) > 0 {
		atts := make([]map[string]any, len(e.Attachments))
		for i, a := range e.Attachments {
			atts[i] = map[string]any{
				"blobId":      a.BlobID,
				"name":        a.Name,
				"type":        a.Type,
				"size":        a.Size,
				"disposition": a.Disposition,
			}
		}
		result["attachments"] = atts
	}

	return result
}

// filterByProperties returns only the fields present in the requested properties list.
// Per JMAP RFC 8620 §5.1, "id" is always included.
func filterByProperties(full map[string]any, properties []string) map[string]any {
	if len(properties) == 0 {
		return full
	}
	result := map[string]any{"id": full["id"]}
	for _, prop := range properties {
		if val, ok := full[prop]; ok {
			result[prop] = val
		}
	}
	return result
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

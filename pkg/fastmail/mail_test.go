package fastmail

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// JMAP method name constants used in test assertions and mock handlers.
const (
	testMethodMailboxGet = "Mailbox/get"
	testMethodEmailGet   = "Email/get"
	testMethodEmailSet   = "Email/set"
)

// sessionResponse returns a valid JMAP session for testing.
func sessionResponse(apiURL string) string {
	return `{
		"capabilities": {
			"urn:ietf:params:jmap:core": {},
			"urn:ietf:params:jmap:mail": {}
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
		"uploadUrl": "https://example.com/upload",
		"eventSourceUrl": "https://example.com/events",
		"state": "s1"
	}`
}

func TestMailService_List(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)

		// First call should be Email/query or Mailbox/get
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		w.Header().Set("Content-Type", "application/json")

		if methodName == testMethodMailboxGet {
			// Return mailboxes response
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Mailbox/get", {
						"accountId": "acc1",
						"state": "m1",
						"list": [
							{"id": "mb-inbox", "name": "Inbox", "role": "inbox"},
							{"id": "mb-sent", "name": "Sent", "role": "sent"},
							{"id": "mb-trash", "name": "Trash", "role": "trash"}
						],
						"notFound": []
					}, "0"]
				]
			}`))
			return
		}

		// Email/query + Email/get response
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/query", {
					"accountId": "acc1",
					"queryState": "q1",
					"canCalculateChanges": true,
					"position": 0,
					"ids": ["email1", "email2"],
					"total": 2
				}, "0"],
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [
						{
							"id": "email1",
							"threadId": "t1",
							"subject": "Hello World",
							"preview": "This is a test email",
							"receivedAt": "2024-01-15T10:30:00Z",
							"size": 1234,
							"keywords": {"$seen": true},
							"mailboxIds": {"mb-inbox": true}
						},
						{
							"id": "email2",
							"threadId": "t2",
							"subject": "Another Email",
							"preview": "More content here",
							"receivedAt": "2024-01-14T09:00:00Z",
							"size": 5678,
							"keywords": {"$flagged": true},
							"mailboxIds": {"mb-inbox": true}
						}
					],
					"notFound": []
				}, "1"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	emails, err := client.Mail().List(ctx, "Inbox", 10)
	require.NoError(t, err)
	require.Len(t, emails, 2)

	assert.Equal(t, "email1", emails[0].ID)
	assert.Equal(t, "Hello World", emails[0].Subject)
	assert.True(t, emails[0].IsRead())
	assert.False(t, emails[0].IsFlagged())

	assert.Equal(t, "email2", emails[1].ID)
	assert.Equal(t, "Another Email", emails[1].Subject)
	assert.False(t, emails[1].IsRead())
	assert.True(t, emails[1].IsFlagged())
}

func TestMailService_Get(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [
						{
							"id": "email123",
							"threadId": "t1",
							"subject": "Test Subject",
							"preview": "Test preview",
							"receivedAt": "2024-01-15T10:30:00Z",
							"size": 999,
							"keywords": {"$seen": true, "$flagged": true},
							"mailboxIds": {"mb1": true}
						}
					],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	email, err := client.Mail().Get(ctx, "email123")
	require.NoError(t, err)
	require.NotNil(t, email)

	assert.Equal(t, "email123", email.ID)
	assert.Equal(t, "Test Subject", email.Subject)
	assert.True(t, email.IsRead())
	assert.True(t, email.IsFlagged())
}

func TestMailService_Get_NotFound(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [],
					"notFound": ["nonexistent"]
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	email, err := client.Mail().Get(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, email)
	assert.Contains(t, err.Error(), "not found")
}

func TestMailService_Search(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)
		filter := args["filter"].(map[string]any)

		assert.Equal(t, "meeting notes", filter["text"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/query", {
					"accountId": "acc1",
					"queryState": "q1",
					"canCalculateChanges": true,
					"position": 0,
					"ids": ["email-match"],
					"total": 1
				}, "0"],
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [
						{
							"id": "email-match",
							"threadId": "t1",
							"subject": "Meeting Notes",
							"preview": "Notes from today's meeting",
							"receivedAt": "2024-01-15T10:30:00Z",
							"size": 2000,
							"keywords": {},
							"mailboxIds": {"mb1": true}
						}
					],
					"notFound": []
				}, "1"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	emails, err := client.Mail().Search(ctx, "meeting notes", 10)
	require.NoError(t, err)
	require.Len(t, emails, 1)

	assert.Equal(t, "email-match", emails[0].ID)
	assert.Equal(t, "Meeting Notes", emails[0].Subject)
}

func TestMailService_Move(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		w.Header().Set("Content-Type", "application/json")

		switch methodName {
		case testMethodMailboxGet:
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Mailbox/get", {
						"accountId": "acc1",
						"state": "m1",
						"list": [
							{"id": "mb-inbox", "name": "Inbox", "role": "inbox"},
							{"id": "mb-archive", "name": "Archive", "role": "archive"}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case testMethodEmailGet:
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Email/get", {
						"accountId": "acc1",
						"state": "e1",
						"list": [
							{
								"id": "email-to-move",
								"threadId": "t1",
								"subject": "Move Me",
								"preview": "...",
								"receivedAt": "2024-01-15T10:30:00Z",
								"size": 100,
								"keywords": {},
								"mailboxIds": {"mb-inbox": true}
							}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case testMethodEmailSet:
			// Verify the update payload
			args := firstCall[1].(map[string]any)
			update := args["update"].(map[string]any)
			emailUpdate := update["email-to-move"].(map[string]any)
			mbIDs := emailUpdate["mailboxIds"].(map[string]any)

			assert.True(t, mbIDs["mb-archive"].(bool))

			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Email/set", {
						"accountId": "acc1",
						"oldState": "e1",
						"newState": "e2",
						"updated": {"email-to-move": null}
					}, "0"]
				]
			}`))
		}
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.Mail().Move(ctx, "email-to-move", "Archive")
	require.NoError(t, err)
}

func TestMailService_Delete_MovesToTrash(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		w.Header().Set("Content-Type", "application/json")

		switch methodName {
		case testMethodMailboxGet:
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Mailbox/get", {
						"accountId": "acc1",
						"state": "m1",
						"list": [
							{"id": "mb-inbox", "name": "Inbox", "role": "inbox"},
							{"id": "mb-trash", "name": "Trash", "role": "trash"}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case testMethodEmailGet:
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Email/get", {
						"accountId": "acc1",
						"state": "e1",
						"list": [
							{
								"id": "email-to-delete",
								"threadId": "t1",
								"subject": "Delete Me",
								"preview": "...",
								"receivedAt": "2024-01-15T10:30:00Z",
								"size": 100,
								"keywords": {},
								"mailboxIds": {"mb-inbox": true}
							}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case testMethodEmailSet:
			args := firstCall[1].(map[string]any)
			// Should be an update (move to trash), not destroy
			_, hasUpdate := args["update"]
			assert.True(t, hasUpdate, "expected update operation (move to trash)")

			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Email/set", {
						"accountId": "acc1",
						"oldState": "e1",
						"newState": "e2",
						"updated": {"email-to-delete": null}
					}, "0"]
				]
			}`))
		}
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.Mail().Delete(ctx, "email-to-delete")
	require.NoError(t, err)
}

func TestMailService_Delete_PermanentlyDestroysFromTrash(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		w.Header().Set("Content-Type", "application/json")

		switch methodName {
		case testMethodMailboxGet:
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Mailbox/get", {
						"accountId": "acc1",
						"state": "m1",
						"list": [
							{"id": "mb-inbox", "name": "Inbox", "role": "inbox"},
							{"id": "mb-trash", "name": "Trash", "role": "trash"}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case testMethodEmailGet:
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Email/get", {
						"accountId": "acc1",
						"state": "e1",
						"list": [
							{
								"id": "email-in-trash",
								"threadId": "t1",
								"subject": "Already in Trash",
								"preview": "...",
								"receivedAt": "2024-01-15T10:30:00Z",
								"size": 100,
								"keywords": {},
								"mailboxIds": {"mb-trash": true}
							}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case testMethodEmailSet:
			args := firstCall[1].(map[string]any)
			// Should be a destroy, not update
			destroy, hasDestroy := args["destroy"].([]any)
			assert.True(t, hasDestroy, "expected destroy operation")
			assert.Contains(t, destroy, "email-in-trash")

			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Email/set", {
						"accountId": "acc1",
						"oldState": "e1",
						"newState": "e2",
						"destroyed": ["email-in-trash"]
					}, "0"]
				]
			}`))
		}
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.Mail().Delete(ctx, "email-in-trash")
	require.NoError(t, err)
}

func TestClient_Connect(t *testing.T) {
	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse("https://api.example.com/jmap/api/")))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)
}

func TestClient_Connect_NoMailAccount(t *testing.T) {
	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"capabilities": {},
			"accounts": {},
			"primaryAccounts": {},
			"username": "test",
			"apiUrl": "https://example.com/api",
			"downloadUrl": "",
			"uploadUrl": "",
			"eventSourceUrl": "",
			"state": ""
		}`))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no mail account")
}

func TestResolveMailbox_ByRole(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/get", {
					"accountId": "acc1",
					"state": "m1",
					"list": [
						{"id": "mb-123", "name": "INBOX", "role": "inbox"},
						{"id": "mb-456", "name": "Junk Email", "role": "junk"}
					],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	// Initialize the client
	err := client.Connect(ctx)
	require.NoError(t, err)

	// Test that "spam" maps to junk role
	mailSvc := client.Mail()
	mailboxID, err := mailSvc.resolveMailbox(ctx, "acc1", "spam")
	require.NoError(t, err)
	assert.Equal(t, "mb-456", mailboxID)
}

func TestMailService_Send(t *testing.T) {
	requestCount := 0
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		w.Header().Set("Content-Type", "application/json")
		requestCount++

		switch methodName {
		case "Identity/get":
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Identity/get", {
						"accountId": "acc1",
						"state": "i1",
						"list": [
							{
								"id": "identity1",
								"name": "Test User",
								"email": "test@example.com"
							}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case testMethodMailboxGet:
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Mailbox/get", {
						"accountId": "acc1",
						"state": "m1",
						"list": [
							{"id": "mb-drafts", "name": "Drafts", "role": "drafts"},
							{"id": "mb-sent", "name": "Sent", "role": "sent"}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case testMethodEmailSet:
			// Verify the create payload has required fields
			args := firstCall[1].(map[string]any)
			create := args["create"].(map[string]any)
			draft := create["draft"].(map[string]any)

			assert.Equal(t, "Test Subject", draft["subject"])
			to := draft["to"].([]any)
			assert.Len(t, to, 1)

			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Email/set", {
						"accountId": "acc1",
						"oldState": "e1",
						"newState": "e2",
						"created": {
							"draft": {
								"id": "email-new-123",
								"blobId": "blob123",
								"threadId": "thread123"
							}
						}
					}, "0"],
					["EmailSubmission/set", {
						"accountId": "acc1",
						"oldState": "sub1",
						"newState": "sub2",
						"created": {
							"sub1": {
								"id": "submission123"
							}
						}
					}, "1"]
				]
			}`))
		}
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	opts := SendOptions{
		To:      []EmailAddress{{Email: "recipient@example.com"}},
		Subject: "Test Subject",
		Body:    "Test body content",
	}

	emailID, err := client.Mail().Send(ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, "email-new-123", emailID)
}

func TestMailService_SetKeywords(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		actions []KeywordAction
		// expectedPatch maps the expected JMAP keyword patch keys to values
		expectedPatch map[string]any
	}{
		{
			name: "mark as read",
			id:   "email-flag-1",
			actions: []KeywordAction{
				{Keyword: "$seen", Set: true},
			},
			expectedPatch: map[string]any{
				"keywords/$seen": true,
			},
		},
		{
			name: "mark as unread",
			id:   "email-flag-2",
			actions: []KeywordAction{
				{Keyword: "$seen", Set: false},
			},
			expectedPatch: map[string]any{
				"keywords/$seen": nil,
			},
		},
		{
			name: "star email",
			id:   "email-flag-3",
			actions: []KeywordAction{
				{Keyword: "$flagged", Set: true},
			},
			expectedPatch: map[string]any{
				"keywords/$flagged": true,
			},
		},
		{
			name: "multiple keywords",
			id:   "email-flag-4",
			actions: []KeywordAction{
				{Keyword: "$seen", Set: true},
				{Keyword: "$flagged", Set: false},
			},
			expectedPatch: map[string]any{
				"keywords/$seen":    true,
				"keywords/$flagged": nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req map[string]any
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)

				methodCalls := req["methodCalls"].([]any)
				firstCall := methodCalls[0].([]any)
				methodName := firstCall[0].(string)

				w.Header().Set("Content-Type", "application/json")

				assert.Equal(t, testMethodEmailSet, methodName)

				// Verify the update payload has the expected keyword patches
				args := firstCall[1].(map[string]any)
				update := args["update"].(map[string]any)
				emailUpdate := update[tt.id].(map[string]any)

				for key, expectedVal := range tt.expectedPatch {
					actualVal, exists := emailUpdate[key]
					assert.True(t, exists, "expected patch key %q", key)
					if expectedVal == nil {
						assert.Nil(t, actualVal, "expected nil for key %q", key)
					} else {
						assert.Equal(t, expectedVal, actualVal, "unexpected value for key %q", key)
					}
				}

				_, _ = w.Write([]byte(`{
					"sessionState": "s1",
					"methodResponses": [
						["Email/set", {
							"accountId": "acc1",
							"oldState": "e1",
							"newState": "e2",
							"updated": {"` + tt.id + `": null}
						}, "0"]
					]
				}`))
			}))
			defer apiServer.Close()

			sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
			}))
			defer sessionServer.Close()

			client := NewClient(sessionServer.URL, "test-token")
			ctx := context.Background()

			err := client.Mail().SetKeywords(ctx, tt.id, tt.actions)
			require.NoError(t, err)
		})
	}
}

func TestMailService_SetKeywords_NotUpdated(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/set", {
					"accountId": "acc1",
					"oldState": "e1",
					"newState": "e1",
					"notUpdated": {
						"bad-email-id": {
							"type": "notFound",
							"description": "Email not found"
						}
					}
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	err := client.Mail().SetKeywords(ctx, "bad-email-id", []KeywordAction{
		{Keyword: "$seen", Set: true},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update keywords")
}

func TestMailService_Reply(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		w.Header().Set("Content-Type", "application/json")

		switch methodName {
		case "Identity/get":
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Identity/get", {
						"accountId": "acc1",
						"state": "i1",
						"list": [
							{
								"id": "identity1",
								"name": "Test User",
								"email": "me@example.com"
							}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case testMethodEmailGet:
			args := firstCall[1].(map[string]any)
			ids := args["ids"].([]any)
			if ids[0] == "original-email-123" {
				// Original email for reply
				_, _ = w.Write([]byte(`{
					"sessionState": "s1",
					"methodResponses": [
						["Email/get", {
							"accountId": "acc1",
							"state": "e1",
							"list": [
								{
									"id": "original-email-123",
									"threadId": "thread-orig",
									"subject": "Original Subject",
									"from": [{"name": "Sender", "email": "sender@example.com"}],
									"to": [{"email": "me@example.com"}],
									"cc": [],
									"messageId": ["<msg123@example.com>"],
									"inReplyTo": [],
									"references": [],
									"receivedAt": "2024-01-15T10:30:00Z",
									"bodyValues": {
										"1": {"value": "Original message body"}
									},
									"textBody": [{"partId": "1", "type": "text/plain"}]
								}
							],
							"notFound": []
						}, "0"]
					]
				}`))
			}
		case testMethodMailboxGet:
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Mailbox/get", {
						"accountId": "acc1",
						"state": "m1",
						"list": [
							{"id": "mb-drafts", "name": "Drafts", "role": "drafts"}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case testMethodEmailSet:
			// Verify the reply has proper threading
			args := firstCall[1].(map[string]any)
			create := args["create"].(map[string]any)
			reply := create["reply"].(map[string]any)

			subject := reply["subject"].(string)
			assert.True(t, len(subject) > 0 && (subject[:4] == "Re: " || subject == "Re: Original Subject"))

			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Email/set", {
						"accountId": "acc1",
						"oldState": "e1",
						"newState": "e2",
						"created": {
							"reply": {
								"id": "reply-email-456",
								"blobId": "blob456",
								"threadId": "thread-orig"
							}
						}
					}, "0"],
					["EmailSubmission/set", {
						"accountId": "acc1",
						"oldState": "sub1",
						"newState": "sub2",
						"created": {
							"sub1": {
								"id": "submission456"
							}
						}
					}, "1"]
				]
			}`))
		}
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	opts := ReplyOptions{
		EmailID:  "original-email-123",
		Body:     "This is my reply",
		ReplyAll: false,
	}

	replyID, err := client.Mail().Reply(ctx, opts)
	require.NoError(t, err)
	assert.Equal(t, "reply-email-456", replyID)
}

func TestMailService_GetThread(t *testing.T) {
	requestNum := 0
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		w.Header().Set("Content-Type", "application/json")
		requestNum++

		switch methodName {
		case "Thread/get":
			// Verify the thread ID is passed correctly
			args := firstCall[1].(map[string]any)
			ids := args["ids"].([]any)
			assert.Equal(t, "thread-abc", ids[0])

			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Thread/get", {
						"accountId": "acc1",
						"state": "t1",
						"list": [
							{
								"id": "thread-abc",
								"emailIds": ["email-1", "email-2", "email-3"]
							}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case testMethodEmailGet:
			// Return the emails in the thread (intentionally out of chronological order)
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Email/get", {
						"accountId": "acc1",
						"state": "e1",
						"list": [
							{
								"id": "email-2",
								"threadId": "thread-abc",
								"subject": "Re: Thread Subject",
								"preview": "Second email",
								"receivedAt": "2024-01-15T12:00:00Z",
								"size": 2000,
								"keywords": {"$seen": true},
								"mailboxIds": {"mb-inbox": true}
							},
							{
								"id": "email-1",
								"threadId": "thread-abc",
								"subject": "Thread Subject",
								"preview": "First email",
								"receivedAt": "2024-01-15T10:00:00Z",
								"size": 1000,
								"keywords": {"$seen": true},
								"mailboxIds": {"mb-inbox": true}
							},
							{
								"id": "email-3",
								"threadId": "thread-abc",
								"subject": "Re: Re: Thread Subject",
								"preview": "Third email",
								"receivedAt": "2024-01-15T14:00:00Z",
								"size": 3000,
								"keywords": {},
								"mailboxIds": {"mb-inbox": true}
							}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		}
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	emails, err := client.Mail().GetThread(ctx, "thread-abc")
	require.NoError(t, err)
	require.Len(t, emails, 3)

	// Should be sorted chronologically (oldest first)
	assert.Equal(t, "email-1", emails[0].ID)
	assert.Equal(t, "Thread Subject", emails[0].Subject)

	assert.Equal(t, "email-2", emails[1].ID)
	assert.Equal(t, "Re: Thread Subject", emails[1].Subject)

	assert.Equal(t, "email-3", emails[2].ID)
	assert.Equal(t, "Re: Re: Thread Subject", emails[2].Subject)
}

func TestMailService_GetThread_NotFound(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Thread/get", {
					"accountId": "acc1",
					"state": "t1",
					"list": [],
					"notFound": ["thread-missing"]
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	emails, err := client.Mail().GetThread(ctx, "thread-missing")
	assert.Error(t, err)
	assert.Nil(t, emails)
	assert.Contains(t, err.Error(), "not found")
}

func TestMailService_Attachments(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [
						{
							"id": "email-with-att",
							"attachments": [
								{
									"blobId": "blob-pdf-1",
									"name": "report.pdf",
									"type": "application/pdf",
									"size": 102400,
									"disposition": "attachment"
								},
								{
									"blobId": "blob-img-2",
									"name": "photo.jpg",
									"type": "image/jpeg",
									"size": 51200,
									"disposition": "attachment"
								}
							]
						}
					],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	attachments, err := client.Mail().Attachments(ctx, "email-with-att")
	require.NoError(t, err)
	require.Len(t, attachments, 2)

	assert.Equal(t, "blob-pdf-1", attachments[0].BlobID)
	assert.Equal(t, "report.pdf", attachments[0].Name)
	assert.Equal(t, "application/pdf", attachments[0].Type)
	assert.Equal(t, uint64(102400), attachments[0].Size)
	assert.Equal(t, "attachment", attachments[0].Disposition)

	assert.Equal(t, "blob-img-2", attachments[1].BlobID)
	assert.Equal(t, "photo.jpg", attachments[1].Name)
	assert.Equal(t, "image/jpeg", attachments[1].Type)
}

func TestMailService_Attachments_NotFound(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [],
					"notFound": ["nonexistent"]
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	attachments, err := client.Mail().Attachments(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, attachments)
	assert.Contains(t, err.Error(), "not found")
}

func TestMailService_Attachments_Empty(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [
						{
							"id": "email-no-att",
							"attachments": []
						}
					],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	attachments, err := client.Mail().Attachments(ctx, "email-no-att")
	require.NoError(t, err)
	assert.Nil(t, attachments)
}

func TestMailService_Get_WithAttachments(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [
						{
							"id": "email-full",
							"threadId": "t1",
							"subject": "Email With Attachment",
							"preview": "See attached",
							"receivedAt": "2024-01-15T10:30:00Z",
							"size": 51200,
							"keywords": {"$seen": true},
							"mailboxIds": {"mb1": true},
							"attachments": [
								{
									"blobId": "blob-doc-1",
									"name": "document.docx",
									"type": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
									"size": 25600,
									"disposition": "attachment"
								}
							]
						}
					],
					"notFound": []
				}, "0"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	email, err := client.Mail().Get(ctx, "email-full")
	require.NoError(t, err)
	require.NotNil(t, email)

	assert.Equal(t, "email-full", email.ID)
	assert.Equal(t, "Email With Attachment", email.Subject)
	require.Len(t, email.Attachments, 1)
	assert.Equal(t, "blob-doc-1", email.Attachments[0].BlobID)
	assert.Equal(t, "document.docx", email.Attachments[0].Name)
}

func TestMailService_DownloadAttachment(t *testing.T) {
	blobContent := "fake-pdf-binary-data"

	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "blob-pdf-1")
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(blobContent))
	}))
	defer downloadServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"capabilities": {
				"urn:ietf:params:jmap:core": {},
				"urn:ietf:params:jmap:mail": {}
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
			"apiUrl": "https://api.example.com/jmap/api/",
			"downloadUrl": "` + downloadServer.URL + `/{accountId}/{blobId}/{name}",
			"uploadUrl": "https://api.example.com/jmap/upload/{accountId}/",
			"eventSourceUrl": "https://api.example.com/jmap/eventsource/",
			"state": "s1"
		}`))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()
	require.NoError(t, client.Connect(ctx))

	reader, err := client.Mail().DownloadAttachment(ctx, "blob-pdf-1", "report.pdf")
	require.NoError(t, err)
	require.NotNil(t, reader)
	defer func() { _ = reader.Close() }()

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, blobContent, string(data))
}

func TestMailService_Import(t *testing.T) {
	requestCount := 0
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")

		// Request 1: Mailbox/get (resolve folder)
		// Request 2: Upload blob (different content-type)
		// Request 3: Email/import
		if r.Header.Get("Content-Type") == "message/rfc822" {
			// Upload blob response
			_, _ = w.Write([]byte(`{
				"accountId": "acc1",
				"blobId": "blob-uploaded-1",
				"type": "message/rfc822",
				"size": 500
			}`))
			return
		}

		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		switch methodName {
		case testMethodMailboxGet:
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Mailbox/get", {
						"accountId": "acc1",
						"state": "m1",
						"list": [
							{"id": "mb-inbox", "name": "Inbox", "role": "inbox"},
							{"id": "mb-archive", "name": "Archive", "role": "archive"}
						],
						"notFound": []
					}, "0"]
				]
			}`))
		case "Email/import":
			args := firstCall[1].(map[string]any)
			emails := args["emails"].(map[string]any)
			imp1 := emails["imp1"].(map[string]any)

			assert.Equal(t, "blob-uploaded-1", imp1["blobId"])
			mbIDs := imp1["mailboxIds"].(map[string]any)
			assert.True(t, mbIDs["mb-inbox"].(bool))

			keywords := imp1["keywords"].(map[string]any)
			assert.True(t, keywords["$seen"].(bool))

			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Email/import", {
						"accountId": "acc1",
						"created": {
							"imp1": {
								"id": "email-imported-1",
								"blobId": "blob-imported-1",
								"threadId": "thread-imported-1",
								"size": 500
							}
						}
					}, "0"]
				]
			}`))
		}
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"capabilities": {
				"urn:ietf:params:jmap:core": {},
				"urn:ietf:params:jmap:mail": {}
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
			"apiUrl": "` + apiServer.URL + `",
			"downloadUrl": "https://example.com/download",
			"uploadUrl": "` + apiServer.URL + `/{accountId}/",
			"eventSourceUrl": "https://example.com/events",
			"state": "s1"
		}`))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	msgData := strings.NewReader("From: test@example.com\r\nSubject: Test\r\n\r\nBody")
	opts := ImportOptions{
		Folder:   "Inbox",
		Keywords: map[string]bool{"$seen": true},
	}

	result, err := client.Mail().Import(ctx, msgData, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "email-imported-1", result.ID)
	assert.Equal(t, "blob-imported-1", result.BlobID)
	assert.Equal(t, "thread-imported-1", result.ThreadID)
	assert.Equal(t, uint64(500), result.Size)
}

func TestMailService_Import_DefaultFolder(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Header.Get("Content-Type") == "message/rfc822" {
			_, _ = w.Write([]byte(`{
				"accountId": "acc1",
				"blobId": "blob-up-2",
				"type": "message/rfc822",
				"size": 100
			}`))
			return
		}

		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		switch methodName {
		case testMethodMailboxGet:
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Mailbox/get", {
						"accountId": "acc1",
						"state": "m1",
						"list": [{"id": "mb-inbox", "name": "Inbox", "role": "inbox"}],
						"notFound": []
					}, "0"]
				]
			}`))
		case "Email/import":
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Email/import", {
						"accountId": "acc1",
						"created": {
							"imp1": {
								"id": "email-imp-2",
								"blobId": "blob-imp-2",
								"threadId": "thread-imp-2",
								"size": 100
							}
						}
					}, "0"]
				]
			}`))
		}
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"capabilities": {
				"urn:ietf:params:jmap:core": {},
				"urn:ietf:params:jmap:mail": {}
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
			"apiUrl": "` + apiServer.URL + `",
			"downloadUrl": "https://example.com/download",
			"uploadUrl": "` + apiServer.URL + `/{accountId}/",
			"eventSourceUrl": "https://example.com/events",
			"state": "s1"
		}`))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	// Empty folder should default to Inbox
	result, err := client.Mail().Import(ctx, strings.NewReader("msg"), ImportOptions{})
	require.NoError(t, err)
	assert.Equal(t, "email-imp-2", result.ID)
}

func TestMailService_SearchWithSnippets(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)

		// Should have 3 method calls: Email/query, Email/get, SearchSnippet/get
		require.Len(t, methodCalls, 3)

		firstCall := methodCalls[0].([]any)
		assert.Equal(t, "Email/query", firstCall[0])

		secondCall := methodCalls[1].([]any)
		assert.Equal(t, "Email/get", secondCall[0])

		thirdCall := methodCalls[2].([]any)
		assert.Equal(t, "SearchSnippet/get", thirdCall[0])

		// Verify the filter in Email/query
		queryArgs := firstCall[1].(map[string]any)
		queryFilter := queryArgs["filter"].(map[string]any)
		assert.Equal(t, "meeting notes", queryFilter["text"])

		// Verify the filter in SearchSnippet/get matches
		snippetArgs := thirdCall[1].(map[string]any)
		snippetFilter := snippetArgs["filter"].(map[string]any)
		assert.Equal(t, "meeting notes", snippetFilter["text"])

		// Verify back-references
		getArgs := secondCall[1].(map[string]any)
		idsRef := getArgs["#ids"].(map[string]any)
		assert.Equal(t, "0", idsRef["resultOf"])
		assert.Equal(t, "Email/query", idsRef["name"])
		assert.Equal(t, "/ids", idsRef["path"])

		emailIDsRef := snippetArgs["#emailIds"].(map[string]any)
		assert.Equal(t, "0", emailIDsRef["resultOf"])
		assert.Equal(t, "Email/query", emailIDsRef["name"])
		assert.Equal(t, "/ids", emailIDsRef["path"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/query", {
					"accountId": "acc1",
					"queryState": "q1",
					"canCalculateChanges": true,
					"position": 0,
					"ids": ["email-1", "email-2"],
					"total": 2
				}, "0"],
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [
						{
							"id": "email-1",
							"threadId": "t1",
							"subject": "Meeting Notes",
							"preview": "Notes from today's meeting about project X",
							"receivedAt": "2024-01-15T10:30:00Z",
							"size": 2000,
							"keywords": {"$seen": true},
							"mailboxIds": {"mb1": true}
						},
						{
							"id": "email-2",
							"threadId": "t2",
							"subject": "Another Email",
							"preview": "Some meeting content here",
							"receivedAt": "2024-01-14T09:00:00Z",
							"size": 1500,
							"keywords": {},
							"mailboxIds": {"mb1": true}
						}
					],
					"notFound": []
				}, "1"],
				["SearchSnippet/get", {
					"accountId": "acc1",
					"list": [
						{
							"emailId": "email-1",
							"subject": "<mark>Meeting</mark> <mark>Notes</mark>",
							"preview": "<mark>Notes</mark> from today's <mark>meeting</mark> about project X"
						},
						{
							"emailId": "email-2",
							"subject": null,
							"preview": "Some <mark>meeting</mark> content here"
						}
					],
					"notFound": []
				}, "2"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	results, err := client.Mail().SearchWithSnippets(ctx, "meeting notes", 10)
	require.NoError(t, err)
	require.Len(t, results, 2)

	// First result: has both subject and preview snippets
	assert.Equal(t, "email-1", results[0].Email.ID)
	assert.Equal(t, "Meeting Notes", results[0].Email.Subject)
	assert.Equal(t, "<mark>Meeting</mark> <mark>Notes</mark>", results[0].SubjectSnippet)
	assert.Equal(t, "<mark>Notes</mark> from today's <mark>meeting</mark> about project X", results[0].PreviewSnippet)

	// Second result: subject snippet is null (empty string), preview has snippet
	assert.Equal(t, "email-2", results[1].Email.ID)
	assert.Equal(t, "", results[1].SubjectSnippet)
	assert.Equal(t, "Some <mark>meeting</mark> content here", results[1].PreviewSnippet)
}

func TestMailService_SearchWithSnippets_NoResults(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/query", {
					"accountId": "acc1",
					"queryState": "q1",
					"canCalculateChanges": true,
					"position": 0,
					"ids": [],
					"total": 0
				}, "0"],
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [],
					"notFound": []
				}, "1"],
				["SearchSnippet/get", {
					"accountId": "acc1",
					"list": [],
					"notFound": []
				}, "2"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	results, err := client.Mail().SearchWithSnippets(ctx, "nonexistent query", 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestMailService_SearchWithSnippets_SnippetError(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/query", {
					"accountId": "acc1",
					"queryState": "q1",
					"canCalculateChanges": true,
					"position": 0,
					"ids": ["email-1"],
					"total": 1
				}, "0"],
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [
						{
							"id": "email-1",
							"threadId": "t1",
							"subject": "Test",
							"preview": "Test preview",
							"receivedAt": "2024-01-15T10:30:00Z",
							"size": 100,
							"keywords": {},
							"mailboxIds": {"mb1": true}
						}
					],
					"notFound": []
				}, "1"],
				["error", {
					"type": "serverFail",
					"description": "snippet generation failed"
				}, "2"]
			]
		}`))
	}))
	defer apiServer.Close()

	sessionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	results, err := client.Mail().SearchWithSnippets(ctx, "test", 10)
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "snippet")
}

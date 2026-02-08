package fastmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

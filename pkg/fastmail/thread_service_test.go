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

func TestThreadService_Get(t *testing.T) {
	callCount := 0

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		w.Header().Set("Content-Type", "application/json")
		callCount++

		if methodName == "Thread/get" {
			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Thread/get", {
						"accountId": "acc1",
						"state": "t1",
						"list": [
							{
								"id": "thread-1",
								"emailIds": ["email-a", "email-b", "email-c"]
							}
						],
						"notFound": []
					}, "0"]
				]
			}`))
			return
		}

		// Email/get response
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Email/get", {
					"accountId": "acc1",
					"state": "e1",
					"list": [
						{
							"id": "email-a",
							"threadId": "thread-1",
							"subject": "Original message",
							"from": [{"name": "Alice", "email": "alice@example.com"}],
							"receivedAt": "2024-01-15T10:00:00Z",
							"preview": "First message",
							"size": 1000,
							"keywords": {},
							"mailboxIds": {"mb-inbox": true}
						},
						{
							"id": "email-b",
							"threadId": "thread-1",
							"subject": "Re: Original message",
							"from": [{"name": "Bob", "email": "bob@example.com"}],
							"receivedAt": "2024-01-15T11:00:00Z",
							"preview": "Reply from Bob",
							"size": 1500,
							"keywords": {"$seen": true},
							"mailboxIds": {"mb-inbox": true}
						},
						{
							"id": "email-c",
							"threadId": "thread-1",
							"subject": "Re: Original message",
							"from": [{"name": "Alice", "email": "alice@example.com"}],
							"receivedAt": "2024-01-15T12:00:00Z",
							"preview": "Follow-up from Alice",
							"size": 2000,
							"keywords": {"$seen": true, "$flagged": true},
							"mailboxIds": {"mb-inbox": true}
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

	emails, err := client.Thread().Get(ctx, "thread-1")
	require.NoError(t, err)
	require.Len(t, emails, 3)

	assert.Equal(t, "email-a", emails[0].ID)
	assert.Equal(t, "Original message", emails[0].Subject)
	assert.False(t, emails[0].IsRead())

	assert.Equal(t, "email-b", emails[1].ID)
	assert.Equal(t, "Re: Original message", emails[1].Subject)
	assert.True(t, emails[1].IsRead())

	assert.Equal(t, "email-c", emails[2].ID)
	assert.True(t, emails[2].IsFlagged())

	// Should have made 2 API calls: Thread/get then Email/get
	assert.Equal(t, 2, callCount)
}

func TestThreadService_Get_NotFound(t *testing.T) {
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

	_, err := client.Thread().Get(ctx, "thread-missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "thread not found")
}

func TestThreadService_Get_EmptyThread(t *testing.T) {
	callCount := 0

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Thread/get", {
					"accountId": "acc1",
					"state": "t1",
					"list": [
						{
							"id": "thread-empty",
							"emailIds": []
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

	emails, err := client.Thread().Get(ctx, "thread-empty")
	require.NoError(t, err)
	assert.Empty(t, emails)

	// Should only make 1 API call (no need for Email/get with no emailIds)
	assert.Equal(t, 1, callCount)
}

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

const testMethodMailboxSet = "Mailbox/set"

func TestMailboxService_List(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/get", {
					"accountId": "acc1",
					"state": "m1",
					"list": [
						{
							"id": "mb-inbox",
							"name": "Inbox",
							"role": "inbox",
							"sortOrder": 1,
							"totalEmails": 150,
							"unreadEmails": 10,
							"totalThreads": 100,
							"unreadThreads": 5
						},
						{
							"id": "mb-sent",
							"name": "Sent",
							"role": "sent",
							"sortOrder": 2,
							"totalEmails": 50,
							"unreadEmails": 0,
							"totalThreads": 30,
							"unreadThreads": 0
						},
						{
							"id": "mb-custom",
							"name": "Projects",
							"parentId": "mb-inbox",
							"sortOrder": 10,
							"totalEmails": 25,
							"unreadEmails": 3,
							"totalThreads": 15,
							"unreadThreads": 2
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

	mailboxes, err := client.Mailbox().List(ctx)
	require.NoError(t, err)
	require.Len(t, mailboxes, 3)

	assert.Equal(t, "mb-inbox", mailboxes[0].ID)
	assert.Equal(t, "Inbox", mailboxes[0].Name)
	assert.Equal(t, MailboxRole("inbox"), mailboxes[0].Role)
	assert.Equal(t, uint64(150), mailboxes[0].TotalEmails)
	assert.Equal(t, uint64(10), mailboxes[0].UnreadEmails)
	assert.True(t, mailboxes[0].HasUnread())

	assert.Equal(t, "mb-sent", mailboxes[1].ID)
	assert.Equal(t, "Sent", mailboxes[1].Name)
	assert.Equal(t, MailboxRole("sent"), mailboxes[1].Role)
	assert.False(t, mailboxes[1].HasUnread())

	assert.Equal(t, "mb-custom", mailboxes[2].ID)
	assert.Equal(t, "Projects", mailboxes[2].Name)
	assert.Equal(t, MailboxRole(""), mailboxes[2].Role)
	assert.Equal(t, "mb-inbox", mailboxes[2].ParentID)
}

func TestMailboxService_Create(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		w.Header().Set("Content-Type", "application/json")

		if methodName == testMethodMailboxSet {
			args := firstCall[1].(map[string]any)
			create := args["create"].(map[string]any)
			newMB := create["new-mailbox"].(map[string]any)

			assert.Equal(t, "My Folder", newMB["name"])

			_, _ = w.Write([]byte(`{
				"sessionState": "s1",
				"methodResponses": [
					["Mailbox/set", {
						"accountId": "acc1",
						"oldState": "m1",
						"newState": "m2",
						"created": {
							"new-mailbox": {
								"id": "mb-new-123",
								"name": "My Folder",
								"sortOrder": 10,
								"totalEmails": 0,
								"unreadEmails": 0,
								"totalThreads": 0,
								"unreadThreads": 0
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
		_, _ = w.Write([]byte(sessionResponse(apiServer.URL)))
	}))
	defer sessionServer.Close()

	client := NewClient(sessionServer.URL, "test-token")
	ctx := context.Background()

	mailbox, err := client.Mailbox().Create(ctx, "My Folder", "")
	require.NoError(t, err)
	require.NotNil(t, mailbox)

	assert.Equal(t, "mb-new-123", mailbox.ID)
	assert.Equal(t, "My Folder", mailbox.Name)
}

func TestMailboxService_CreateWithParent(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)

		w.Header().Set("Content-Type", "application/json")

		args := firstCall[1].(map[string]any)
		create := args["create"].(map[string]any)
		newMB := create["new-mailbox"].(map[string]any)

		assert.Equal(t, "Sub Folder", newMB["name"])
		assert.Equal(t, "parent-mb-id", newMB["parentId"])

		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/set", {
					"accountId": "acc1",
					"oldState": "m1",
					"newState": "m2",
					"created": {
						"new-mailbox": {
							"id": "mb-sub-456",
							"name": "Sub Folder",
							"parentId": "parent-mb-id",
							"sortOrder": 10,
							"totalEmails": 0,
							"unreadEmails": 0,
							"totalThreads": 0,
							"unreadThreads": 0
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

	mailbox, err := client.Mailbox().Create(ctx, "Sub Folder", "parent-mb-id")
	require.NoError(t, err)
	require.NotNil(t, mailbox)

	assert.Equal(t, "mb-sub-456", mailbox.ID)
	assert.Equal(t, "Sub Folder", mailbox.Name)
	assert.Equal(t, "parent-mb-id", mailbox.ParentID)
}

func TestMailboxService_Create_Error(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/set", {
					"accountId": "acc1",
					"oldState": "m1",
					"newState": "m1",
					"created": {},
					"notCreated": {
						"new-mailbox": {
							"type": "invalidProperties",
							"description": "name already exists"
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

	mailbox, err := client.Mailbox().Create(ctx, "Inbox", "")
	assert.Error(t, err)
	assert.Nil(t, mailbox)
	assert.Contains(t, err.Error(), "name already exists")
}

func TestMailboxService_Rename(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)

		w.Header().Set("Content-Type", "application/json")

		args := firstCall[1].(map[string]any)
		update := args["update"].(map[string]any)
		mbUpdate := update["mb-123"].(map[string]any)

		assert.Equal(t, "New Name", mbUpdate["name"])

		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/set", {
					"accountId": "acc1",
					"oldState": "m1",
					"newState": "m2",
					"updated": {"mb-123": null}
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

	err := client.Mailbox().Rename(ctx, "mb-123", "New Name")
	require.NoError(t, err)
}

func TestMailboxService_Rename_NotFound(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/set", {
					"accountId": "acc1",
					"oldState": "m1",
					"newState": "m1",
					"updated": {},
					"notUpdated": {
						"mb-nonexistent": {
							"type": "notFound",
							"description": "Mailbox not found"
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

	err := client.Mailbox().Rename(ctx, "mb-nonexistent", "New Name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Mailbox not found")
}

func TestMailboxService_Delete(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)

		w.Header().Set("Content-Type", "application/json")

		args := firstCall[1].(map[string]any)
		destroy := args["destroy"].([]any)

		assert.Contains(t, destroy, "mb-to-delete")

		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/set", {
					"accountId": "acc1",
					"oldState": "m1",
					"newState": "m2",
					"destroyed": ["mb-to-delete"]
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

	err := client.Mailbox().Delete(ctx, "mb-to-delete")
	require.NoError(t, err)
}

func TestMailboxService_Delete_Error(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/set", {
					"accountId": "acc1",
					"oldState": "m1",
					"newState": "m1",
					"destroyed": [],
					"notDestroyed": {
						"mb-has-mail": {
							"type": "mailboxHasEmail",
							"description": "Mailbox still has messages"
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

	err := client.Mailbox().Delete(ctx, "mb-has-mail")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Mailbox still has messages")
}

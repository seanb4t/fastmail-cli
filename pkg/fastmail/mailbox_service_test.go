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

func TestMailboxService_List(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)

		assert.Equal(t, "Mailbox/get", methodName)

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
							"totalEmails": 150,
							"unreadEmails": 5,
							"totalThreads": 120,
							"unreadThreads": 3
						},
						{
							"id": "mb-sent",
							"name": "Sent",
							"role": "sent",
							"totalEmails": 42,
							"unreadEmails": 0,
							"totalThreads": 30,
							"unreadThreads": 0
						},
						{
							"id": "mb-custom",
							"name": "Projects",
							"parentId": "mb-inbox",
							"totalEmails": 10,
							"unreadEmails": 2,
							"totalThreads": 8,
							"unreadThreads": 1
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
	assert.Equal(t, RoleInbox, mailboxes[0].Role)
	assert.Equal(t, uint64(150), mailboxes[0].TotalEmails)
	assert.Equal(t, uint64(5), mailboxes[0].UnreadEmails)
	assert.Equal(t, uint64(120), mailboxes[0].TotalThreads)
	assert.Equal(t, uint64(3), mailboxes[0].UnreadThreads)
	assert.True(t, mailboxes[0].IsStandard())

	assert.Equal(t, "mb-sent", mailboxes[1].ID)
	assert.Equal(t, RoleSent, mailboxes[1].Role)
	assert.False(t, mailboxes[1].HasUnread())

	assert.Equal(t, "mb-custom", mailboxes[2].ID)
	assert.Equal(t, "Projects", mailboxes[2].Name)
	assert.Equal(t, MailboxRole(""), mailboxes[2].Role)
	assert.Equal(t, "mb-inbox", mailboxes[2].ParentID)
	assert.False(t, mailboxes[2].IsStandard())
}

func TestMailboxService_Create(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "Mailbox/set", methodName)

		// Verify create payload
		create := args["create"].(map[string]any)
		mb1 := create["mb1"].(map[string]any)
		assert.Equal(t, "New Folder", mb1["name"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/set", {
					"accountId": "acc1",
					"oldState": "m1",
					"newState": "m2",
					"created": {
						"mb1": {
							"id": "mb-new-123",
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

	mb, err := client.Mailbox().Create(ctx, "New Folder", "")
	require.NoError(t, err)
	require.NotNil(t, mb)

	assert.Equal(t, "mb-new-123", mb.ID)
	assert.Equal(t, "New Folder", mb.Name)
}

func TestMailboxService_Create_WithParent(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		args := firstCall[1].(map[string]any)

		// Verify parentId is included
		create := args["create"].(map[string]any)
		mb1 := create["mb1"].(map[string]any)
		assert.Equal(t, "Subfolder", mb1["name"])
		assert.Equal(t, "mb-parent", mb1["parentId"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/set", {
					"accountId": "acc1",
					"oldState": "m1",
					"newState": "m2",
					"created": {
						"mb1": {
							"id": "mb-sub-456",
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

	mb, err := client.Mailbox().Create(ctx, "Subfolder", "mb-parent")
	require.NoError(t, err)
	require.NotNil(t, mb)

	assert.Equal(t, "mb-sub-456", mb.ID)
	assert.Equal(t, "Subfolder", mb.Name)
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
					"notCreated": {
						"mb1": {
							"type": "invalidProperties",
							"description": "Name already exists"
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

	mb, err := client.Mailbox().Create(ctx, "Duplicate", "")
	assert.Error(t, err)
	assert.Nil(t, mb)
	assert.Contains(t, err.Error(), "invalidProperties")
}

func TestMailboxService_Rename(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "Mailbox/set", methodName)

		// Verify update payload
		update := args["update"].(map[string]any)
		mbUpdate := update["mb-to-rename"].(map[string]any)
		assert.Equal(t, "Renamed Folder", mbUpdate["name"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/set", {
					"accountId": "acc1",
					"oldState": "m1",
					"newState": "m2",
					"updated": {"mb-to-rename": null}
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

	err := client.Mailbox().Rename(ctx, "mb-to-rename", "Renamed Folder")
	require.NoError(t, err)
}

func TestMailboxService_Rename_Error(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessionState": "s1",
			"methodResponses": [
				["Mailbox/set", {
					"accountId": "acc1",
					"oldState": "m1",
					"newState": "m1",
					"notUpdated": {
						"mb-to-rename": {
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

	err := client.Mailbox().Rename(ctx, "mb-to-rename", "New Name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notFound")
}

func TestMailboxService_Delete(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		methodCalls := req["methodCalls"].([]any)
		firstCall := methodCalls[0].([]any)
		methodName := firstCall[0].(string)
		args := firstCall[1].(map[string]any)

		assert.Equal(t, "Mailbox/set", methodName)

		// Verify destroy payload
		destroy := args["destroy"].([]any)
		assert.Contains(t, destroy, "mb-to-delete")

		w.Header().Set("Content-Type", "application/json")
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
					"notDestroyed": {
						"mb-to-delete": {
							"type": "mailboxHasChild",
							"description": "Mailbox has child mailboxes"
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

	err := client.Mailbox().Delete(ctx, "mb-to-delete")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mailboxHasChild")
}

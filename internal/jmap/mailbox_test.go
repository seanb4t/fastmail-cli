package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMailboxGetArgs_AllMailboxes(t *testing.T) {
	args := NewMailboxGet("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	// nil ids means get all mailboxes
	assert.Nil(t, args["ids"])
}

func TestMailboxGetArgs_SpecificIDs(t *testing.T) {
	args := NewMailboxGet("account-1").
		IDs("mb-1", "mb-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"mb-1", "mb-2"}, args["ids"])
}

func TestMailboxGetArgs_WithProperties(t *testing.T) {
	args := NewMailboxGet("account-1").
		Properties("id", "name", "role", "totalEmails").
		Build()

	assert.Equal(t, []string{"id", "name", "role", "totalEmails"}, args["properties"])
}

func TestMailboxGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"state": "mb-state-123",
		"list": [
			{
				"id": "inbox-id",
				"name": "Inbox",
				"role": "inbox",
				"parentId": null,
				"sortOrder": 1,
				"totalEmails": 150,
				"unreadEmails": 10,
				"totalThreads": 100,
				"unreadThreads": 5
			},
			{
				"id": "sent-id",
				"name": "Sent",
				"role": "sent",
				"parentId": null,
				"sortOrder": 2,
				"totalEmails": 50,
				"unreadEmails": 0
			}
		],
		"notFound": []
	}`

	var resp MailboxGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "mb-state-123", resp.State)
	require.Len(t, resp.List, 2)

	inbox := resp.List[0]
	assert.Equal(t, "inbox-id", inbox.ID)
	assert.Equal(t, "Inbox", inbox.Name)
	assert.Equal(t, "inbox", inbox.Role)
	assert.Equal(t, uint64(150), inbox.TotalEmails)
	assert.Equal(t, uint64(10), inbox.UnreadEmails)

	sent := resp.List[1]
	assert.Equal(t, "sent-id", sent.ID)
	assert.Equal(t, "Sent", sent.Name)
	assert.Equal(t, "sent", sent.Role)
}

func TestMailboxSetBuilder_Create(t *testing.T) {
	args := NewMailboxSet("account-1").
		Create("new-mb", map[string]any{
			"name": "My Folder",
		}).
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	create, ok := args["create"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, create, "new-mb")
	assert.Equal(t, "My Folder", create["new-mb"]["name"])
}

func TestMailboxSetBuilder_CreateWithParent(t *testing.T) {
	args := NewMailboxSet("account-1").
		Create("new-mb", map[string]any{
			"name":     "Sub Folder",
			"parentId": "parent-mb-id",
		}).
		Build()

	create := args["create"].(map[string]map[string]any)
	assert.Equal(t, "Sub Folder", create["new-mb"]["name"])
	assert.Equal(t, "parent-mb-id", create["new-mb"]["parentId"])
}

func TestMailboxSetBuilder_Update(t *testing.T) {
	args := NewMailboxSet("account-1").
		Update("mb-123", map[string]any{
			"name": "Renamed Folder",
		}).
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, update, "mb-123")
	assert.Equal(t, "Renamed Folder", update["mb-123"]["name"])
}

func TestMailboxSetBuilder_Destroy(t *testing.T) {
	args := NewMailboxSet("account-1").
		Destroy("mb-1", "mb-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"mb-1", "mb-2"}, args["destroy"])
}

func TestMailboxSetBuilder_Combined(t *testing.T) {
	args := NewMailboxSet("account-1").
		Create("new-1", map[string]any{"name": "New"}).
		Update("mb-1", map[string]any{"name": "Updated"}).
		Destroy("mb-old").
		Build()

	create := args["create"].(map[string]map[string]any)
	assert.Contains(t, create, "new-1")

	update := args["update"].(map[string]map[string]any)
	assert.Contains(t, update, "mb-1")

	assert.Equal(t, []string{"mb-old"}, args["destroy"])
}

func TestMailboxSetBuilder_EmptyBuild(t *testing.T) {
	args := NewMailboxSet("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Nil(t, args["create"])
	assert.Nil(t, args["update"])
	assert.Nil(t, args["destroy"])
}

func TestMailboxSetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "mb-old",
		"newState": "mb-new",
		"created": {
			"new-1": {
				"id": "mb-created-id",
				"name": "Created Mailbox",
				"sortOrder": 10,
				"totalEmails": 0,
				"unreadEmails": 0,
				"totalThreads": 0,
				"unreadThreads": 0
			}
		},
		"updated": {"mb-1": null},
		"destroyed": ["mb-deleted"],
		"notCreated": {},
		"notUpdated": {},
		"notDestroyed": {}
	}`

	var resp MailboxSetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "mb-old", resp.OldState)
	assert.Equal(t, "mb-new", resp.NewState)

	require.Contains(t, resp.Created, "new-1")
	assert.Equal(t, "mb-created-id", resp.Created["new-1"].ID)
	assert.Equal(t, "Created Mailbox", resp.Created["new-1"].Name)

	assert.Equal(t, []string{"mb-deleted"}, resp.Destroyed)
}

func TestMailboxSetResponse_WithErrors(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "mb-old",
		"newState": "mb-new",
		"created": {},
		"updated": {},
		"destroyed": [],
		"notCreated": {
			"new-1": {
				"type": "invalidProperties",
				"description": "name is required"
			}
		},
		"notUpdated": {
			"mb-1": {
				"type": "notFound",
				"description": "Mailbox not found"
			}
		},
		"notDestroyed": {
			"mb-2": {
				"type": "mailboxHasEmail",
				"description": "Mailbox still has messages"
			}
		}
	}`

	var resp MailboxSetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	require.Contains(t, resp.NotCreated, "new-1")
	assert.Equal(t, "invalidProperties", resp.NotCreated["new-1"].Type)

	require.Contains(t, resp.NotUpdated, "mb-1")
	assert.Equal(t, "notFound", resp.NotUpdated["mb-1"].Type)

	require.Contains(t, resp.NotDestroyed, "mb-2")
	assert.Equal(t, "mailboxHasEmail", resp.NotDestroyed["mb-2"].Type)
}

func TestMailbox_SpecialRoles(t *testing.T) {
	mailboxes := []Mailbox{
		{ID: "mb-1", Role: "inbox"},
		{ID: "mb-2", Role: "sent"},
		{ID: "mb-3", Role: "drafts"},
		{ID: "mb-4", Role: "trash"},
		{ID: "mb-5", Role: "archive"},
		{ID: "mb-6", Role: ""}, // custom folder
	}

	// Verify standard roles are strings
	assert.Equal(t, "inbox", mailboxes[0].Role)
	assert.Equal(t, "sent", mailboxes[1].Role)
	assert.Equal(t, "drafts", mailboxes[2].Role)
	assert.Equal(t, "trash", mailboxes[3].Role)
	assert.Equal(t, "archive", mailboxes[4].Role)
	assert.Equal(t, "", mailboxes[5].Role)
}

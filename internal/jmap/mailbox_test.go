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

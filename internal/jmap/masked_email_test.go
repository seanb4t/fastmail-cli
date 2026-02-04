package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaskedEmailGetArgs_WithIDs(t *testing.T) {
	args := NewMaskedEmailGet("account-1").
		IDs("masked-1", "masked-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"masked-1", "masked-2"}, args["ids"])
}

func TestMaskedEmailGetArgs_WithProperties(t *testing.T) {
	args := NewMaskedEmailGet("account-1").
		IDs("masked-1").
		Properties("id", "email", "forDomain", "state").
		Build()

	assert.Equal(t, []string{"id", "email", "forDomain", "state"}, args["properties"])
}

func TestMaskedEmailGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"state": "s456",
		"list": [
			{
				"id": "masked-1",
				"email": "abc123@fastmail.com",
				"state": "enabled",
				"forDomain": "example.com",
				"description": "Test service",
				"createdAt": "2024-01-15T10:30:00Z",
				"lastMessageAt": "2024-01-20T14:00:00Z"
			}
		],
		"notFound": ["masked-missing"]
	}`

	var resp MaskedEmailGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "s456", resp.State)
	require.Len(t, resp.List, 1)
	assert.Equal(t, "masked-1", resp.List[0].ID)
	assert.Equal(t, "abc123@fastmail.com", resp.List[0].Email)
	assert.Equal(t, MaskedEmailStateEnabled, resp.List[0].State)
	assert.Equal(t, "example.com", resp.List[0].ForDomain)
	assert.Equal(t, "Test service", resp.List[0].Description)
	assert.Equal(t, []string{"masked-missing"}, resp.NotFound)
}

func TestMaskedEmailSetArgs_Create(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Create("new-1", &MaskedEmail{
			ForDomain:   "example.com",
			Description: "Test alias",
			State:       MaskedEmailStateEnabled,
		}).
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	create, ok := args["create"].(map[string]map[string]any)
	require.True(t, ok, "create should be a map")
	require.Contains(t, create, "new-1")

	newEmail := create["new-1"]
	assert.Equal(t, "example.com", newEmail["forDomain"])
	assert.Equal(t, "Test alias", newEmail["description"])
	assert.Equal(t, MaskedEmailStateEnabled, newEmail["state"])
}

func TestMaskedEmailSetArgs_Update(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Update("masked-1", map[string]any{
			"state":       MaskedEmailStateDisabled,
			"description": "Updated description",
		}).
		Build()

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok, "update should be a map")
	require.Contains(t, update, "masked-1")

	changes := update["masked-1"]
	assert.Equal(t, MaskedEmailStateDisabled, changes["state"])
	assert.Equal(t, "Updated description", changes["description"])
}

func TestMaskedEmailSetArgs_Destroy(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Destroy("masked-1", "masked-2").
		Build()

	destroy, ok := args["destroy"].([]string)
	require.True(t, ok, "destroy should be a string slice")
	assert.Equal(t, []string{"masked-1", "masked-2"}, destroy)
}

func TestMaskedEmailSetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "old-state",
		"newState": "new-state",
		"created": {
			"new-1": {
				"id": "masked-created",
				"email": "xyz789@fastmail.com",
				"state": "enabled",
				"forDomain": "example.com"
			}
		},
		"updated": {
			"masked-1": null
		},
		"destroyed": ["masked-deleted"],
		"notCreated": {
			"failed-1": {
				"type": "invalidProperties",
				"description": "Invalid domain"
			}
		}
	}`

	var resp MaskedEmailSetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "old-state", resp.OldState)
	assert.Equal(t, "new-state", resp.NewState)

	require.Contains(t, resp.Created, "new-1")
	created := resp.Created["new-1"]
	assert.Equal(t, "masked-created", created.ID)
	assert.Equal(t, "xyz789@fastmail.com", created.Email)

	assert.Equal(t, []string{"masked-deleted"}, resp.Destroyed)

	require.Contains(t, resp.NotCreated, "failed-1")
	notCreated := resp.NotCreated["failed-1"]
	assert.Equal(t, "invalidProperties", notCreated.Type)
	assert.Equal(t, "Invalid domain", notCreated.Description)
}

func TestMaskedEmailChangesArgs(t *testing.T) {
	args := NewMaskedEmailChanges("account-1", "state-abc").
		MaxChanges(50).
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, "state-abc", args["sinceState"])
	assert.Equal(t, uint64(50), args["maxChanges"])
}

func TestMaskedEmailChangesResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "old-state",
		"newState": "new-state",
		"hasMoreChanges": true,
		"created": ["masked-new"],
		"updated": ["masked-modified"],
		"destroyed": ["masked-deleted"]
	}`

	var resp MaskedEmailChangesResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "old-state", resp.OldState)
	assert.Equal(t, "new-state", resp.NewState)
	assert.True(t, resp.HasMoreChanges)
	assert.Equal(t, []string{"masked-new"}, resp.Created)
	assert.Equal(t, []string{"masked-modified"}, resp.Updated)
	assert.Equal(t, []string{"masked-deleted"}, resp.Destroyed)
}

func TestMaskedEmailState_Values(t *testing.T) {
	assert.Equal(t, MaskedEmailState("enabled"), MaskedEmailStateEnabled)
	assert.Equal(t, MaskedEmailState("disabled"), MaskedEmailStateDisabled)
	assert.Equal(t, MaskedEmailState("pending"), MaskedEmailStatePending)
	assert.Equal(t, MaskedEmailState("deleted"), MaskedEmailStateDeleted)
}

func TestMaskedEmailCapability(t *testing.T) {
	assert.Equal(t, "https://www.fastmail.com/dev/maskedemail", MaskedEmailCapability)
}

func TestMaskedEmail_JSONRoundTrip(t *testing.T) {
	original := MaskedEmail{
		ID:            "masked-123",
		Email:         "alias@fastmail.com",
		State:         MaskedEmailStateEnabled,
		ForDomain:     "github.com",
		Description:   "GitHub account",
		URL:           "https://github.com/signup",
		CreatedBy:     "user",
		CreatedAt:     "2024-01-15T10:00:00Z",
		LastMessageAt: "2024-01-20T15:30:00Z",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded MaskedEmail
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}

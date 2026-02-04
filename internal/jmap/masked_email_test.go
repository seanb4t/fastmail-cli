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
		Properties("id", "email", "state", "forDomain").
		Build()

	assert.Equal(t, []string{"id", "email", "state", "forDomain"}, args["properties"])
}

func TestMaskedEmailGetArgs_Empty(t *testing.T) {
	args := NewMaskedEmailGet("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Nil(t, args["ids"])
	assert.Nil(t, args["properties"])
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
				"description": "Test masked email",
				"createdAt": "2024-01-15T10:30:00Z",
				"lastMessageAt": "2024-01-16T12:00:00Z"
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
	assert.Equal(t, "Test masked email", resp.List[0].Description)
	assert.Equal(t, []string{"masked-missing"}, resp.NotFound)
}

func TestMaskedEmailSetArgs_Create(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Create("new-1", map[string]any{
			"state":       "enabled",
			"forDomain":   "example.com",
			"description": "New masked email",
		}).
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	create, ok := args["create"].(map[string]map[string]any)
	require.True(t, ok, "create should be a map")
	require.Contains(t, create, "new-1")
	assert.Equal(t, "enabled", create["new-1"]["state"])
	assert.Equal(t, "example.com", create["new-1"]["forDomain"])
}

func TestMaskedEmailSetArgs_Update(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Update("masked-1", map[string]any{
			"state": "disabled",
		}).
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok, "update should be a map")
	require.Contains(t, update, "masked-1")
	assert.Equal(t, "disabled", update["masked-1"]["state"])
}

func TestMaskedEmailSetArgs_Destroy(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Destroy("masked-1", "masked-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"masked-1", "masked-2"}, args["destroy"])
}

func TestMaskedEmailSetArgs_Combined(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Create("new-1", map[string]any{
			"forDomain": "newsite.com",
		}).
		Update("masked-1", map[string]any{
			"state": "disabled",
		}).
		Destroy("masked-old").
		Build()

	assert.NotNil(t, args["create"])
	assert.NotNil(t, args["update"])
	assert.NotNil(t, args["destroy"])
}

func TestMaskedEmailSetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "old-state",
		"newState": "new-state",
		"created": {
			"new-1": {
				"id": "masked-created",
				"email": "new123@fastmail.com",
				"state": "enabled"
			}
		},
		"updated": {
			"masked-1": null
		},
		"destroyed": ["masked-deleted"],
		"notCreated": {},
		"notUpdated": {},
		"notDestroyed": {}
	}`

	var resp MaskedEmailSetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "old-state", resp.OldState)
	assert.Equal(t, "new-state", resp.NewState)
	require.Contains(t, resp.Created, "new-1")
	assert.Equal(t, "masked-created", resp.Created["new-1"].ID)
	assert.Equal(t, "new123@fastmail.com", resp.Created["new-1"].Email)
	assert.Equal(t, []string{"masked-deleted"}, resp.Destroyed)
}

func TestMaskedEmailChangesArgs(t *testing.T) {
	args := NewMaskedEmailChanges("account-1", "state-abc").
		MaxChanges(100).
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, "state-abc", args["sinceState"])
	assert.Equal(t, uint64(100), args["maxChanges"])
}

func TestMaskedEmailChangesArgs_NoMaxChanges(t *testing.T) {
	args := NewMaskedEmailChanges("account-1", "state-abc").Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, "state-abc", args["sinceState"])
	assert.Nil(t, args["maxChanges"])
}

func TestMaskedEmailChangesResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "old-state",
		"newState": "new-state",
		"hasMoreChanges": false,
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
	assert.False(t, resp.HasMoreChanges)
	assert.Equal(t, []string{"masked-new"}, resp.Created)
	assert.Equal(t, []string{"masked-modified"}, resp.Updated)
	assert.Equal(t, []string{"masked-deleted"}, resp.Destroyed)
}

func TestMaskedEmailState_Values(t *testing.T) {
	assert.Equal(t, MaskedEmailState("enabled"), MaskedEmailStateEnabled)
	assert.Equal(t, MaskedEmailState("disabled"), MaskedEmailStateDisabled)
	assert.Equal(t, MaskedEmailState("deleted"), MaskedEmailStateDeleted)
	assert.Equal(t, MaskedEmailState("pending"), MaskedEmailStatePending)
}

func TestMaskedEmail_JSONRoundtrip(t *testing.T) {
	original := MaskedEmail{
		ID:            "masked-1",
		Email:         "test123@fastmail.com",
		State:         MaskedEmailStateEnabled,
		ForDomain:     "example.com",
		Description:   "Test description",
		URL:           "https://example.com/signup",
		CreatedBy:     "user@example.com",
		CreatedAt:     "2024-01-15T10:30:00Z",
		LastMessageAt: "2024-01-16T12:00:00Z",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded MaskedEmail
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}

func TestCapMaskedEmail(t *testing.T) {
	assert.Equal(t, "https://www.fastmail.com/dev/maskedemail", CapMaskedEmail)
}

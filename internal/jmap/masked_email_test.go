package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaskedEmailCapability(t *testing.T) {
	assert.Equal(t, "https://www.fastmail.com/dev/maskedemail", MaskedEmailCapability)
}

func TestMaskedEmailState_Values(t *testing.T) {
	assert.Equal(t, MaskedEmailState("enabled"), MaskedEmailStateEnabled)
	assert.Equal(t, MaskedEmailState("disabled"), MaskedEmailStateDisabled)
	assert.Equal(t, MaskedEmailState("pending"), MaskedEmailStatePending)
	assert.Equal(t, MaskedEmailState("deleted"), MaskedEmailStateDeleted)
}

func TestMaskedEmail_JSON(t *testing.T) {
	email := MaskedEmail{
		ID:            "masked-1",
		Email:         "abc123@mydomain.com",
		State:         MaskedEmailStateEnabled,
		ForDomain:     "example.com",
		Description:   "Example Service",
		URL:           "https://example.com",
		CreatedBy:     "maskedemail-cli",
		CreatedAt:     "2024-01-15T10:30:00Z",
		LastMessageAt: "2024-01-16T08:00:00Z",
	}

	data, err := json.Marshal(email)
	require.NoError(t, err)

	var parsed MaskedEmail
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, email.ID, parsed.ID)
	assert.Equal(t, email.Email, parsed.Email)
	assert.Equal(t, email.State, parsed.State)
	assert.Equal(t, email.ForDomain, parsed.ForDomain)
	assert.Equal(t, email.Description, parsed.Description)
	assert.Equal(t, email.URL, parsed.URL)
	assert.Equal(t, email.CreatedBy, parsed.CreatedBy)
	assert.Equal(t, email.CreatedAt, parsed.CreatedAt)
	assert.Equal(t, email.LastMessageAt, parsed.LastMessageAt)
}

func TestMaskedEmailGetBuilder_Basic(t *testing.T) {
	args := NewMaskedEmailGet("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Nil(t, args["ids"])
	assert.Nil(t, args["properties"])
}

func TestMaskedEmailGetBuilder_WithIDs(t *testing.T) {
	args := NewMaskedEmailGet("account-1").
		IDs("masked-1", "masked-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"masked-1", "masked-2"}, args["ids"])
}

func TestMaskedEmailGetBuilder_WithProperties(t *testing.T) {
	args := NewMaskedEmailGet("account-1").
		IDs("masked-1").
		Properties("id", "email", "state", "forDomain").
		Build()

	assert.Equal(t, []string{"id", "email", "state", "forDomain"}, args["properties"])
}

func TestMaskedEmailGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"state": "s789",
		"list": [
			{
				"id": "masked-1",
				"email": "abc123@mydomain.com",
				"state": "enabled",
				"forDomain": "example.com",
				"description": "Example Service",
				"createdAt": "2024-01-15T10:30:00Z"
			}
		],
		"notFound": ["masked-missing"]
	}`

	var resp MaskedEmailGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "s789", resp.State)
	require.Len(t, resp.List, 1)
	assert.Equal(t, "masked-1", resp.List[0].ID)
	assert.Equal(t, "abc123@mydomain.com", resp.List[0].Email)
	assert.Equal(t, MaskedEmailStateEnabled, resp.List[0].State)
	assert.Equal(t, "example.com", resp.List[0].ForDomain)
	assert.Equal(t, []string{"masked-missing"}, resp.NotFound)
}

func TestMaskedEmailSetBuilder_Create(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Create("new-1", "example.com", "Example Service", MaskedEmailStateEnabled).
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	create, ok := args["create"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, create, "new-1")

	newEmail := create["new-1"]
	assert.Equal(t, "example.com", newEmail["forDomain"])
	assert.Equal(t, "Example Service", newEmail["description"])
	assert.Equal(t, MaskedEmailStateEnabled, newEmail["state"])
}

func TestMaskedEmailSetBuilder_Update(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Update("masked-1", map[string]any{
			"description": "Updated Description",
		}).
		Build()

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, update, "masked-1")
	assert.Equal(t, "Updated Description", update["masked-1"]["description"])
}

func TestMaskedEmailSetBuilder_Enable(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Enable("masked-1").
		Build()

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, update, "masked-1")
	assert.Equal(t, MaskedEmailStateEnabled, update["masked-1"]["state"])
}

func TestMaskedEmailSetBuilder_Disable(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Disable("masked-1").
		Build()

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, update, "masked-1")
	assert.Equal(t, MaskedEmailStateDisabled, update["masked-1"]["state"])
}

func TestMaskedEmailSetBuilder_Destroy(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Destroy("masked-1", "masked-2").
		Build()

	assert.Equal(t, []string{"masked-1", "masked-2"}, args["destroy"])
}

func TestMaskedEmailSetBuilder_Combined(t *testing.T) {
	args := NewMaskedEmailSet("account-1").
		Create("new-1", "newdomain.com", "New Service", MaskedEmailStatePending).
		Enable("masked-1").
		Destroy("masked-old").
		Build()

	// Verify create
	create := args["create"].(map[string]map[string]any)
	assert.Contains(t, create, "new-1")

	// Verify update (from Enable)
	update := args["update"].(map[string]map[string]any)
	assert.Contains(t, update, "masked-1")

	// Verify destroy
	assert.Equal(t, []string{"masked-old"}, args["destroy"])
}

func TestMaskedEmailSetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "old-state",
		"newState": "new-state",
		"created": {
			"new-1": {
				"id": "masked-new",
				"email": "xyz789@mydomain.com",
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
	assert.Equal(t, "masked-new", resp.Created["new-1"].ID)
	assert.Equal(t, "xyz789@mydomain.com", resp.Created["new-1"].Email)

	assert.Equal(t, []string{"masked-deleted"}, resp.Destroyed)
}

func TestMaskedEmailSetResponse_WithErrors(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "old-state",
		"newState": "new-state",
		"created": {},
		"updated": {},
		"destroyed": [],
		"notCreated": {
			"new-1": {
				"type": "forbidden",
				"description": "Masked email feature not enabled"
			}
		},
		"notUpdated": {},
		"notDestroyed": {}
	}`

	var resp MaskedEmailSetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	require.Contains(t, resp.NotCreated, "new-1")
	assert.Equal(t, "forbidden", resp.NotCreated["new-1"].Type)
	assert.Equal(t, "Masked email feature not enabled", resp.NotCreated["new-1"].Description)
}

func TestMaskedEmailChangesBuilder_Basic(t *testing.T) {
	args := NewMaskedEmailChanges("account-1", "state-abc").Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, "state-abc", args["sinceState"])
	assert.Nil(t, args["maxChanges"])
}

func TestMaskedEmailChangesBuilder_WithMaxChanges(t *testing.T) {
	args := NewMaskedEmailChanges("account-1", "state-abc").
		MaxChanges(50).
		Build()

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

func TestMaskedEmailIntegration_CreateAndGet(t *testing.T) {
	// Test building a request that creates a masked email and then gets it
	req := NewRequest().
		WithCapabilities(CapCore, MaskedEmailCapability)

	// Create a new masked email
	setCallID := req.Invoke("MaskedEmail/set",
		NewMaskedEmailSet("account-1").
			Create("new-1", "example.com", "Example Service", MaskedEmailStateEnabled).
			Build(),
	)

	// Get the created masked email using result reference
	req.Invoke("MaskedEmail/get", map[string]any{
		"accountId": "account-1",
		"#ids":      ResultReference(setCallID, "MaskedEmail/set", "/created/new-1/id"),
	})

	// Verify JSON structure
	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Verify capabilities include masked email
	using := parsed["using"].([]any)
	assert.Contains(t, using, MaskedEmailCapability)

	methodCalls := parsed["methodCalls"].([]any)
	require.Len(t, methodCalls, 2)

	// Verify MaskedEmail/set
	setCall := methodCalls[0].([]any)
	assert.Equal(t, "MaskedEmail/set", setCall[0])

	// Verify MaskedEmail/get with result reference
	getCall := methodCalls[1].([]any)
	assert.Equal(t, "MaskedEmail/get", getCall[0])
	getArgs := getCall[1].(map[string]any)
	ref := getArgs["#ids"].(map[string]any)
	assert.Equal(t, setCallID, ref["resultOf"])
}

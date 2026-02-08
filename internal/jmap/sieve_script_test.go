package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSieveScriptCapability(t *testing.T) {
	assert.Equal(t, "urn:ietf:params:jmap:sieve", SieveScriptCapability)
}

func TestSieveScript_JSON(t *testing.T) {
	script := SieveScript{
		ID:        "sieve-1",
		Name:      "My Filter",
		BlobID:    "blob-abc",
		Script:    `require "fileinto"; if header :is "from" "test@example.com" { fileinto "Test"; }`,
		IsActive:  true,
		CreatedAt: "2024-01-15T10:30:00Z",
		UpdatedAt: "2024-01-16T08:00:00Z",
	}

	data, err := json.Marshal(script)
	require.NoError(t, err)

	var parsed SieveScript
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, script.ID, parsed.ID)
	assert.Equal(t, script.Name, parsed.Name)
	assert.Equal(t, script.BlobID, parsed.BlobID)
	assert.Equal(t, script.Script, parsed.Script)
	assert.Equal(t, script.IsActive, parsed.IsActive)
	assert.Equal(t, script.CreatedAt, parsed.CreatedAt)
	assert.Equal(t, script.UpdatedAt, parsed.UpdatedAt)
}

func TestSieveScriptGetBuilder_Basic(t *testing.T) {
	args := NewSieveScriptGet("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Nil(t, args["ids"])
	assert.Nil(t, args["properties"])
}

func TestSieveScriptGetBuilder_WithIDs(t *testing.T) {
	args := NewSieveScriptGet("account-1").
		IDs("sieve-1", "sieve-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"sieve-1", "sieve-2"}, args["ids"])
}

func TestSieveScriptGetBuilder_WithProperties(t *testing.T) {
	args := NewSieveScriptGet("account-1").
		IDs("sieve-1").
		Properties("id", "name", "isActive").
		Build()

	assert.Equal(t, []string{"id", "name", "isActive"}, args["properties"])
}

func TestSieveScriptGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"state": "s123",
		"list": [
			{
				"id": "sieve-1",
				"name": "My Filter",
				"blobId": "blob-abc",
				"isActive": true,
				"createdAt": "2024-01-15T10:30:00Z"
			}
		],
		"notFound": ["sieve-missing"]
	}`

	var resp SieveScriptGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "s123", resp.State)
	require.Len(t, resp.List, 1)
	assert.Equal(t, "sieve-1", resp.List[0].ID)
	assert.Equal(t, "My Filter", resp.List[0].Name)
	assert.Equal(t, "blob-abc", resp.List[0].BlobID)
	assert.True(t, resp.List[0].IsActive)
	assert.Equal(t, []string{"sieve-missing"}, resp.NotFound)
}

func TestSieveScriptSetBuilder_Create(t *testing.T) {
	args := NewSieveScriptSet("account-1").
		Create("new-1", "My Filter", `require "fileinto";`).
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	create, ok := args["create"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, create, "new-1")

	newScript := create["new-1"]
	assert.Equal(t, "My Filter", newScript["name"])
	assert.Equal(t, `require "fileinto";`, newScript["script"])
}

func TestSieveScriptSetBuilder_CreateActive(t *testing.T) {
	args := NewSieveScriptSet("account-1").
		CreateActive("new-1", "Active Filter", `require "reject";`).
		Build()

	create := args["create"].(map[string]map[string]any)
	newScript := create["new-1"]
	assert.Equal(t, "Active Filter", newScript["name"])
	assert.Equal(t, true, newScript["isActive"])
}

func TestSieveScriptSetBuilder_Activate(t *testing.T) {
	args := NewSieveScriptSet("account-1").
		Activate("sieve-1").
		Build()

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, update, "sieve-1")
	assert.Equal(t, true, update["sieve-1"]["isActive"])
}

func TestSieveScriptSetBuilder_Deactivate(t *testing.T) {
	args := NewSieveScriptSet("account-1").
		Deactivate("sieve-1").
		Build()

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, update, "sieve-1")
	assert.Equal(t, false, update["sieve-1"]["isActive"])
}

func TestSieveScriptSetBuilder_Destroy(t *testing.T) {
	args := NewSieveScriptSet("account-1").
		Destroy("sieve-1", "sieve-2").
		Build()

	assert.Equal(t, []string{"sieve-1", "sieve-2"}, args["destroy"])
}

func TestSieveScriptSetBuilder_Combined(t *testing.T) {
	args := NewSieveScriptSet("account-1").
		Create("new-1", "New Filter", `require "fileinto";`).
		Activate("sieve-2").
		Destroy("sieve-old").
		Build()

	create := args["create"].(map[string]map[string]any)
	assert.Contains(t, create, "new-1")

	update := args["update"].(map[string]map[string]any)
	assert.Contains(t, update, "sieve-2")

	assert.Equal(t, []string{"sieve-old"}, args["destroy"])
}

func TestSieveScriptSetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "old-state",
		"newState": "new-state",
		"created": {
			"new-1": {
				"id": "sieve-new",
				"name": "My Filter",
				"isActive": false
			}
		},
		"updated": {
			"sieve-1": null
		},
		"destroyed": ["sieve-deleted"],
		"notCreated": {},
		"notUpdated": {},
		"notDestroyed": {}
	}`

	var resp SieveScriptSetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "old-state", resp.OldState)
	assert.Equal(t, "new-state", resp.NewState)

	require.Contains(t, resp.Created, "new-1")
	assert.Equal(t, "sieve-new", resp.Created["new-1"].ID)
	assert.Equal(t, "My Filter", resp.Created["new-1"].Name)

	assert.Equal(t, []string{"sieve-deleted"}, resp.Destroyed)
}

func TestSieveScriptSetResponse_WithErrors(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "old-state",
		"newState": "new-state",
		"created": {},
		"updated": {},
		"destroyed": [],
		"notCreated": {
			"new-1": {
				"type": "invalidScript",
				"description": "Syntax error in sieve script"
			}
		},
		"notUpdated": {},
		"notDestroyed": {}
	}`

	var resp SieveScriptSetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	require.Contains(t, resp.NotCreated, "new-1")
	assert.Equal(t, "invalidScript", resp.NotCreated["new-1"].Type)
	assert.Equal(t, "Syntax error in sieve script", resp.NotCreated["new-1"].Description)
}

func TestSieveScriptValidateBuilder(t *testing.T) {
	args := NewSieveScriptValidate("account-1", `require "fileinto";`).Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, `require "fileinto";`, args["script"])
}

func TestSieveScriptValidateResponse_Valid(t *testing.T) {
	responseJSON := `{
		"accountId": "A1"
	}`

	var resp SieveScriptValidateResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Nil(t, resp.Error)
}

func TestSieveScriptValidateResponse_Invalid(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"error": {
			"type": "invalidScript",
			"description": "line 1: unknown command 'badcommand'"
		}
	}`

	var resp SieveScriptValidateResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	require.NotNil(t, resp.Error)
	assert.Equal(t, "invalidScript", resp.Error.Type)
	assert.Contains(t, resp.Error.Description, "badcommand")
}

func TestSieveScriptIntegration_CreateAndGet(t *testing.T) {
	req := NewRequest().
		WithCapabilities(CapCore, SieveScriptCapability)

	setCallID := req.Invoke("SieveScript/set",
		NewSieveScriptSet("account-1").
			Create("new-1", "My Filter", `require "fileinto";`).
			Build(),
	)

	req.Invoke("SieveScript/get", map[string]any{
		"accountId": "account-1",
		"#ids":      ResultReference(setCallID, "SieveScript/set", "/created/new-1/id"),
	})

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	using := parsed["using"].([]any)
	assert.Contains(t, using, SieveScriptCapability)

	methodCalls := parsed["methodCalls"].([]any)
	require.Len(t, methodCalls, 2)

	setCall := methodCalls[0].([]any)
	assert.Equal(t, "SieveScript/set", setCall[0])

	getCall := methodCalls[1].([]any)
	assert.Equal(t, "SieveScript/get", getCall[0])
	getArgs := getCall[1].(map[string]any)
	ref := getArgs["#ids"].(map[string]any)
	assert.Equal(t, setCallID, ref["resultOf"])
}

package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentity_JSON(t *testing.T) {
	identity := Identity{
		ID:    "id-1",
		Name:  "Test User",
		Email: "test@example.com",
		ReplyTo: []EmailAddress{
			{Name: "Reply", Email: "reply@example.com"},
		},
		Bcc: []EmailAddress{
			{Email: "bcc@example.com"},
		},
		TextSignature: "-- \nSent from fastmail-cli",
		HTMLSignature: "<p>Sent from fastmail-cli</p>",
		MayDelete:     true,
	}

	data, err := json.Marshal(identity)
	require.NoError(t, err)

	var parsed Identity
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, identity.ID, parsed.ID)
	assert.Equal(t, identity.Name, parsed.Name)
	assert.Equal(t, identity.Email, parsed.Email)
	require.Len(t, parsed.ReplyTo, 1)
	assert.Equal(t, "Reply", parsed.ReplyTo[0].Name)
	assert.Equal(t, "reply@example.com", parsed.ReplyTo[0].Email)
	require.Len(t, parsed.Bcc, 1)
	assert.Equal(t, "bcc@example.com", parsed.Bcc[0].Email)
	assert.Equal(t, identity.TextSignature, parsed.TextSignature)
	assert.Equal(t, identity.HTMLSignature, parsed.HTMLSignature)
	assert.True(t, parsed.MayDelete)
}

func TestIdentity_JSON_OmitEmpty(t *testing.T) {
	identity := Identity{
		ID:    "id-1",
		Name:  "Test User",
		Email: "test@example.com",
	}

	data, err := json.Marshal(identity)
	require.NoError(t, err)

	var raw map[string]any
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Contains(t, raw, "id")
	assert.Contains(t, raw, "name")
	assert.Contains(t, raw, "email")
	assert.NotContains(t, raw, "replyTo")
	assert.NotContains(t, raw, "bcc")
	assert.NotContains(t, raw, "textSignature")
	assert.NotContains(t, raw, "htmlSignature")
}

func TestIdentityGetBuilder_Basic(t *testing.T) {
	args := NewIdentityGet("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Nil(t, args["ids"])
}

func TestIdentityGetBuilder_WithIDs(t *testing.T) {
	args := NewIdentityGet("account-1").IDs("id-1", "id-2").Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"id-1", "id-2"}, args["ids"])
}

func TestIdentityGetBuilder_WithProperties(t *testing.T) {
	args := NewIdentityGet("account-1").
		Properties("id", "name", "email").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"id", "name", "email"}, args["properties"])
}

func TestIdentityGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "acc1",
		"state": "i1",
		"list": [
			{
				"id": "id-1",
				"name": "Test User",
				"email": "test@example.com",
				"replyTo": [{"name": "Reply", "email": "reply@example.com"}],
				"bcc": [{"email": "bcc@example.com"}],
				"textSignature": "-- \nSent from CLI",
				"htmlSignature": "<p>Sent from CLI</p>",
				"mayDelete": false
			},
			{
				"id": "id-2",
				"name": "Alt Identity",
				"email": "alt@example.com",
				"mayDelete": true
			}
		],
		"notFound": []
	}`

	var resp IdentityGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "acc1", resp.AccountID)
	assert.Equal(t, "i1", resp.State)
	require.Len(t, resp.List, 2)

	first := resp.List[0]
	assert.Equal(t, "id-1", first.ID)
	assert.Equal(t, "Test User", first.Name)
	assert.Equal(t, "test@example.com", first.Email)
	require.Len(t, first.ReplyTo, 1)
	assert.Equal(t, "reply@example.com", first.ReplyTo[0].Email)
	require.Len(t, first.Bcc, 1)
	assert.Equal(t, "bcc@example.com", first.Bcc[0].Email)
	assert.Equal(t, "-- \nSent from CLI", first.TextSignature)
	assert.False(t, first.MayDelete)

	second := resp.List[1]
	assert.Equal(t, "id-2", second.ID)
	assert.Equal(t, "Alt Identity", second.Name)
	assert.True(t, second.MayDelete)
	assert.Empty(t, second.ReplyTo)
	assert.Empty(t, second.Bcc)
	assert.Empty(t, resp.NotFound)
}

func TestIdentitySetBuilder_Update(t *testing.T) {
	args := NewIdentitySet("account-1").
		Update("id-1", map[string]any{
			"name":          "Updated Name",
			"textSignature": "New signature",
		}).
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok)
	require.Contains(t, update, "id-1")
	assert.Equal(t, "Updated Name", update["id-1"]["name"])
	assert.Equal(t, "New signature", update["id-1"]["textSignature"])
}

func TestIdentitySetBuilder_MultipleUpdates(t *testing.T) {
	args := NewIdentitySet("account-1").
		Update("id-1", map[string]any{"name": "Name 1"}).
		Update("id-2", map[string]any{"name": "Name 2"}).
		Build()

	update, ok := args["update"].(map[string]map[string]any)
	require.True(t, ok)
	assert.Len(t, update, 2)
	assert.Contains(t, update, "id-1")
	assert.Contains(t, update, "id-2")
}

func TestIdentitySetBuilder_EmptyUpdate(t *testing.T) {
	args := NewIdentitySet("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.NotContains(t, args, "update")
}

func TestIdentitySetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "acc1",
		"oldState": "i1",
		"newState": "i2",
		"updated": {"id-1": null},
		"notUpdated": {}
	}`

	var resp IdentitySetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "acc1", resp.AccountID)
	assert.Equal(t, "i1", resp.OldState)
	assert.Equal(t, "i2", resp.NewState)
	assert.Contains(t, resp.Updated, "id-1")
	assert.Empty(t, resp.NotUpdated)
}

func TestIdentitySetResponse_WithError(t *testing.T) {
	responseJSON := `{
		"accountId": "acc1",
		"oldState": "i1",
		"newState": "i1",
		"updated": {},
		"notUpdated": {
			"id-1": {
				"type": "invalidProperties",
				"description": "name is too long"
			}
		}
	}`

	var resp IdentitySetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	require.Contains(t, resp.NotUpdated, "id-1")
	assert.Equal(t, "invalidProperties", resp.NotUpdated["id-1"].Type)
	assert.Equal(t, "name is too long", resp.NotUpdated["id-1"].Description)
}

func TestIdentityIntegration_GetAndSet(t *testing.T) {
	req := NewRequest().
		WithCapabilities(CapCore, CapSubmission)

	getCallID := req.Invoke("Identity/get",
		NewIdentityGet("account-1").Build(),
	)

	req.Invoke("Identity/set",
		NewIdentitySet("account-1").
			Update("id-1", map[string]any{
				"name":          "Updated Name",
				"textSignature": "-- \nNew sig",
			}).
			Build(),
	)

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	using := parsed["using"].([]any)
	assert.Contains(t, using, CapCore)
	assert.Contains(t, using, CapSubmission)

	methodCalls := parsed["methodCalls"].([]any)
	require.Len(t, methodCalls, 2)

	getCall := methodCalls[0].([]any)
	assert.Equal(t, "Identity/get", getCall[0])
	assert.Equal(t, getCallID, getCall[2])

	setCall := methodCalls[1].([]any)
	assert.Equal(t, "Identity/set", setCall[0])
	setArgs := setCall[1].(map[string]any)
	update := setArgs["update"].(map[string]any)
	idUpdate := update["id-1"].(map[string]any)
	assert.Equal(t, "Updated Name", idUpdate["name"])
}

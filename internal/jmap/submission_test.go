package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailSubmissionQueryArgs_UndoStatus(t *testing.T) {
	args := NewEmailSubmissionQuery("account-1").
		UndoStatus("pending").
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	filter, ok := args["filter"].(map[string]any)
	require.True(t, ok, "filter should be a map")
	assert.Equal(t, "pending", filter["undoStatus"])
}

func TestEmailSubmissionQueryArgs_Sort(t *testing.T) {
	args := NewEmailSubmissionQuery("account-1").
		SortBy("sendAt", false). // ascending
		Build()

	sort, ok := args["sort"].([]Comparator)
	require.True(t, ok, "sort should be []Comparator")
	require.Len(t, sort, 1)
	assert.Equal(t, "sendAt", sort[0].Property)
	assert.True(t, sort[0].IsAscending)
}

func TestEmailSubmissionQueryArgs_Limit(t *testing.T) {
	args := NewEmailSubmissionQuery("account-1").
		Limit(25).
		Build()

	assert.Equal(t, uint64(25), args["limit"])
}

func TestEmailSubmissionQueryArgs_Complex(t *testing.T) {
	args := NewEmailSubmissionQuery("account-1").
		UndoStatus("pending").
		CreatedAfter("2024-01-01T00:00:00Z").
		SortBy("sendAt", false).
		Limit(10).
		Build()

	filter, ok := args["filter"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "pending", filter["undoStatus"])
	assert.Equal(t, "2024-01-01T00:00:00Z", filter["createdAfter"])
	assert.Equal(t, uint64(10), args["limit"])
}

func TestEmailSubmissionQueryResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"queryState": "q123",
		"canCalculateChanges": false,
		"position": 0,
		"ids": ["sub-1", "sub-2"],
		"total": 2
	}`

	var resp EmailSubmissionQueryResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, []string{"sub-1", "sub-2"}, resp.IDs)
	assert.Equal(t, uint64(2), resp.Total)
}

func TestEmailSubmissionGetArgs_WithIDs(t *testing.T) {
	args := NewEmailSubmissionGet("account-1").
		IDs("sub-1", "sub-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"sub-1", "sub-2"}, args["ids"])
}

func TestEmailSubmissionGetArgs_WithProperties(t *testing.T) {
	args := NewEmailSubmissionGet("account-1").
		IDs("sub-1").
		Properties("id", "emailId", "sendAt", "undoStatus").
		Build()

	assert.Equal(t, []string{"id", "emailId", "sendAt", "undoStatus"}, args["properties"])
}

func TestEmailSubmissionGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"state": "s1",
		"list": [
			{
				"id": "sub-1",
				"identityId": "id-1",
				"emailId": "email-1",
				"sendAt": "2024-06-15T14:00:00Z",
				"undoStatus": "pending"
			}
		],
		"notFound": []
	}`

	var resp EmailSubmissionGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	require.Len(t, resp.List, 1)
	assert.Equal(t, "sub-1", resp.List[0].ID)
	assert.Equal(t, "email-1", resp.List[0].EmailID)
	assert.Equal(t, "2024-06-15T14:00:00Z", resp.List[0].SendAt)
	assert.Equal(t, "pending", resp.List[0].UndoStatus)
}

func TestEmailSubmissionDestroyArgs(t *testing.T) {
	args := NewEmailSubmissionDestroy("account-1").
		Destroy("sub-1", "sub-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"sub-1", "sub-2"}, args["destroy"])
}

func TestEmailSubmissionSetArgs_WithSendAt(t *testing.T) {
	args := NewEmailSubmissionSet("account-1").
		Create("sub1", map[string]any{
			"identityId": "id-1",
			"emailId":    "#draft",
			"sendAt":     "2024-06-15T14:00:00Z",
		}).
		Build()

	create, ok := args["create"].(map[string]map[string]any)
	require.True(t, ok)
	sub, ok := create["sub1"]
	require.True(t, ok)
	assert.Equal(t, "2024-06-15T14:00:00Z", sub["sendAt"])
}

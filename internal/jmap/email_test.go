package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailQueryArgs_InMailbox(t *testing.T) {
	args := NewEmailQuery("account-1").
		InMailbox("mailbox-id").
		Build()

	assert.Equal(t, "account-1", args["accountId"])

	filter, ok := args["filter"].(map[string]any)
	require.True(t, ok, "filter should be a map")
	assert.Equal(t, "mailbox-id", filter["inMailbox"])
}

func TestEmailQueryArgs_Limit(t *testing.T) {
	args := NewEmailQuery("account-1").
		Limit(50).
		Build()

	assert.Equal(t, uint64(50), args["limit"])
}

func TestEmailQueryArgs_Sort(t *testing.T) {
	args := NewEmailQuery("account-1").
		SortBy("receivedAt", true). // descending
		Build()

	sort, ok := args["sort"].([]Comparator)
	require.True(t, ok, "sort should be []Comparator")
	require.Len(t, sort, 1)
	assert.Equal(t, "receivedAt", sort[0].Property)
	assert.True(t, sort[0].IsAscending == false)
}

func TestEmailQueryArgs_ComplexFilter(t *testing.T) {
	args := NewEmailQuery("account-1").
		InMailbox("inbox-id").
		From("sender@example.com").
		HasKeyword("$seen").
		Build()

	filter, ok := args["filter"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "inbox-id", filter["inMailbox"])
	assert.Equal(t, "sender@example.com", filter["from"])
	assert.Equal(t, "$seen", filter["hasKeyword"])
}

func TestEmailQueryResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"queryState": "q123",
		"canCalculateChanges": true,
		"position": 0,
		"ids": ["email-1", "email-2", "email-3"],
		"total": 100
	}`

	var resp EmailQueryResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "q123", resp.QueryState)
	assert.True(t, resp.CanCalculateChanges)
	assert.Equal(t, uint64(0), resp.Position)
	assert.Equal(t, []string{"email-1", "email-2", "email-3"}, resp.IDs)
	assert.Equal(t, uint64(100), resp.Total)
}

func TestEmailGetArgs_WithIDs(t *testing.T) {
	args := NewEmailGet("account-1").
		IDs("email-1", "email-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"email-1", "email-2"}, args["ids"])
}

func TestEmailGetArgs_WithProperties(t *testing.T) {
	args := NewEmailGet("account-1").
		IDs("email-1").
		Properties("id", "subject", "from", "receivedAt").
		Build()

	assert.Equal(t, []string{"id", "subject", "from", "receivedAt"}, args["properties"])
}

func TestEmailGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"state": "s456",
		"list": [
			{
				"id": "email-1",
				"subject": "Hello World",
				"receivedAt": "2024-01-15T10:30:00Z"
			}
		],
		"notFound": ["email-missing"]
	}`

	var resp EmailGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "s456", resp.State)
	require.Len(t, resp.List, 1)
	assert.Equal(t, "email-1", resp.List[0].ID)
	assert.Equal(t, "Hello World", resp.List[0].Subject)
	assert.Equal(t, []string{"email-missing"}, resp.NotFound)
}

func TestEmailChangesArgs(t *testing.T) {
	args := NewEmailChanges("account-1", "state-abc").
		MaxChanges(100).
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, "state-abc", args["sinceState"])
	assert.Equal(t, uint64(100), args["maxChanges"])
}

func TestEmailChangesResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"oldState": "old-state",
		"newState": "new-state",
		"hasMoreChanges": false,
		"created": ["email-new"],
		"updated": ["email-modified"],
		"destroyed": ["email-deleted"]
	}`

	var resp EmailChangesResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "old-state", resp.OldState)
	assert.Equal(t, "new-state", resp.NewState)
	assert.False(t, resp.HasMoreChanges)
	assert.Equal(t, []string{"email-new"}, resp.Created)
	assert.Equal(t, []string{"email-modified"}, resp.Updated)
	assert.Equal(t, []string{"email-deleted"}, resp.Destroyed)
}

func TestThreadGetArgs_WithIDs(t *testing.T) {
	args := NewThreadGet("account-1").
		IDs("thread-1", "thread-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, []string{"thread-1", "thread-2"}, args["ids"])
}

func TestThreadGetArgs_MinimalBuild(t *testing.T) {
	args := NewThreadGet("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	_, hasIDs := args["ids"]
	assert.False(t, hasIDs, "ids should not be present when none are set")
}

func TestThreadGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"state": "t789",
		"list": [
			{
				"id": "thread-1",
				"emailIds": ["email-a", "email-b", "email-c"]
			}
		],
		"notFound": ["thread-missing"]
	}`

	var resp ThreadGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	assert.Equal(t, "t789", resp.State)
	require.Len(t, resp.List, 1)
	assert.Equal(t, "thread-1", resp.List[0].ID)
	assert.Equal(t, []string{"email-a", "email-b", "email-c"}, resp.List[0].EmailIDs)
	assert.Equal(t, []string{"thread-missing"}, resp.NotFound)
}

func TestThreadGetResponse_DecodeMultipleThreads(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"state": "t100",
		"list": [
			{
				"id": "thread-1",
				"emailIds": ["email-1"]
			},
			{
				"id": "thread-2",
				"emailIds": ["email-2", "email-3"]
			}
		],
		"notFound": []
	}`

	var resp ThreadGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	require.Len(t, resp.List, 2)
	assert.Equal(t, "thread-1", resp.List[0].ID)
	assert.Equal(t, "thread-2", resp.List[1].ID)
	assert.Len(t, resp.List[1].EmailIDs, 2)
}

func TestEmailQueryThenGet_ResultReference(t *testing.T) {
	// Test the pattern: Email/query -> Email/get using result references
	req := NewRequest().
		WithCapabilities(CapCore, CapMail)

	// Query for email IDs in inbox
	queryCallID := req.Invoke("Email/query",
		NewEmailQuery("account-1").
			InMailbox("inbox-id").
			SortBy("receivedAt", true).
			Limit(10).
			Build(),
	)

	// Get emails using result reference from query
	req.Invoke("Email/get", map[string]any{
		"accountId":  "account-1",
		"#ids":       ResultReference(queryCallID, "Email/query", "/ids"),
		"properties": []string{"id", "subject", "from", "receivedAt"},
	})

	// Verify JSON structure
	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	methodCalls := parsed["methodCalls"].([]any)
	require.Len(t, methodCalls, 2)

	// Verify Email/query
	queryCall := methodCalls[0].([]any)
	assert.Equal(t, "Email/query", queryCall[0])
	queryArgs := queryCall[1].(map[string]any)
	assert.Equal(t, "account-1", queryArgs["accountId"])
	assert.NotNil(t, queryArgs["filter"])
	assert.NotNil(t, queryArgs["sort"])

	// Verify Email/get with result reference
	getCall := methodCalls[1].([]any)
	assert.Equal(t, "Email/get", getCall[0])
	getArgs := getCall[1].(map[string]any)
	ref := getArgs["#ids"].(map[string]any)
	assert.Equal(t, queryCallID, ref["resultOf"])
	assert.Equal(t, "Email/query", ref["name"])
	assert.Equal(t, "/ids", ref["path"])
}

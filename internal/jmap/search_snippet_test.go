package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchSnippetGetBuilder_Build(t *testing.T) {
	filter := map[string]any{"text": "meeting notes"}
	args := NewSearchSnippetGet("account-1").
		Filter(filter).
		EmailIDs("email-1", "email-2").
		Build()

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, filter, args["filter"])
	assert.Equal(t, []string{"email-1", "email-2"}, args["emailIds"])
}

func TestSearchSnippetGetBuilder_FilterFromQuery(t *testing.T) {
	queryBuilder := NewEmailQuery("account-1").
		InMailbox("inbox-id").
		Limit(10)

	queryArgs := queryBuilder.Build()

	snippetBuilder := NewSearchSnippetGet("account-1").
		FilterFromQuery(queryBuilder).
		EmailIDs("email-1")

	args := snippetBuilder.Build()
	assert.Equal(t, "account-1", args["accountId"])

	// The filter should match the query filter exactly
	snippetFilter := args["filter"].(map[string]any)
	queryFilter := queryArgs["filter"].(map[string]any)
	assert.Equal(t, queryFilter, snippetFilter)
}

func TestSearchSnippetGetBuilder_WithBackReference(t *testing.T) {
	// Test building args with a back-reference for emailIds
	filter := map[string]any{"text": "meeting"}
	args := NewSearchSnippetGet("account-1").
		Filter(filter).
		Build()

	// Add back-reference manually (as the caller would do)
	args["#emailIds"] = ResultReference("0", "Email/query", "/ids")

	assert.Equal(t, "account-1", args["accountId"])
	assert.Equal(t, filter, args["filter"])

	ref, ok := args["#emailIds"].(ResultRef)
	require.True(t, ok)
	assert.Equal(t, "0", ref.ResultOf)
	assert.Equal(t, "Email/query", ref.Name)
	assert.Equal(t, "/ids", ref.Path)
}

func TestSearchSnippetGetResponse_Decode(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"list": [
			{
				"emailId": "email-1",
				"subject": "Meeting <mark>notes</mark> from today",
				"preview": "Here are the <mark>notes</mark> from our meeting..."
			},
			{
				"emailId": "email-2",
				"subject": null,
				"preview": "I found some <mark>notes</mark> you might like"
			}
		],
		"notFound": ["email-3"]
	}`

	var resp SearchSnippetGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "A1", resp.AccountID)
	require.Len(t, resp.List, 2)

	assert.Equal(t, "email-1", resp.List[0].EmailID)
	assert.NotNil(t, resp.List[0].Subject)
	assert.Equal(t, "Meeting <mark>notes</mark> from today", *resp.List[0].Subject)
	assert.NotNil(t, resp.List[0].Preview)
	assert.Equal(t, "Here are the <mark>notes</mark> from our meeting...", *resp.List[0].Preview)

	assert.Equal(t, "email-2", resp.List[1].EmailID)
	assert.Nil(t, resp.List[1].Subject)
	assert.NotNil(t, resp.List[1].Preview)
	assert.Equal(t, "I found some <mark>notes</mark> you might like", *resp.List[1].Preview)

	assert.Equal(t, []string{"email-3"}, resp.NotFound)
}

func TestSearchSnippetGetResponse_AllNulls(t *testing.T) {
	responseJSON := `{
		"accountId": "A1",
		"list": [
			{
				"emailId": "email-1",
				"subject": null,
				"preview": null
			}
		],
		"notFound": []
	}`

	var resp SearchSnippetGetResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	require.Len(t, resp.List, 1)
	assert.Equal(t, "email-1", resp.List[0].EmailID)
	assert.Nil(t, resp.List[0].Subject)
	assert.Nil(t, resp.List[0].Preview)
}

func TestSearchSnippetGetBuilder_MinimalBuild(t *testing.T) {
	args := NewSearchSnippetGet("account-1").Build()

	assert.Equal(t, "account-1", args["accountId"])
	_, hasFilter := args["filter"]
	assert.False(t, hasFilter, "filter should not be present when not set")
	_, hasIDs := args["emailIds"]
	assert.False(t, hasIDs, "emailIds should not be present when not set")
}

func TestSearchSnippet_ThreeCallChain(t *testing.T) {
	// Test the full 3-call chain pattern: Email/query + Email/get + SearchSnippet/get
	req := NewRequest().
		WithCapabilities(CapCore, CapMail)

	filter := map[string]any{"text": "meeting"}

	// 1. Email/query
	queryCallID := req.Invoke("Email/query", map[string]any{
		"accountId": "account-1",
		"filter":    filter,
		"sort":      []Comparator{{Property: "receivedAt", IsAscending: false}},
	})

	// 2. Email/get with back-reference to query
	getArgs := NewEmailGet("account-1").
		Properties("id", "threadId", "subject", "preview", "receivedAt", "size", "keywords", "mailboxIds").
		Build()
	getArgs["#ids"] = ResultReference(queryCallID, "Email/query", "/ids")
	req.Invoke("Email/get", getArgs)

	// 3. SearchSnippet/get with back-reference to query
	snippetArgs := NewSearchSnippetGet("account-1").
		Filter(filter).
		Build()
	snippetArgs["#emailIds"] = ResultReference(queryCallID, "Email/query", "/ids")
	req.Invoke("SearchSnippet/get", snippetArgs)

	// Verify JSON structure
	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	methodCalls := parsed["methodCalls"].([]any)
	require.Len(t, methodCalls, 3)

	// Verify Email/query
	queryCall := methodCalls[0].([]any)
	assert.Equal(t, "Email/query", queryCall[0])

	// Verify Email/get with #ids back-reference
	getCall := methodCalls[1].([]any)
	assert.Equal(t, "Email/get", getCall[0])
	getCallArgs := getCall[1].(map[string]any)
	idsRef := getCallArgs["#ids"].(map[string]any)
	assert.Equal(t, queryCallID, idsRef["resultOf"])
	assert.Equal(t, "Email/query", idsRef["name"])
	assert.Equal(t, "/ids", idsRef["path"])

	// Verify SearchSnippet/get with #emailIds back-reference
	snippetCall := methodCalls[2].([]any)
	assert.Equal(t, "SearchSnippet/get", snippetCall[0])
	snippetCallArgs := snippetCall[1].(map[string]any)
	emailIDsRef := snippetCallArgs["#emailIds"].(map[string]any)
	assert.Equal(t, queryCallID, emailIDsRef["resultOf"])
	assert.Equal(t, "Email/query", emailIDsRef["name"])
	assert.Equal(t, "/ids", emailIDsRef["path"])

	// Verify the filter is present in SearchSnippet/get
	snippetFilter := snippetCallArgs["filter"].(map[string]any)
	assert.Equal(t, "meeting", snippetFilter["text"])
}

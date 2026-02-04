package jmap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequest(t *testing.T) {
	req := NewRequest()
	require.NotNil(t, req)
	assert.Empty(t, req.Using)
	assert.Empty(t, req.MethodCalls)
}

func TestRequest_Using(t *testing.T) {
	req := NewRequest().
		WithCapabilities(CapCore, CapMail)

	assert.Equal(t, []string{CapCore, CapMail}, req.Using)
}

func TestRequest_JSON(t *testing.T) {
	req := NewRequest().
		WithCapabilities(CapCore, CapMail)
	req.Invoke("Mailbox/get", map[string]any{
		"accountId": "A1",
		"ids":       nil,
	})

	data, err := json.Marshal(req)
	require.NoError(t, err)

	// Parse back to verify structure
	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Verify "using" array
	using, ok := parsed["using"].([]any)
	require.True(t, ok)
	assert.Equal(t, []any{CapCore, CapMail}, using)

	// Verify "methodCalls" is an array of invocations
	methodCalls, ok := parsed["methodCalls"].([]any)
	require.True(t, ok)
	require.Len(t, methodCalls, 1)

	// Each invocation is [methodName, args, callId]
	invocation, ok := methodCalls[0].([]any)
	require.True(t, ok)
	assert.Len(t, invocation, 3)
	assert.Equal(t, "Mailbox/get", invocation[0])
	assert.Equal(t, "0", invocation[2]) // callId
}

func TestRequest_MultipleInvocations(t *testing.T) {
	req := NewRequest().
		WithCapabilities(CapCore, CapMail)

	callID1 := req.Invoke("Mailbox/get", map[string]any{"accountId": "A1"})
	callID2 := req.Invoke("Email/query", map[string]any{"accountId": "A1"})
	callID3 := req.Invoke("Email/get", map[string]any{"accountId": "A1"})

	// Call IDs should be sequential strings
	assert.Equal(t, "0", callID1)
	assert.Equal(t, "1", callID2)
	assert.Equal(t, "2", callID3)

	// Verify JSON structure
	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	methodCalls, ok := parsed["methodCalls"].([]any)
	require.True(t, ok)
	require.Len(t, methodCalls, 3)

	// Verify each call has correct method and ID
	for i, call := range methodCalls {
		invocation := call.([]any)
		assert.Equal(t, []string{"Mailbox/get", "Email/query", "Email/get"}[i], invocation[0])
		assert.Equal(t, []string{"0", "1", "2"}[i], invocation[2])
	}
}

func TestRequest_ResultReference(t *testing.T) {
	req := NewRequest().
		WithCapabilities(CapCore, CapMail)

	// First call: query for email IDs
	queryCallID := req.Invoke("Email/query", map[string]any{
		"accountId": "A1",
		"filter":    map[string]any{"inMailbox": "inbox"},
	})

	// Second call: get emails using result reference
	req.Invoke("Email/get", map[string]any{
		"accountId": "A1",
		"#ids":      ResultReference(queryCallID, "Email/query", "/ids"),
	})

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	methodCalls := parsed["methodCalls"].([]any)
	require.Len(t, methodCalls, 2)

	// Check the result reference in second call
	getCall := methodCalls[1].([]any)
	args := getCall[1].(map[string]any)

	ref, ok := args["#ids"].(map[string]any)
	require.True(t, ok, "result reference should be a map")
	assert.Equal(t, "0", ref["resultOf"])
	assert.Equal(t, "Email/query", ref["name"])
	assert.Equal(t, "/ids", ref["path"])
}

func TestResponse_Parse(t *testing.T) {
	responseJSON := `{
		"sessionState": "abc123",
		"methodResponses": [
			["Mailbox/get", {"accountId": "A1", "state": "m1", "list": [], "notFound": []}, "0"],
			["Email/query", {"accountId": "A1", "queryState": "q1", "ids": ["e1", "e2"]}, "1"]
		]
	}`

	var resp Response
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, "abc123", resp.SessionState)
	require.Len(t, resp.MethodResponses, 2)

	// Check first response
	assert.Equal(t, "Mailbox/get", resp.MethodResponses[0].Name)
	assert.Equal(t, "0", resp.MethodResponses[0].CallID)

	// Check second response
	assert.Equal(t, "Email/query", resp.MethodResponses[1].Name)
	assert.Equal(t, "1", resp.MethodResponses[1].CallID)
}

func TestResponse_GetResult(t *testing.T) {
	responseJSON := `{
		"sessionState": "abc123",
		"methodResponses": [
			["Mailbox/get", {"accountId": "A1", "state": "m1", "list": [{"id": "mb1"}]}, "c0"],
			["Email/query", {"accountId": "A1", "ids": ["e1", "e2"]}, "c1"]
		]
	}`

	var resp Response
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	// Get result by call ID
	result, err := resp.GetResult("c0")
	require.NoError(t, err)
	assert.Equal(t, "Mailbox/get", result.Name)

	result, err = resp.GetResult("c1")
	require.NoError(t, err)
	assert.Equal(t, "Email/query", result.Name)

	// Non-existent call ID
	_, err = resp.GetResult("nonexistent")
	assert.Error(t, err)
}

func TestResponse_Error(t *testing.T) {
	responseJSON := `{
		"sessionState": "abc123",
		"methodResponses": [
			["error", {"type": "unknownMethod", "description": "Method not found"}, "0"]
		]
	}`

	var resp Response
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	result, err := resp.GetResult("0")
	require.NoError(t, err)
	assert.True(t, result.IsError())

	jmapErr := result.Error()
	require.NotNil(t, jmapErr)
	assert.Equal(t, "unknownMethod", jmapErr.Type)
	assert.Equal(t, "Method not found", jmapErr.Description)
}

func TestMethodResult_Decode(t *testing.T) {
	responseJSON := `{
		"sessionState": "abc123",
		"methodResponses": [
			["Email/query", {"accountId": "A1", "queryState": "q1", "ids": ["e1", "e2"], "position": 0}, "0"]
		]
	}`

	var resp Response
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)

	result, err := resp.GetResult("0")
	require.NoError(t, err)

	// Decode into a specific type
	type QueryResult struct {
		AccountID  string   `json:"accountId"`
		QueryState string   `json:"queryState"`
		IDs        []string `json:"ids"`
		Position   int      `json:"position"`
	}

	var qr QueryResult
	err = result.Decode(&qr)
	require.NoError(t, err)

	assert.Equal(t, "A1", qr.AccountID)
	assert.Equal(t, "q1", qr.QueryState)
	assert.Equal(t, []string{"e1", "e2"}, qr.IDs)
	assert.Equal(t, 0, qr.Position)
}

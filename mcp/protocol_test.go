package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID_StringID(t *testing.T) {
	id := NewStringID("abc-123")

	assert.False(t, id.IsNull())
	assert.Equal(t, "abc-123", id.String())
	assert.Equal(t, 0, id.Int())

	data, err := json.Marshal(id)
	require.NoError(t, err)
	assert.Equal(t, `"abc-123"`, string(data))

	var parsed RequestID
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "abc-123", parsed.String())
}

func TestRequestID_IntID(t *testing.T) {
	id := NewIntID(42)

	assert.False(t, id.IsNull())
	assert.Equal(t, "", id.String())
	assert.Equal(t, 42, id.Int())

	data, err := json.Marshal(id)
	require.NoError(t, err)
	assert.Equal(t, "42", string(data))

	var parsed RequestID
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, 42, parsed.Int())
}

func TestRequestID_NullID(t *testing.T) {
	var id RequestID

	assert.True(t, id.IsNull())

	data, err := json.Marshal(id)
	require.NoError(t, err)
	assert.Equal(t, "null", string(data))
}

func TestParseRequest(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantID  string
		method  string
		isNotif bool
	}{
		{
			name:    "basic request with string ID",
			json:    `{"jsonrpc":"2.0","id":"req-1","method":"tools/list"}`,
			wantID:  "req-1",
			method:  "tools/list",
			isNotif: false,
		},
		{
			name:    "request with int ID",
			json:    `{"jsonrpc":"2.0","id":1,"method":"tools/call"}`,
			method:  "tools/call",
			isNotif: false,
		},
		{
			name:    "notification (no ID)",
			json:    `{"jsonrpc":"2.0","method":"notifications/initialized"}`,
			method:  "notifications/initialized",
			isNotif: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseRequest([]byte(tt.json))
			require.NoError(t, err)

			assert.Equal(t, JSONRPCVersion, req.JSONRPC)
			assert.Equal(t, tt.method, req.Method)
			assert.Equal(t, tt.isNotif, req.IsNotification())

			if tt.wantID != "" {
				assert.Equal(t, tt.wantID, req.ID.String())
			}
		})
	}
}

func TestRequest_DecodeParams(t *testing.T) {
	jsonData := `{"jsonrpc":"2.0","id":"1","method":"tools/call","params":{"name":"test","args":{"query":"hello"}}}`

	req, err := ParseRequest([]byte(jsonData))
	require.NoError(t, err)

	var params struct {
		Name string         `json:"name"`
		Args map[string]any `json:"args"`
	}
	err = req.DecodeParams(&params)
	require.NoError(t, err)

	assert.Equal(t, "test", params.Name)
	assert.Equal(t, "hello", params.Args["query"])
}

func TestResponse_Marshal(t *testing.T) {
	resp, err := NewResponse(NewStringID("req-1"), map[string]string{"status": "ok"})
	require.NoError(t, err)

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "2.0", parsed["jsonrpc"])
	assert.Equal(t, "req-1", parsed["id"])
	assert.NotNil(t, parsed["result"])
	assert.Nil(t, parsed["error"])
}

func TestResponse_ErrorResponse(t *testing.T) {
	resp := NewErrorResponse(NewIntID(1), ErrorCodeMethodNotFound, "Method not found")

	assert.True(t, resp.IsError())
	assert.Equal(t, ErrorCodeMethodNotFound, resp.Error.Code)
	assert.Equal(t, "Method not found", resp.Error.Message)

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed Response
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.True(t, parsed.IsError())
	assert.Equal(t, "Method not found", parsed.Error.Error())
}

func TestResponse_DecodeResult(t *testing.T) {
	resp, err := NewResponse(NewStringID("1"), map[string][]string{
		"tools": {"email/send", "email/read"},
	})
	require.NoError(t, err)

	var result struct {
		Tools []string `json:"tools"`
	}
	err = resp.DecodeResult(&result)
	require.NoError(t, err)

	assert.Equal(t, []string{"email/send", "email/read"}, result.Tools)
}

func TestTool_Builder(t *testing.T) {
	tool := NewTool("email/send", "Send an email").
		WithProperty("to", "string", "Recipient email address").
		WithProperty("subject", "string", "Email subject").
		WithProperty("body", "string", "Email body content").
		WithRequired("to", "subject")

	assert.Equal(t, "email/send", tool.Name)
	assert.Equal(t, "Send an email", tool.Description)
	assert.Equal(t, "object", tool.InputSchema.Type)
	assert.Len(t, tool.InputSchema.Properties, 3)
	assert.Equal(t, []string{"to", "subject"}, tool.InputSchema.Required)

	// Verify JSON structure
	data, err := json.Marshal(tool)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "email/send", parsed["name"])
	schema := parsed["inputSchema"].(map[string]any)
	assert.Equal(t, "object", schema["type"])
}

func TestResource_Builder(t *testing.T) {
	resource := NewResource("mailbox://inbox", "Inbox").
		WithDescription("Primary inbox mailbox").
		WithMimeType("application/json")

	assert.Equal(t, "mailbox://inbox", resource.URI)
	assert.Equal(t, "Inbox", resource.Name)
	assert.Equal(t, "Primary inbox mailbox", resource.Description)
	assert.Equal(t, "application/json", resource.MimeType)

	// Verify JSON structure
	data, err := json.Marshal(resource)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "mailbox://inbox", parsed["uri"])
	assert.Equal(t, "Inbox", parsed["name"])
}

func TestResourceContent_Marshal(t *testing.T) {
	content := ResourceContent{
		URI:      "mailbox://inbox",
		MimeType: "text/plain",
		Text:     "Hello, world!",
	}

	data, err := json.Marshal(content)
	require.NoError(t, err)

	var parsed ResourceContent
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, content.URI, parsed.URI)
	assert.Equal(t, content.Text, parsed.Text)
}

func TestPrompt_Marshal(t *testing.T) {
	prompt := Prompt{
		Name:        "compose-email",
		Description: "Compose a new email",
		Arguments: []PromptArgument{
			{Name: "recipient", Description: "Email recipient", Required: true},
			{Name: "tone", Description: "Email tone (formal/casual)", Required: false},
		},
	}

	data, err := json.Marshal(prompt)
	require.NoError(t, err)

	var parsed Prompt
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "compose-email", parsed.Name)
	assert.Len(t, parsed.Arguments, 2)
	assert.True(t, parsed.Arguments[0].Required)
	assert.False(t, parsed.Arguments[1].Required)
}

func TestMethodConstants(t *testing.T) {
	// Verify method constants match MCP spec
	assert.Equal(t, "initialize", MethodInitialize)
	assert.Equal(t, "tools/list", MethodToolsList)
	assert.Equal(t, "tools/call", MethodToolsCall)
	assert.Equal(t, "resources/list", MethodResourcesList)
	assert.Equal(t, "resources/read", MethodResourcesRead)
	assert.Equal(t, "prompts/list", MethodPromptsList)
	assert.Equal(t, "prompts/get", MethodPromptsGet)
}

func TestErrorConstants(t *testing.T) {
	// Verify error codes match JSON-RPC spec
	assert.Equal(t, -32700, ErrorCodeParseError)
	assert.Equal(t, -32600, ErrorCodeInvalidRequest)
	assert.Equal(t, -32601, ErrorCodeMethodNotFound)
	assert.Equal(t, -32602, ErrorCodeInvalidParams)
	assert.Equal(t, -32603, ErrorCodeInternalError)
}

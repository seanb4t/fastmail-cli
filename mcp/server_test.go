package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	server := NewServer("test-server", "1.0.0")

	assert.NotNil(t, server)
	assert.Equal(t, "test-server", server.info.Name)
	assert.Equal(t, "1.0.0", server.info.Version)
	assert.NotNil(t, server.tools)
	assert.NotNil(t, server.resources)
}

func TestServer_RegisterTool(t *testing.T) {
	server := NewServer("test", "1.0.0")

	tool := NewTool("test_tool", "A test tool").
		WithProperty("input", "string", "Test input").
		WithRequired("input")

	handler := func(_ context.Context, _ map[string]any) (any, error) {
		return "result", nil
	}

	server.RegisterTool(tool, handler)

	assert.Len(t, server.tools, 1)
	assert.NotNil(t, server.tools["test_tool"])
	assert.Equal(t, "test_tool", server.tools["test_tool"].tool.Name)
}

func TestServer_RegisterResource(t *testing.T) {
	server := NewServer("test", "1.0.0")

	resource := NewResource("test://resource", "Test Resource").
		WithDescription("A test resource").
		WithMimeType("text/plain")

	reader := func(_ context.Context, _ string) (*ResourceContent, error) {
		return &ResourceContent{
			URI:      "test://resource",
			MimeType: "text/plain",
			Text:     "Hello, World!",
		}, nil
	}

	server.RegisterResource(resource, reader)

	assert.Len(t, server.resources, 1)
	assert.NotNil(t, server.resources["test://resource"])
	assert.NotNil(t, server.readers["test://resource"])
}

func TestServer_Initialize(t *testing.T) {
	server := NewServer("fastmail-mcp", "1.0.0")

	input := `{"jsonrpc":"2.0","id":"init-1","method":"initialize","params":{}}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Run server in goroutine with input that ends
	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	// Wait for response
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Parse response
	var resp Response
	err := json.Unmarshal(output.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, JSONRPCVersion, resp.JSONRPC)
	assert.Equal(t, "init-1", resp.ID.String())
	assert.False(t, resp.IsError())

	var result InitializeResult
	err = resp.DecodeResult(&result)
	require.NoError(t, err)

	assert.Equal(t, ProtocolVersion, result.ProtocolVersion)
	assert.Equal(t, "fastmail-mcp", result.ServerInfo.Name)
	assert.Equal(t, "1.0.0", result.ServerInfo.Version)
	assert.NotNil(t, result.Capabilities.Tools)
	assert.NotNil(t, result.Capabilities.Resources)
}

func TestServer_ToolsList(t *testing.T) {
	server := NewServer("test", "1.0.0")

	// Register some tools
	tool1 := NewTool("email_list", "List emails").
		WithProperty("folder", "string", "Folder name")
	tool2 := NewTool("email_send", "Send email").
		WithProperty("to", "string", "Recipient").
		WithRequired("to")

	server.RegisterTool(tool1, func(_ context.Context, _ map[string]any) (any, error) { return nil, nil })
	server.RegisterTool(tool2, func(_ context.Context, _ map[string]any) (any, error) { return nil, nil })

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	var resp Response
	err := json.Unmarshal(output.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.IsError())

	var result ToolsListResult
	err = resp.DecodeResult(&result)
	require.NoError(t, err)

	assert.Len(t, result.Tools, 2)

	// Check tools are present (order may vary)
	toolNames := make(map[string]bool)
	for _, tool := range result.Tools {
		toolNames[tool.Name] = true
	}
	assert.True(t, toolNames["email_list"])
	assert.True(t, toolNames["email_send"])
}

func TestServer_ToolsCall(t *testing.T) {
	server := NewServer("test", "1.0.0")

	// Register a tool that echoes input
	tool := NewTool("echo", "Echo input").
		WithProperty("message", "string", "Message to echo").
		WithRequired("message")

	server.RegisterTool(tool, func(_ context.Context, args map[string]any) (any, error) {
		msg, _ := args["message"].(string)
		return "Echo: " + msg, nil
	})

	input := `{"jsonrpc":"2.0","id":"call-1","method":"tools/call","params":{"name":"echo","arguments":{"message":"hello"}}}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	var resp Response
	err := json.Unmarshal(output.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.IsError())

	var result ToolsCallResult
	err = resp.DecodeResult(&result)
	require.NoError(t, err)

	assert.False(t, result.IsError)
	require.Len(t, result.Content, 1)
	assert.Equal(t, "text", result.Content[0].Type)
	assert.Equal(t, "Echo: hello", result.Content[0].Text)
}

func TestServer_ToolsCall_Error(t *testing.T) {
	server := NewServer("test", "1.0.0")

	// Register a tool that always fails
	tool := NewTool("failing_tool", "Always fails")

	server.RegisterTool(tool, func(_ context.Context, _ map[string]any) (any, error) {
		return nil, fmt.Errorf("intentional failure")
	})

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"failing_tool"}}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	var resp Response
	err := json.Unmarshal(output.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.IsError()) // JSON-RPC success, but tool result is error

	var result ToolsCallResult
	err = resp.DecodeResult(&result)
	require.NoError(t, err)

	assert.True(t, result.IsError)
	require.Len(t, result.Content, 1)
	assert.Contains(t, result.Content[0].Text, "intentional failure")
}

func TestServer_ToolsCall_UnknownTool(t *testing.T) {
	server := NewServer("test", "1.0.0")

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"nonexistent"}}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	var resp Response
	err := json.Unmarshal(output.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.IsError())
	assert.Equal(t, ErrorCodeInvalidParams, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Unknown tool")
}

func TestServer_ResourcesList(t *testing.T) {
	server := NewServer("test", "1.0.0")

	// Register some resources
	res1 := NewResource("fastmail://inbox", "Inbox").
		WithDescription("Recent inbox messages")
	res2 := NewResource("fastmail://contacts", "Contacts").
		WithDescription("Contact list")

	dummyReader := func(_ context.Context, _ string) (*ResourceContent, error) {
		return &ResourceContent{}, nil
	}

	server.RegisterResource(res1, dummyReader)
	server.RegisterResource(res2, dummyReader)

	input := `{"jsonrpc":"2.0","id":1,"method":"resources/list"}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	var resp Response
	err := json.Unmarshal(output.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.IsError())

	var result ResourcesListResult
	err = resp.DecodeResult(&result)
	require.NoError(t, err)

	assert.Len(t, result.Resources, 2)

	// Check resources are present
	resourceURIs := make(map[string]bool)
	for _, res := range result.Resources {
		resourceURIs[res.URI] = true
	}
	assert.True(t, resourceURIs["fastmail://inbox"])
	assert.True(t, resourceURIs["fastmail://contacts"])
}

func TestServer_ResourcesRead(t *testing.T) {
	server := NewServer("test", "1.0.0")

	resource := NewResource("fastmail://inbox", "Inbox").
		WithMimeType("application/json")

	server.RegisterResource(resource, func(_ context.Context, uri string) (*ResourceContent, error) {
		return &ResourceContent{
			URI:      uri,
			MimeType: "application/json",
			Text:     `[{"id":"1","subject":"Test Email"}]`,
		}, nil
	})

	input := `{"jsonrpc":"2.0","id":"read-1","method":"resources/read","params":{"uri":"fastmail://inbox"}}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	var resp Response
	err := json.Unmarshal(output.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.IsError())

	var result ResourcesReadResult
	err = resp.DecodeResult(&result)
	require.NoError(t, err)

	require.Len(t, result.Contents, 1)
	assert.Equal(t, "fastmail://inbox", result.Contents[0].URI)
	assert.Equal(t, "application/json", result.Contents[0].MimeType)
	assert.Contains(t, result.Contents[0].Text, "Test Email")
}

func TestServer_ResourcesRead_UnknownResource(t *testing.T) {
	server := NewServer("test", "1.0.0")

	input := `{"jsonrpc":"2.0","id":1,"method":"resources/read","params":{"uri":"unknown://resource"}}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	var resp Response
	err := json.Unmarshal(output.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.IsError())
	assert.Equal(t, ErrorCodeInvalidParams, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Unknown resource")
}

func TestServer_MethodNotFound(t *testing.T) {
	server := NewServer("test", "1.0.0")

	input := `{"jsonrpc":"2.0","id":1,"method":"unknown/method"}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	var resp Response
	err := json.Unmarshal(output.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.IsError())
	assert.Equal(t, ErrorCodeMethodNotFound, resp.Error.Code)
}

func TestServer_ParseError(t *testing.T) {
	server := NewServer("test", "1.0.0")

	input := `{invalid json}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	var resp Response
	err := json.Unmarshal(output.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.IsError())
	assert.Equal(t, ErrorCodeParseError, resp.Error.Code)
}

func TestServer_Notification(t *testing.T) {
	server := NewServer("test", "1.0.0")

	// Notification has no ID - should not produce a response
	input := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	// No response expected for notifications
	assert.Empty(t, output.String())
}

func TestServer_ContextCancellation(t *testing.T) {
	server := NewServer("test", "1.0.0")

	// Input that would block (no newline at end)
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx, input, output)
	}()

	// Cancel immediately
	cancel()

	select {
	case err := <-done:
		assert.Equal(t, context.Canceled, err)
	case <-time.After(time.Second):
		t.Fatal("Server did not stop after context cancellation")
	}
}

func TestServer_MultipleRequests(t *testing.T) {
	server := NewServer("test", "1.0.0")

	tool := NewTool("add", "Add numbers").
		WithProperty("a", "number", "First number").
		WithProperty("b", "number", "Second number")

	server.RegisterTool(tool, func(_ context.Context, args map[string]any) (any, error) {
		a, _ := args["a"].(float64)
		b, _ := args["b"].(float64)
		return a + b, nil
	})

	// Send multiple requests
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"add","arguments":{"a":2,"b":3}}}
`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input), output)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()

	// Parse all responses
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	assert.Len(t, lines, 3)

	// Verify each response
	for i, line := range lines {
		var resp Response
		err := json.Unmarshal([]byte(line), &resp)
		require.NoError(t, err)
		assert.False(t, resp.IsError(), "Response %d should not be an error", i+1)
	}
}

func TestServer_ToolsCall_JSONResult(t *testing.T) {
	server := NewServer("test", "1.0.0")

	type Result struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}

	tool := NewTool("get_status", "Get status")

	server.RegisterTool(tool, func(_ context.Context, _ map[string]any) (any, error) {
		return Result{Status: "ok", Count: 42}, nil
	})

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_status"}}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	var resp Response
	err := json.Unmarshal(output.Bytes(), &resp)
	require.NoError(t, err)

	var result ToolsCallResult
	err = resp.DecodeResult(&result)
	require.NoError(t, err)

	require.Len(t, result.Content, 1)
	// Result should be JSON-encoded
	assert.Contains(t, result.Content[0].Text, `"status":"ok"`)
	assert.Contains(t, result.Content[0].Text, `"count":42`)
}

func TestServer_EmptyLines(t *testing.T) {
	server := NewServer("test", "1.0.0")

	// Input with empty lines that should be ignored
	input := `
{"jsonrpc":"2.0","id":1,"method":"initialize"}

`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Run(ctx, strings.NewReader(input), output)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	// Should have exactly one response
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	assert.Len(t, lines, 1)

	var resp Response
	err := json.Unmarshal([]byte(lines[0]), &resp)
	require.NoError(t, err)
	assert.False(t, resp.IsError())
}

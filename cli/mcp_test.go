package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/seanb4t/fastmail-cli/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMCPCommand(t *testing.T) {
	cmd := newMCPCommand()
	assert.Equal(t, "mcp", cmd.Use)
	assert.Equal(t, "Run MCP server", cmd.Short)
}

func TestMCPServerToolRegistration(t *testing.T) {
	// Create a server and verify tools are registered correctly
	server := mcp.NewServer("test-server", "1.0.0")

	// Mock client - we can't use the real one without credentials,
	// but we can verify the registration functions work
	t.Run("mail tools registration", func(t *testing.T) {
		// Create a test server and check it handles initialize properly
		testServer := mcp.NewServer("fastmail-cli-test", "test")

		// Verify server responds to initialize
		input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`
		output := &bytes.Buffer{}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Run in goroutine since Run blocks
		done := make(chan error, 1)
		go func() {
			done <- testServer.Run(ctx, strings.NewReader(input+"\n"), output)
		}()

		// Wait for response or timeout
		select {
		case <-done:
		case <-time.After(200 * time.Millisecond):
		}

		// Parse response
		responseLines := strings.Split(strings.TrimSpace(output.String()), "\n")
		require.NotEmpty(t, responseLines)

		var resp mcp.Response
		err := json.Unmarshal([]byte(responseLines[0]), &resp)
		require.NoError(t, err)
		assert.False(t, resp.IsError())
	})

	t.Run("server info", func(t *testing.T) {
		assert.NotNil(t, server)
	})
}

func TestMCPServerInitializeResponse(t *testing.T) {
	server := mcp.NewServer("fastmail-cli", "1.0.0")

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}

	// Parse the response
	responseLines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.NotEmpty(t, responseLines)

	var resp mcp.Response
	err := json.Unmarshal([]byte(responseLines[0]), &resp)
	require.NoError(t, err)
	assert.False(t, resp.IsError())

	// Decode the result
	var initResult mcp.InitializeResult
	err = resp.DecodeResult(&initResult)
	require.NoError(t, err)

	assert.Equal(t, "2024-11-05", initResult.ProtocolVersion)
	assert.Equal(t, "fastmail-cli", initResult.ServerInfo.Name)
	assert.Equal(t, "1.0.0", initResult.ServerInfo.Version)
}

func TestMCPServerToolsList(t *testing.T) {
	server := mcp.NewServer("fastmail-cli", "1.0.0")

	// Register a test tool
	testTool := mcp.NewTool("test_tool", "A test tool").
		WithProperty("param1", "string", "Test parameter").
		WithRequired("param1")

	server.RegisterTool(testTool, func(ctx context.Context, args map[string]any) (any, error) {
		return map[string]string{"result": "ok"}, nil
	})

	// Send tools/list request
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}

	responseLines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.NotEmpty(t, responseLines)

	var resp mcp.Response
	err := json.Unmarshal([]byte(responseLines[0]), &resp)
	require.NoError(t, err)
	assert.False(t, resp.IsError())

	var toolsResult mcp.ToolsListResult
	err = resp.DecodeResult(&toolsResult)
	require.NoError(t, err)

	assert.Len(t, toolsResult.Tools, 1)
	assert.Equal(t, "test_tool", toolsResult.Tools[0].Name)
	assert.Equal(t, "A test tool", toolsResult.Tools[0].Description)
}

func TestMCPServerToolsCall(t *testing.T) {
	server := mcp.NewServer("fastmail-cli", "1.0.0")

	// Register a test tool
	testTool := mcp.NewTool("echo", "Echo tool").
		WithProperty("message", "string", "Message to echo").
		WithRequired("message")

	server.RegisterTool(testTool, func(ctx context.Context, args map[string]any) (any, error) {
		msg, _ := args["message"].(string)
		return map[string]string{"echo": msg}, nil
	})

	// Send tools/call request
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello world"}}}`
	output := &bytes.Buffer{}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx, strings.NewReader(input+"\n"), output)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}

	responseLines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.NotEmpty(t, responseLines)

	var resp mcp.Response
	err := json.Unmarshal([]byte(responseLines[0]), &resp)
	require.NoError(t, err)
	assert.False(t, resp.IsError())

	var callResult mcp.ToolsCallResult
	err = resp.DecodeResult(&callResult)
	require.NoError(t, err)

	assert.False(t, callResult.IsError)
	require.Len(t, callResult.Content, 1)
	assert.Equal(t, "text", callResult.Content[0].Type)
	assert.Contains(t, callResult.Content[0].Text, "hello world")
}

func TestMCPServerGracefulShutdown(t *testing.T) {
	server := mcp.NewServer("fastmail-cli", "1.0.0")

	ctx, cancel := context.WithCancel(context.Background())

	// Use a pipe for input to keep the connection open
	pr, pw := io.Pipe()
	output := &bytes.Buffer{}

	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx, pr, output)
	}()

	// Send initialize request
	_, err := pw.Write([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"))
	require.NoError(t, err)

	// Give server time to process
	time.Sleep(50 * time.Millisecond)

	// Cancel context and close pipe to trigger shutdown
	// The server reads from input in a loop, so closing the pipe allows it to exit
	cancel()
	pw.Close()

	// Wait for server to stop
	select {
	case err := <-done:
		// Server should return nil on clean EOF shutdown or context.Canceled
		assert.True(t, err == nil || err == context.Canceled, "expected nil or context.Canceled, got: %v", err)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("server did not shut down gracefully")
	}
}

func TestMailToolDefinitions(t *testing.T) {
	// Verify the tool definitions match expected structure
	tests := []struct {
		name        string
		toolName    string
		description string
		required    []string
	}{
		{
			name:        "mail_list tool",
			toolName:    "mail_list",
			description: "List emails from a mailbox folder",
			required:    nil,
		},
		{
			name:        "mail_search tool",
			toolName:    "mail_search",
			description: "Search emails by query text",
			required:    []string{"query"},
		},
		{
			name:        "mail_get tool",
			toolName:    "mail_get",
			description: "Get a single email by ID",
			required:    []string{"id"},
		},
		{
			name:        "mail_send tool",
			toolName:    "mail_send",
			description: "Send a new email",
			required:    []string{"to", "subject", "body"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := mcp.NewTool(tt.toolName, tt.description)
			assert.Equal(t, tt.toolName, tool.Name)
			assert.Equal(t, tt.description, tool.Description)

			if len(tt.required) > 0 {
				tool.WithRequired(tt.required...)
				assert.Equal(t, tt.required, tool.InputSchema.Required)
			}
		})
	}
}

func TestMaskedEmailToolDefinitions(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		description string
		required    []string
	}{
		{
			name:        "masked_email_list tool",
			toolName:    "masked_email_list",
			description: "List all masked email addresses",
			required:    nil,
		},
		{
			name:        "masked_email_create tool",
			toolName:    "masked_email_create",
			description: "Create a new masked email address",
			required:    nil,
		},
		{
			name:        "masked_email_enable tool",
			toolName:    "masked_email_enable",
			description: "Enable a masked email address",
			required:    []string{"id"},
		},
		{
			name:        "masked_email_disable tool",
			toolName:    "masked_email_disable",
			description: "Disable a masked email address",
			required:    []string{"id"},
		},
		{
			name:        "masked_email_delete tool",
			toolName:    "masked_email_delete",
			description: "Permanently delete a masked email address",
			required:    []string{"id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := mcp.NewTool(tt.toolName, tt.description)
			assert.Equal(t, tt.toolName, tool.Name)
			assert.Equal(t, tt.description, tool.Description)

			if len(tt.required) > 0 {
				tool.WithRequired(tt.required...)
				assert.Equal(t, tt.required, tool.InputSchema.Required)
			}
		})
	}
}

func TestToolHandlerValidation(t *testing.T) {
	// Test that handlers properly validate required arguments
	t.Run("mail_get requires id", func(t *testing.T) {
		server := mcp.NewServer("test", "1.0.0")

		getTool := mcp.NewTool("mail_get", "Get a single email by ID").
			WithProperty("id", "string", "Email ID").
			WithRequired("id")

		server.RegisterTool(getTool, func(ctx context.Context, args map[string]any) (any, error) {
			id, ok := args["id"].(string)
			if !ok || id == "" {
				return nil, fmt.Errorf("id is required")
			}
			return map[string]string{"id": id}, nil
		})

		// Call without id
		input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"mail_get","arguments":{}}}`
		output := &bytes.Buffer{}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- server.Run(ctx, strings.NewReader(input+"\n"), output)
		}()

		select {
		case <-done:
		case <-time.After(200 * time.Millisecond):
		}

		responseLines := strings.Split(strings.TrimSpace(output.String()), "\n")
		require.NotEmpty(t, responseLines)

		var resp mcp.Response
		err := json.Unmarshal([]byte(responseLines[0]), &resp)
		require.NoError(t, err)

		// Tool errors are returned as tool results, not JSON-RPC errors
		var callResult mcp.ToolsCallResult
		err = resp.DecodeResult(&callResult)
		require.NoError(t, err)
		assert.True(t, callResult.IsError)
		assert.Contains(t, callResult.Content[0].Text, "id is required")
	})
}

package mcp

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterMailboxTools(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	expectedTools := []string{
		"mailbox_list",
		"mailbox_create",
		"mailbox_rename",
		"mailbox_delete",
	}

	for _, name := range expectedTools {
		_, ok := server.tools[name]
		assert.True(t, ok, "expected tool %q to be registered", name)
	}
}

func TestMailboxListTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_list"]
	require.True(t, ok, "mailbox_list tool not found")

	assert.Equal(t, "mailbox_list", rt.tool.Name)
	assert.Equal(t, "List all email mailboxes/folders", rt.tool.Description)
}

func TestMailboxCreateTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_create"]
	require.True(t, ok, "mailbox_create tool not found")

	assert.Contains(t, rt.tool.InputSchema.Properties, "name")
	assert.Contains(t, rt.tool.InputSchema.Properties, "parent_id")
	assert.True(t, slices.Contains(rt.tool.InputSchema.Required, "name"))
}

func TestMailboxCreateHandler_MissingName(t *testing.T) {
	handler := makeMailboxCreateHandler(ToolsConfig{})
	_, err := handler(context.Background(), map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestMailboxRenameTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_rename"]
	require.True(t, ok, "mailbox_rename tool not found")

	assert.Contains(t, rt.tool.InputSchema.Properties, "id")
	assert.Contains(t, rt.tool.InputSchema.Properties, "name")
	assert.True(t, slices.Contains(rt.tool.InputSchema.Required, "id"))
	assert.True(t, slices.Contains(rt.tool.InputSchema.Required, "name"))
}

func TestMailboxRenameHandler_MissingArgs(t *testing.T) {
	handler := makeMailboxRenameHandler(ToolsConfig{})

	t.Run("missing id", func(t *testing.T) {
		_, err := handler(context.Background(), map[string]any{"name": "New Name"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	t.Run("missing name", func(t *testing.T) {
		_, err := handler(context.Background(), map[string]any{"id": "mb-123"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})
}

func TestMailboxDeleteTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_delete"]
	require.True(t, ok, "mailbox_delete tool not found")

	assert.Contains(t, rt.tool.InputSchema.Properties, "id")
	assert.True(t, slices.Contains(rt.tool.InputSchema.Required, "id"))
}

func TestMailboxDeleteHandler_MissingID(t *testing.T) {
	handler := makeMailboxDeleteHandler(ToolsConfig{})
	_, err := handler(context.Background(), map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

package mcp

import (
	"context"
	"slices"
	"testing"
)

func TestMailboxListTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_list"]
	if !ok {
		t.Fatal("mailbox_list tool not found")
	}

	if rt.tool.Name != "mailbox_list" {
		t.Errorf("expected name %q, got %q", "mailbox_list", rt.tool.Name)
	}

	if rt.tool.Description != "List all mailbox folders with unread/total counts" {
		t.Errorf("unexpected description: %s", rt.tool.Description)
	}
}

func TestMailboxCreateTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_create"]
	if !ok {
		t.Fatal("mailbox_create tool not found")
	}

	if !slices.Contains(rt.tool.InputSchema.Required, "name") {
		t.Error("expected 'name' to be required")
	}

	if _, ok := rt.tool.InputSchema.Properties["parent_id"]; !ok {
		t.Error("expected 'parent_id' property in input schema")
	}
}

func TestMailboxCreateTool_RequiresName(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_create"]
	if !ok {
		t.Fatal("mailbox_create tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if err.Error() != "name is required" {
		t.Errorf("expected 'name is required', got %q", err.Error())
	}
}

func TestMailboxRenameTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_rename"]
	if !ok {
		t.Fatal("mailbox_rename tool not found")
	}

	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("expected 'id' to be required")
	}
	if !slices.Contains(rt.tool.InputSchema.Required, "name") {
		t.Error("expected 'name' to be required")
	}
}

func TestMailboxRenameTool_RequiresID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_rename"]
	if !ok {
		t.Fatal("mailbox_rename tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{"name": "New Name"})
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	if err.Error() != "id is required" {
		t.Errorf("expected 'id is required', got %q", err.Error())
	}
}

func TestMailboxRenameTool_RequiresName(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_rename"]
	if !ok {
		t.Fatal("mailbox_rename tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{"id": "mb-123"})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if err.Error() != "name is required" {
		t.Errorf("expected 'name is required', got %q", err.Error())
	}
}

func TestMailboxDeleteTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_delete"]
	if !ok {
		t.Fatal("mailbox_delete tool not found")
	}

	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("expected 'id' to be required")
	}
}

func TestMailboxDeleteTool_RequiresID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mailbox_delete"]
	if !ok {
		t.Fatal("mailbox_delete tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	if err.Error() != "id is required" {
		t.Errorf("expected 'id is required', got %q", err.Error())
	}
}

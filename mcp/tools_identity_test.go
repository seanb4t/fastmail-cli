package mcp

import (
	"context"
	"slices"
	"testing"
)

func TestIdentityToolsRegistered(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	expectedTools := []string{
		"identity_list",
		"identity_set",
	}

	for _, name := range expectedTools {
		if _, ok := server.tools[name]; !ok {
			t.Errorf("expected tool %q to be registered", name)
		}
	}
}

func TestIdentityListTool_Definition(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["identity_list"]
	if !ok {
		t.Fatal("identity_list tool not found")
	}

	if rt.tool.Name != "identity_list" {
		t.Errorf("expected name %q, got %q", "identity_list", rt.tool.Name)
	}

	if rt.tool.Description != "List all sender identities" {
		t.Errorf("unexpected description: %s", rt.tool.Description)
	}
}

func TestIdentitySetTool_Definition(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["identity_set"]
	if !ok {
		t.Fatal("identity_set tool not found")
	}

	if rt.tool.Name != "identity_set" {
		t.Errorf("expected name %q, got %q", "identity_set", rt.tool.Name)
	}

	// Verify required fields
	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("expected 'id' to be required")
	}

	// Verify optional properties exist
	for _, prop := range []string{"id", "name", "reply_to", "text_signature", "html_signature"} {
		if _, ok := rt.tool.InputSchema.Properties[prop]; !ok {
			t.Errorf("expected %q property in input schema", prop)
		}
	}
}

func TestIdentitySetTool_RequiresID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["identity_set"]
	if !ok {
		t.Fatal("identity_set tool not found")
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

func TestIdentitySetTool_RequiresAtLeastOneField(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["identity_set"]
	if !ok {
		t.Fatal("identity_set tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{
		"id": "id-1",
	})
	if err == nil {
		t.Fatal("expected error for no fields specified")
	}
	if err.Error() != "at least one of name, reply_to, text_signature, or html_signature must be specified" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegisterMailTools_IncludesIdentityTools(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	expectedTools := []string{
		"identity_list",
		"identity_set",
	}

	for _, name := range expectedTools {
		found := false
		for _, rt := range server.tools {
			if rt.tool.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected tool %q to be registered via RegisterMailTools", name)
		}
	}
}

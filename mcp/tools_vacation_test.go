package mcp

import (
	"context"
	"slices"
	"testing"
)

func TestVacationToolsRegistered(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	expectedTools := []string{
		"vacation_status",
		"vacation_set",
	}

	for _, name := range expectedTools {
		if _, ok := server.tools[name]; !ok {
			t.Errorf("expected tool %q to be registered", name)
		}
	}
}

func TestVacationStatusTool_Definition(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["vacation_status"]
	if !ok {
		t.Fatal("vacation_status tool not found")
	}

	if rt.tool.Name != "vacation_status" {
		t.Errorf("expected name %q, got %q", "vacation_status", rt.tool.Name)
	}
}

func TestVacationSetTool_Definition(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["vacation_set"]
	if !ok {
		t.Fatal("vacation_set tool not found")
	}

	if rt.tool.Name != "vacation_set" {
		t.Errorf("expected name %q, got %q", "vacation_set", rt.tool.Name)
	}

	// Verify required fields
	if !slices.Contains(rt.tool.InputSchema.Required, "enabled") {
		t.Error("expected 'enabled' to be required")
	}

	// Verify optional properties exist
	for _, prop := range []string{"enabled", "subject", "body", "from_date", "to_date"} {
		if _, ok := rt.tool.InputSchema.Properties[prop]; !ok {
			t.Errorf("expected %q property in input schema", prop)
		}
	}
}

func TestVacationSetTool_RequiresEnabled(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["vacation_set"]
	if !ok {
		t.Fatal("vacation_set tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing enabled field")
	}
}

func TestVacationSetTool_RequiresSubjectAndBody(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["vacation_set"]
	if !ok {
		t.Fatal("vacation_set tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{
		"enabled": true,
	})
	if err == nil {
		t.Fatal("expected error for missing subject/body")
	}
	if err.Error() != "subject and body are required when enabling vacation response" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVacationSetTool_RejectsFromAfterTo(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["vacation_set"]
	if !ok {
		t.Fatal("vacation_set tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{
		"enabled":   true,
		"subject":   "OOO",
		"body":      "I am away",
		"from_date": "2025-03-01T00:00:00Z",
		"to_date":   "2025-02-01T00:00:00Z",
	})
	if err == nil {
		t.Fatal("expected error for from_date after to_date")
	}
	if err.Error() != "from_date must be before to_date" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegisterMailTools_IncludesVacationTools(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	// Verify vacation tools are included in the full tool registration
	expectedTools := []string{
		"vacation_status",
		"vacation_set",
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

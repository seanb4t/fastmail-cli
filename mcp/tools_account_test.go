package mcp

import (
	"testing"
)

func TestQuotaGetToolRegistered(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["quota_get"]
	if !ok {
		t.Fatal("quota_get tool not found")
	}

	if rt.tool.Name != "quota_get" {
		t.Errorf("expected name %q, got %q", "quota_get", rt.tool.Name)
	}

	if rt.tool.Description == "" {
		t.Error("expected non-empty description for quota_get tool")
	}
}

func TestQuotaGetToolInAllTools(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	found := false
	for _, rt := range server.tools {
		if rt.tool.Name == "quota_get" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected quota_get to be registered among all tools")
	}
}

func TestRegisterMailTools_IncludesQuotaGet(t *testing.T) {
	server := NewServer("test", "1.0")
	cfg := ToolsConfig{}

	RegisterMailTools(server, cfg)

	// Verify quota_get is among registered tools
	expectedTools := []string{
		"mail_list",
		"mail_get",
		"quota_get",
	}

	for _, name := range expectedTools {
		if _, ok := server.tools[name]; !ok {
			t.Errorf("expected tool %q to be registered", name)
		}
	}
}

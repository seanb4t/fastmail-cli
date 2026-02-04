package mcp

import (
	"context"
	"regexp"
	"testing"
)

func TestResourceRegistry_List(t *testing.T) {
	registry := NewResourceRegistry(nil)

	resources := registry.List()
	if len(resources) == 0 {
		t.Error("expected at least one resource, got none")
	}

	// Verify expected resources are registered
	expectedURIs := map[string]bool{
		"fastmail://inbox":          false,
		"fastmail://mail/{id}":      false,
		"fastmail://contacts":       false,
		"fastmail://calendar/today": false,
		"fastmail://masked-emails":  false,
	}

	for _, res := range resources {
		if _, ok := expectedURIs[res.URI]; ok {
			expectedURIs[res.URI] = true
		}
	}

	for uri, found := range expectedURIs {
		if !found {
			t.Errorf("expected resource %q not found", uri)
		}
	}
}

func TestResourceRegistry_Templates(t *testing.T) {
	registry := NewResourceRegistry(nil)

	templates := registry.Templates()
	if len(templates) == 0 {
		t.Error("expected at least one template, got none")
	}

	// Verify mail template is registered
	found := false
	for _, tmpl := range templates {
		if tmpl.URITemplate == "fastmail://mail/{id}" {
			found = true
			if tmpl.Name == "" {
				t.Error("template name should not be empty")
			}
			if tmpl.Description == "" {
				t.Error("template description should not be empty")
			}
		}
	}
	if !found {
		t.Error("mail template not found")
	}
}

func TestResourceRegistry_Read_NotFound(t *testing.T) {
	registry := NewResourceRegistry(nil)

	_, err := registry.Read(context.Background(), "fastmail://nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent resource")
	}
}

func TestResourceRegistry_Read_NoClient(t *testing.T) {
	registry := NewResourceRegistry(nil)

	// Should return error when client is nil
	_, err := registry.Read(context.Background(), "fastmail://inbox")
	if err == nil {
		t.Error("expected error when client is not configured")
	}
}

func TestResourceRegistry_URIPatternMatching(t *testing.T) {
	registry := NewResourceRegistry(nil)

	tests := []struct {
		uri       string
		wantMatch bool
	}{
		{"fastmail://mail/abc123", true},
		{"fastmail://mail/email-id-with-dashes", true},
		{"fastmail://mail/", false},
		{"fastmail://mail/id/extra", false},
		{"fastmail://inbox", true},
		{"fastmail://contacts", true},
		{"fastmail://calendar/today", true},
		{"fastmail://masked-emails", true},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			_, err := registry.Read(context.Background(), tt.uri)
			// We expect "client not configured" errors for valid URIs
			// and "resource not found" errors for invalid URIs
			if tt.wantMatch {
				if err != nil && err.Error() == "resource not found: "+tt.uri {
					t.Errorf("URI %q should match a resource", tt.uri)
				}
			} else {
				if err == nil || err.Error() != "resource not found: "+tt.uri {
					t.Errorf("URI %q should not match any resource", tt.uri)
				}
			}
		})
	}
}

func TestExtractParams(t *testing.T) {
	pattern := regexp.MustCompile(`^fastmail://mail/(?P<id>[^/]+)$`)
	matches := pattern.FindStringSubmatch("fastmail://mail/test123")

	params := extractParams(pattern, matches)

	if params["id"] != "test123" {
		t.Errorf("expected id=test123, got id=%s", params["id"])
	}
}

func TestFormatEmailList_Empty(t *testing.T) {
	result := formatEmailList(nil)
	if result != "No emails found." {
		t.Errorf("expected 'No emails found.', got %q", result)
	}
}

func TestFormatContactList_Empty(t *testing.T) {
	result := formatContactList(nil)
	if result != "No contacts found." {
		t.Errorf("expected 'No contacts found.', got %q", result)
	}
}

func TestFormatMaskedEmailList_Empty(t *testing.T) {
	result := formatMaskedEmailList(nil)
	if result != "No masked emails found." {
		t.Errorf("expected 'No masked emails found.', got %q", result)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer string", 10, "this is..."},
		{"", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestNewResource(t *testing.T) {
	res := NewResource("fastmail://test", "Test Resource")

	if res.URI != "fastmail://test" {
		t.Errorf("expected URI 'fastmail://test', got %q", res.URI)
	}
	if res.Name != "Test Resource" {
		t.Errorf("expected Name 'Test Resource', got %q", res.Name)
	}
}

func TestResource_WithDescription(t *testing.T) {
	res := NewResource("fastmail://test", "Test").
		WithDescription("A test resource")

	if res.Description != "A test resource" {
		t.Errorf("expected description 'A test resource', got %q", res.Description)
	}
}

func TestResource_WithMimeType(t *testing.T) {
	res := NewResource("fastmail://test", "Test").
		WithMimeType("application/json")

	if res.MimeType != "application/json" {
		t.Errorf("expected MimeType 'application/json', got %q", res.MimeType)
	}
}

func TestResourceContent(t *testing.T) {
	content := &ResourceContent{
		URI:      "fastmail://test",
		MimeType: "text/plain",
		Text:     "Test content",
	}

	if content.URI != "fastmail://test" {
		t.Errorf("expected URI 'fastmail://test', got %q", content.URI)
	}
	if content.Text != "Test content" {
		t.Errorf("expected Text 'Test content', got %q", content.Text)
	}
}

func TestCalendarToday_NoConfig(t *testing.T) {
	registry := NewResourceRegistry(nil)

	// Calendar should return informative message even without client
	content, err := registry.Read(context.Background(), "fastmail://calendar/today")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content.Text == "" {
		t.Error("expected informative message about calendar config")
	}
}

package mcp

import (
	"bytes"
	"context"
	"slices"
	"testing"
	"time"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func TestRegisterMailTools(t *testing.T) {
	server := NewServer("test", "1.0")
	cfg := ToolsConfig{}

	RegisterMailTools(server, cfg)

	// Verify all expected tools are registered
	expectedTools := []string{
		"mail_list",
		"mail_get",
		"mail_search",
		"mail_send",
		"mail_reply",
		"mail_move",
		"mail_delete",
		"mail_flag",
		"mail_thread",
		"mail_attachments",
		"mail_download",
		"mail_upload",
		"mail_import",
		"mailbox_list",
		"mailbox_create",
		"mailbox_rename",
		"mailbox_delete",
		"masked_email_list",
		"masked_email_create",
		"masked_email_enable",
		"masked_email_disable",
		"masked_email_delete",
		"contacts_list",
		"contacts_get",
		"contacts_create",
		"contacts_update",
		"contacts_delete",
		"calendar_events",
		"calendar_create",
		"vacation_status",
		"vacation_set",
		"quota_get",
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
			t.Errorf("expected tool %q to be registered", name)
		}
	}
}

func TestMailListTool(t *testing.T) {
	server := NewServer("test", "1.0")

	// Create a mock that tracks calls
	listCalled := false
	listFolder := ""
	listLimit := uint64(0)

	// We need to use the actual handler pattern
	// For now, test the tool definition
	RegisterMailTools(server, ToolsConfig{})

	// Find the mail_list tool
	rt, ok := server.tools["mail_list"]
	if !ok {
		t.Fatal("mail_list tool not found")
	}

	// Verify tool definition
	if rt.tool.Name != "mail_list" {
		t.Errorf("expected name %q, got %q", "mail_list", rt.tool.Name)
	}

	if rt.tool.Description != "List emails from a mailbox folder" {
		t.Errorf("unexpected description: %s", rt.tool.Description)
	}

	// Verify input schema has expected properties
	if _, ok := rt.tool.InputSchema.Properties["folder"]; !ok {
		t.Error("expected 'folder' property in input schema")
	}
	if _, ok := rt.tool.InputSchema.Properties["limit"]; !ok {
		t.Error("expected 'limit' property in input schema")
	}

	_ = listCalled
	_ = listFolder
	_ = listLimit
}

func TestMailGetTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_get"]
	if !ok {
		t.Fatal("mail_get tool not found")
	}

	// Verify required fields
	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("expected 'id' to be required")
	}
}

func TestMailSendTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_send"]
	if !ok {
		t.Fatal("mail_send tool not found")
	}

	// Verify required fields
	required := make(map[string]bool)
	for _, req := range rt.tool.InputSchema.Required {
		required[req] = true
	}

	if !required["to"] {
		t.Error("expected 'to' to be required")
	}
	if !required["subject"] {
		t.Error("expected 'subject' to be required")
	}
	if !required["body"] {
		t.Error("expected 'body' to be required")
	}
}

func TestMailReplyTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_reply"]
	if !ok {
		t.Fatal("mail_reply tool not found")
	}

	// Verify has reply_all property
	if _, ok := rt.tool.InputSchema.Properties["reply_all"]; !ok {
		t.Error("expected 'reply_all' property in input schema")
	}
}

func TestMaskedEmailTools(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	// Verify masked email tools
	if _, ok := server.tools["masked_email_list"]; !ok {
		t.Error("masked_email_list tool not found")
	}
	if _, ok := server.tools["masked_email_create"]; !ok {
		t.Error("masked_email_create tool not found")
	}
}

func TestContactTools(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	// Verify contact tools
	if _, ok := server.tools["contacts_list"]; !ok {
		t.Error("contacts_list tool not found")
	}
	if _, ok := server.tools["contacts_get"]; !ok {
		t.Error("contacts_get tool not found")
	}
	if _, ok := server.tools["contacts_create"]; !ok {
		t.Error("contacts_create tool not found")
	}
}

func TestCalendarTools(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	// Verify calendar tools
	if _, ok := server.tools["calendar_events"]; !ok {
		t.Error("calendar_events tool not found")
	}
	if _, ok := server.tools["calendar_create"]; !ok {
		t.Error("calendar_create tool not found")
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("getStringArg", func(t *testing.T) {
		args := map[string]any{
			"name": "test",
		}

		if got := getStringArg(args, "name", "default"); got != "test" {
			t.Errorf("expected 'test', got %q", got)
		}
		if got := getStringArg(args, "missing", "default"); got != "default" {
			t.Errorf("expected 'default', got %q", got)
		}
	})

	t.Run("getUint64Arg", func(t *testing.T) {
		args := map[string]any{
			"float":   float64(42),
			"int":     int(42),
			"int64":   int64(42),
			"uint64":  uint64(42),
			"invalid": "not a number",
		}

		for _, key := range []string{"float", "int", "int64", "uint64"} {
			if got := getUint64Arg(args, key, 0); got != 42 {
				t.Errorf("getUint64Arg(%s): expected 42, got %d", key, got)
			}
		}

		if got := getUint64Arg(args, "invalid", 10); got != 10 {
			t.Errorf("expected default 10, got %d", got)
		}
		if got := getUint64Arg(args, "missing", 10); got != 10 {
			t.Errorf("expected default 10, got %d", got)
		}
	})

	t.Run("getBoolArg", func(t *testing.T) {
		args := map[string]any{
			"true":  true,
			"false": false,
		}

		if got := getBoolArg(args, "true", false); !got {
			t.Error("expected true")
		}
		if got := getBoolArg(args, "false", true); got {
			t.Error("expected false")
		}
		if got := getBoolArg(args, "missing", true); !got {
			t.Error("expected default true")
		}
	})

	t.Run("getStringSliceArg", func(t *testing.T) {
		args := map[string]any{
			"strings": []string{"a", "b"},
			"anys":    []any{"c", "d"},
		}

		if got := getStringSliceArg(args, "strings"); len(got) != 2 || got[0] != "a" {
			t.Errorf("expected [a b], got %v", got)
		}
		if got := getStringSliceArg(args, "anys"); len(got) != 2 || got[0] != "c" {
			t.Errorf("expected [c d], got %v", got)
		}
		if got := getStringSliceArg(args, "missing"); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
}

func TestEmailConversion(t *testing.T) {
	now := time.Now()
	email := fastmail.Email{
		ID:         "email-123",
		ThreadID:   "thread-456",
		Subject:    "Test Subject",
		Preview:    "Test preview...",
		ReceivedAt: now,
		Size:       1024,
		Keywords:   []string{"$seen", "$flagged"},
		MailboxIDs: []string{"inbox-id"},
	}

	mcp := convertEmailToMCP(&email)

	if mcp.ID != "email-123" {
		t.Errorf("expected ID 'email-123', got %q", mcp.ID)
	}
	if mcp.ThreadID != "thread-456" {
		t.Errorf("expected ThreadID 'thread-456', got %q", mcp.ThreadID)
	}
	if mcp.Subject != "Test Subject" {
		t.Errorf("expected Subject 'Test Subject', got %q", mcp.Subject)
	}
	if mcp.Size != 1024 {
		t.Errorf("expected Size 1024, got %d", mcp.Size)
	}
	if !mcp.IsRead {
		t.Error("expected IsRead to be true")
	}
	if !mcp.IsFlagged {
		t.Error("expected IsFlagged to be true")
	}
}

func TestEventConversion(t *testing.T) {
	now := time.Now()
	event := fastmail.Event{
		ID:          "event-123",
		CalendarID:  "cal-456",
		Summary:     "Test Event",
		Description: "Test description",
		Location:    "Test Location",
		Start:       now,
		End:         now.Add(time.Hour),
		AllDay:      false,
		Status:      "CONFIRMED",
	}

	mcp := convertEventToMCP(&event)

	if mcp.ID != "event-123" {
		t.Errorf("expected ID 'event-123', got %q", mcp.ID)
	}
	if mcp.CalendarID != "cal-456" {
		t.Errorf("expected CalendarID 'cal-456', got %q", mcp.CalendarID)
	}
	if mcp.Summary != "Test Event" {
		t.Errorf("expected Summary 'Test Event', got %q", mcp.Summary)
	}
	if mcp.Status != "CONFIRMED" {
		t.Errorf("expected Status 'CONFIRMED', got %q", mcp.Status)
	}
}

func TestParseAddresses(t *testing.T) {
	addrs := []string{"alice@example.com", "bob@example.com"}
	result := parseAddresses(addrs)

	if len(result) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(result))
	}
	if result[0].Email != "alice@example.com" {
		t.Errorf("expected 'alice@example.com', got %q", result[0].Email)
	}
	if result[1].Email != "bob@example.com" {
		t.Errorf("expected 'bob@example.com', got %q", result[1].Email)
	}
}

func TestMaskedEmailEnable_RequiresID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["masked_email_enable"]
	if !ok {
		t.Fatal("masked_email_enable tool not found")
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

func TestMaskedEmailDisable_RequiresID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["masked_email_disable"]
	if !ok {
		t.Fatal("masked_email_disable tool not found")
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

func TestMaskedEmailDelete_RequiresID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["masked_email_delete"]
	if !ok {
		t.Fatal("masked_email_delete tool not found")
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

func TestContactsUpdate_RequiresID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{
		Contacts: &fastmail.ContactsClient{},
	})

	rt, ok := server.tools["contacts_update"]
	if !ok {
		t.Fatal("contacts_update tool not found")
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

func TestContactsDelete_RequiresID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{
		Contacts: &fastmail.ContactsClient{},
	})

	rt, ok := server.tools["contacts_delete"]
	if !ok {
		t.Fatal("contacts_delete tool not found")
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

func TestContactsUpdate_NilClient(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["contacts_update"]
	if !ok {
		t.Fatal("contacts_update tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{"id": "test-123"})
	if err == nil {
		t.Fatal("expected error for nil contacts client")
	}
	if err.Error() != "contacts client not configured" {
		t.Errorf("expected 'contacts client not configured', got %q", err.Error())
	}
}

func TestContactsDelete_NilClient(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["contacts_delete"]
	if !ok {
		t.Fatal("contacts_delete tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{"id": "test-123"})
	if err == nil {
		t.Fatal("expected error for nil contacts client")
	}
	if err.Error() != "contacts client not configured" {
		t.Errorf("expected 'contacts client not configured', got %q", err.Error())
	}
}

func TestMailFlagTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_flag"]
	if !ok {
		t.Fatal("mail_flag tool not found")
	}

	// Verify tool definition
	if rt.tool.Name != "mail_flag" {
		t.Errorf("expected name %q, got %q", "mail_flag", rt.tool.Name)
	}

	if rt.tool.Description != "Set or remove flags/keywords on an email" {
		t.Errorf("unexpected description: %s", rt.tool.Description)
	}

	// Verify input schema has expected properties
	if _, ok := rt.tool.InputSchema.Properties["id"]; !ok {
		t.Error("expected 'id' property in input schema")
	}
	if _, ok := rt.tool.InputSchema.Properties["keywords"]; !ok {
		t.Error("expected 'keywords' property in input schema")
	}

	// Verify required fields
	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("expected 'id' to be required")
	}
	if !slices.Contains(rt.tool.InputSchema.Required, "keywords") {
		t.Error("expected 'keywords' to be required")
	}
}

func TestMailFlagTool_RequiresID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_flag"]
	if !ok {
		t.Fatal("mail_flag tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{
		"keywords": map[string]any{"$seen": true},
	})
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	if err.Error() != "id is required" {
		t.Errorf("expected 'id is required', got %q", err.Error())
	}
}

func TestMailFlagTool_RequiresKeywords(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_flag"]
	if !ok {
		t.Fatal("mail_flag tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{
		"id": "email-123",
	})
	if err == nil {
		t.Fatal("expected error for missing keywords")
	}
	if err.Error() != "keywords is required" {
		t.Errorf("expected 'keywords is required', got %q", err.Error())
	}
}

func TestMailFlagTool_RejectsNonBooleanValues(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_flag"]
	if !ok {
		t.Fatal("mail_flag tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{
		"id":       "email-123",
		"keywords": map[string]any{"$seen": "yes"},
	})
	if err == nil {
		t.Fatal("expected error for non-boolean keyword value")
	}
}

func TestMailFlagTool_RejectsEmptyKeywords(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_flag"]
	if !ok {
		t.Fatal("mail_flag tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{
		"id":       "email-123",
		"keywords": map[string]any{},
	})
	if err == nil {
		t.Fatal("expected error for empty keywords")
	}
}

func TestMailThreadTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_thread"]
	if !ok {
		t.Fatal("mail_thread tool not found")
	}

	// Verify tool definition
	if rt.tool.Name != "mail_thread" {
		t.Errorf("expected name %q, got %q", "mail_thread", rt.tool.Name)
	}

	if rt.tool.Description != "Get all emails in a conversation thread" {
		t.Errorf("unexpected description: %s", rt.tool.Description)
	}

	// Verify input schema has expected properties
	if _, ok := rt.tool.InputSchema.Properties["thread_id"]; !ok {
		t.Error("expected 'thread_id' property in input schema")
	}

	// Verify required fields
	if !slices.Contains(rt.tool.InputSchema.Required, "thread_id") {
		t.Error("expected 'thread_id' to be required")
	}
}

func TestMailThreadTool_RequiresThreadID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_thread"]
	if !ok {
		t.Fatal("mail_thread tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing thread_id")
	}
	if err.Error() != "thread_id is required" {
		t.Errorf("expected 'thread_id is required', got %q", err.Error())
	}
}

func TestMailImportTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_import"]
	if !ok {
		t.Fatal("mail_import tool not found")
	}

	if rt.tool.Name != "mail_import" {
		t.Errorf("expected name %q, got %q", "mail_import", rt.tool.Name)
	}

	if rt.tool.Description != "Import an RFC 5322 email message into a mailbox" {
		t.Errorf("unexpected description: %s", rt.tool.Description)
	}

	// Verify input schema has expected properties
	if _, ok := rt.tool.InputSchema.Properties["content"]; !ok {
		t.Error("expected 'content' property in input schema")
	}
	if _, ok := rt.tool.InputSchema.Properties["folder"]; !ok {
		t.Error("expected 'folder' property in input schema")
	}
	if _, ok := rt.tool.InputSchema.Properties["seen"]; !ok {
		t.Error("expected 'seen' property in input schema")
	}
	if _, ok := rt.tool.InputSchema.Properties["flagged"]; !ok {
		t.Error("expected 'flagged' property in input schema")
	}

	// Verify required fields
	if !slices.Contains(rt.tool.InputSchema.Required, "content") {
		t.Error("expected 'content' to be required")
	}
}

func TestMailImportTool_RequiresContent(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_import"]
	if !ok {
		t.Fatal("mail_import tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing content")
	}
	if err.Error() != "content is required" {
		t.Errorf("expected 'content is required', got %q", err.Error())
	}
}

func TestMailSearchTool_HasSnippetsProperty(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_search"]
	if !ok {
		t.Fatal("mail_search tool not found")
	}

	if _, ok := rt.tool.InputSchema.Properties["snippets"]; !ok {
		t.Error("expected 'snippets' property in mail_search input schema")
	}
}

func TestMailSearchTool_SnippetsNotRequired(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_search"]
	if !ok {
		t.Fatal("mail_search tool not found")
	}

	// snippets should NOT be in the required list
	for _, req := range rt.tool.InputSchema.Required {
		if req == "snippets" {
			t.Error("'snippets' should not be required")
		}
	}
}

func TestMailAttachmentsTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_attachments"]
	if !ok {
		t.Fatal("mail_attachments tool not found")
	}

	if rt.tool.Name != "mail_attachments" {
		t.Errorf("expected name %q, got %q", "mail_attachments", rt.tool.Name)
	}

	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("expected 'id' to be required")
	}
}

func TestMailAttachmentsTool_RequiresID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_attachments"]
	if !ok {
		t.Fatal("mail_attachments tool not found")
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

func TestMailDownloadTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_download"]
	if !ok {
		t.Fatal("mail_download tool not found")
	}

	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("expected 'id' to be required")
	}
	if !slices.Contains(rt.tool.InputSchema.Required, "blob_id") {
		t.Error("expected 'blob_id' to be required")
	}
}

func TestMailDownloadTool_RequiresID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_download"]
	if !ok {
		t.Fatal("mail_download tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{"blob_id": "blob123"})
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	if err.Error() != "id is required" {
		t.Errorf("expected 'id is required', got %q", err.Error())
	}
}

func TestMailDownloadTool_RequiresBlobID(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_download"]
	if !ok {
		t.Fatal("mail_download tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{"id": "email123"})
	if err == nil {
		t.Fatal("expected error for missing blob_id")
	}
	if err.Error() != "blob_id is required" {
		t.Errorf("expected 'blob_id is required', got %q", err.Error())
	}
}

func TestMailUploadTool(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_upload"]
	if !ok {
		t.Fatal("mail_upload tool not found")
	}

	if !slices.Contains(rt.tool.InputSchema.Required, "content") {
		t.Error("expected 'content' to be required")
	}
	if !slices.Contains(rt.tool.InputSchema.Required, "content_type") {
		t.Error("expected 'content_type' to be required")
	}
	if !slices.Contains(rt.tool.InputSchema.Required, "filename") {
		t.Error("expected 'filename' to be required")
	}
}

func TestMailUploadTool_RequiresContent(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_upload"]
	if !ok {
		t.Fatal("mail_upload tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{
		"content_type": "application/pdf",
		"filename":     "test.pdf",
	})
	if err == nil {
		t.Fatal("expected error for missing content")
	}
	if err.Error() != "content is required" {
		t.Errorf("expected 'content is required', got %q", err.Error())
	}
}

func TestMailUploadTool_RequiresContentType(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	rt, ok := server.tools["mail_upload"]
	if !ok {
		t.Fatal("mail_upload tool not found")
	}

	ctx := context.Background()
	_, err := rt.handler(ctx, map[string]any{
		"content":  "SGVsbG8=",
		"filename": "test.txt",
	})
	if err == nil {
		t.Fatal("expected error for missing content_type")
	}
	if err.Error() != "content_type is required" {
		t.Errorf("expected 'content_type is required', got %q", err.Error())
	}
}

func TestBase64Helpers(t *testing.T) {
	original := []byte("Hello, World!")
	encoded := base64Encode(original)
	decoded, err := base64Decode(encoded)
	if err != nil {
		t.Fatalf("base64Decode error: %v", err)
	}
	if !bytes.Equal(decoded, original) {
		t.Errorf("expected %q, got %q", original, decoded)
	}
}

func TestSearchResultConversion(t *testing.T) {
	now := time.Now()
	results := []fastmail.SearchResult{
		{
			Email: fastmail.Email{
				ID:         "email-1",
				ThreadID:   "thread-1",
				Subject:    "Meeting Notes",
				Preview:    "Notes from today",
				ReceivedAt: now,
				Size:       1024,
				Keywords:   []string{"$seen"},
				MailboxIDs: []string{"inbox"},
			},
			SubjectSnippet: "<mark>Meeting</mark> Notes",
			PreviewSnippet: "<mark>Notes</mark> from today",
		},
		{
			Email: fastmail.Email{
				ID:         "email-2",
				ThreadID:   "thread-2",
				Subject:    "Other Email",
				Preview:    "Other content",
				ReceivedAt: now,
				Size:       512,
				Keywords:   []string{},
				MailboxIDs: []string{"inbox"},
			},
			SubjectSnippet: "",
			PreviewSnippet: "Other <mark>content</mark>",
		},
	}

	mcpResults := convertSearchResultsToMCP(results)

	if len(mcpResults) != 2 {
		t.Fatalf("expected 2 results, got %d", len(mcpResults))
	}

	if mcpResults[0].ID != "email-1" {
		t.Errorf("expected ID 'email-1', got %q", mcpResults[0].ID)
	}
	if mcpResults[0].SubjectSnippet != "<mark>Meeting</mark> Notes" {
		t.Errorf("expected SubjectSnippet '<mark>Meeting</mark> Notes', got %q", mcpResults[0].SubjectSnippet)
	}
	if mcpResults[0].PreviewSnippet != "<mark>Notes</mark> from today" {
		t.Errorf("expected PreviewSnippet '<mark>Notes</mark> from today', got %q", mcpResults[0].PreviewSnippet)
	}
	if !mcpResults[0].IsRead {
		t.Error("expected IsRead to be true")
	}

	if mcpResults[1].SubjectSnippet != "" {
		t.Errorf("expected empty SubjectSnippet, got %q", mcpResults[1].SubjectSnippet)
	}
	if mcpResults[1].PreviewSnippet != "Other <mark>content</mark>" {
		t.Errorf("expected PreviewSnippet 'Other <mark>content</mark>', got %q", mcpResults[1].PreviewSnippet)
	}
}

package mcp

import (
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
		"mailbox_list",
		"mailbox_create",
		"mailbox_rename",
		"mailbox_delete",
		"masked_email_list",
		"masked_email_create",
		"contacts_list",
		"contacts_get",
		"contacts_create",
		"contacts_update",
		"contacts_delete",
		"calendar_events",
		"calendar_create",
		"calendar_get",
		"calendar_update",
		"calendar_delete",
		"calendar_calendars",
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
	if _, ok := server.tools["contacts_update"]; !ok {
		t.Error("contacts_update tool not found")
	}
	if _, ok := server.tools["contacts_delete"]; !ok {
		t.Error("contacts_delete tool not found")
	}

	// Verify contacts_update requires id
	rt := server.tools["contacts_update"]
	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("contacts_update: expected 'id' to be required")
	}

	// Verify contacts_delete requires id
	rt = server.tools["contacts_delete"]
	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("contacts_delete: expected 'id' to be required")
	}
}

func TestContactsUpdateHandler(t *testing.T) {
	t.Run("nil contacts returns error", func(t *testing.T) {
		handler := makeContactsUpdateHandler(ToolsConfig{})
		_, err := handler(t.Context(), map[string]any{"id": "contact-1"})
		if err == nil {
			t.Error("expected error for nil contacts")
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		handler := makeContactsUpdateHandler(ToolsConfig{
			Contacts: fastmail.NewContactsClient("http://localhost", "user", "pass"),
		})
		_, err := handler(t.Context(), map[string]any{})
		if err == nil {
			t.Error("expected error for empty id")
		}
	})
}

func TestContactsDeleteHandler(t *testing.T) {
	t.Run("nil contacts returns error", func(t *testing.T) {
		handler := makeContactsDeleteHandler(ToolsConfig{})
		_, err := handler(t.Context(), map[string]any{"id": "contact-1"})
		if err == nil {
			t.Error("expected error for nil contacts")
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		handler := makeContactsDeleteHandler(ToolsConfig{
			Contacts: fastmail.NewContactsClient("http://localhost", "user", "pass"),
		})
		_, err := handler(t.Context(), map[string]any{})
		if err == nil {
			t.Error("expected error for empty id")
		}
	})
}

func TestCalendarTools(t *testing.T) {
	server := NewServer("test", "1.0")
	RegisterMailTools(server, ToolsConfig{})

	// Verify calendar tools
	calendarTools := []string{
		"calendar_events",
		"calendar_create",
		"calendar_get",
		"calendar_update",
		"calendar_delete",
		"calendar_calendars",
	}
	for _, name := range calendarTools {
		if _, ok := server.tools[name]; !ok {
			t.Errorf("%s tool not found", name)
		}
	}

	// Verify calendar_get requires id
	rt := server.tools["calendar_get"]
	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("calendar_get: expected 'id' to be required")
	}

	// Verify calendar_update requires id
	rt = server.tools["calendar_update"]
	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("calendar_update: expected 'id' to be required")
	}

	// Verify calendar_delete requires id
	rt = server.tools["calendar_delete"]
	if !slices.Contains(rt.tool.InputSchema.Required, "id") {
		t.Error("calendar_delete: expected 'id' to be required")
	}
}

func TestCalendarGetHandler(t *testing.T) {
	t.Run("nil calendar returns error", func(t *testing.T) {
		handler := makeCalendarGetHandler(ToolsConfig{})
		_, err := handler(t.Context(), map[string]any{"id": "evt-1"})
		if err == nil {
			t.Error("expected error for nil calendar")
		}
	})

	t.Run("empty id returns error", func(t *testing.T) {
		handler := makeCalendarGetHandler(ToolsConfig{
			Calendar: &CalendarAdapter{
				GetEventFunc: func(_ context.Context, _ string) (*fastmail.Event, error) {
					return &fastmail.Event{}, nil
				},
			},
		})
		_, err := handler(t.Context(), map[string]any{})
		if err == nil {
			t.Error("expected error for empty id")
		}
	})

	t.Run("returns event", func(t *testing.T) {
		now := time.Now()
		handler := makeCalendarGetHandler(ToolsConfig{
			Calendar: &CalendarAdapter{
				GetEventFunc: func(_ context.Context, id string) (*fastmail.Event, error) {
					return &fastmail.Event{
						ID:      id,
						Summary: "Test Event",
						Start:   now,
						End:     now.Add(time.Hour),
					}, nil
				},
			},
		})
		result, err := handler(t.Context(), map[string]any{"id": "evt-1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		evt, ok := result.(*Event)
		if !ok {
			t.Fatalf("expected *Event, got %T", result)
		}
		if evt.ID != "evt-1" {
			t.Errorf("expected ID 'evt-1', got %q", evt.ID)
		}
	})
}

func TestCalendarDeleteHandler(t *testing.T) {
	t.Run("nil calendar returns error", func(t *testing.T) {
		handler := makeCalendarDeleteHandler(ToolsConfig{})
		_, err := handler(t.Context(), map[string]any{"id": "evt-1"})
		if err == nil {
			t.Error("expected error for nil calendar")
		}
	})

	t.Run("deletes event", func(t *testing.T) {
		deletedID := ""
		handler := makeCalendarDeleteHandler(ToolsConfig{
			Calendar: &CalendarAdapter{
				DeleteEventFunc: func(_ context.Context, id string) error {
					deletedID = id
					return nil
				},
			},
		})
		result, err := handler(t.Context(), map[string]any{"id": "evt-1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if deletedID != "evt-1" {
			t.Errorf("expected delete of 'evt-1', got %q", deletedID)
		}
		m, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}
		if m["status"] != "deleted" {
			t.Errorf("expected status 'deleted', got %v", m["status"])
		}
	})
}

func TestCalendarUpdateHandler(t *testing.T) {
	t.Run("nil calendar returns error", func(t *testing.T) {
		handler := makeCalendarUpdateHandler(ToolsConfig{})
		_, err := handler(t.Context(), map[string]any{"id": "evt-1"})
		if err == nil {
			t.Error("expected error for nil calendar")
		}
	})

	t.Run("updates event fields", func(t *testing.T) {
		now := time.Now()
		var updatedEvent *fastmail.Event
		handler := makeCalendarUpdateHandler(ToolsConfig{
			Calendar: &CalendarAdapter{
				GetEventFunc: func(_ context.Context, id string) (*fastmail.Event, error) {
					return &fastmail.Event{
						ID:      id,
						Summary: "Original",
						Start:   now,
						End:     now.Add(time.Hour),
					}, nil
				},
				UpdateEventFunc: func(_ context.Context, event *fastmail.Event) error {
					updatedEvent = event
					return nil
				},
			},
		})
		result, err := handler(t.Context(), map[string]any{
			"id":      "evt-1",
			"summary": "Updated Title",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updatedEvent == nil {
			t.Fatal("UpdateEventFunc was not called")
		}
		if updatedEvent.Summary != "Updated Title" {
			t.Errorf("expected summary 'Updated Title', got %q", updatedEvent.Summary)
		}
		m, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}
		if m["status"] != "updated" {
			t.Errorf("expected status 'updated', got %v", m["status"])
		}
	})
}

func TestCalendarCalendarsHandler(t *testing.T) {
	t.Run("nil calendar returns error", func(t *testing.T) {
		handler := makeCalendarCalendarsHandler(ToolsConfig{})
		_, err := handler(t.Context(), nil)
		if err == nil {
			t.Error("expected error for nil calendar")
		}
	})

	t.Run("returns calendars", func(t *testing.T) {
		handler := makeCalendarCalendarsHandler(ToolsConfig{
			Calendar: &CalendarAdapter{
				ListCalendarsFunc: func(_ context.Context) ([]fastmail.Calendar, error) {
					return []fastmail.Calendar{
						{ID: "cal-1", Name: "Personal"},
						{ID: "cal-2", Name: "Work"},
					}, nil
				},
			},
		})
		result, err := handler(t.Context(), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		cals, ok := result.([]fastmail.Calendar)
		if !ok {
			t.Fatalf("expected []fastmail.Calendar, got %T", result)
		}
		if len(cals) != 2 {
			t.Errorf("expected 2 calendars, got %d", len(cals))
		}
	})
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

package mcp

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

		got, err := getStringArg(args, "name", "default")
		require.NoError(t, err)
		assert.Equal(t, "test", got)

		got, err = getStringArg(args, "missing", "default")
		require.NoError(t, err)
		assert.Equal(t, "default", got)
	})

	t.Run("getUint64Arg", func(t *testing.T) {
		args := map[string]any{
			"float":  float64(42),
			"int":    int(42),
			"int64":  int64(42),
			"uint64": uint64(42),
		}

		for _, key := range []string{"float", "int", "int64", "uint64"} {
			got, err := getUint64Arg(args, key, 0)
			require.NoError(t, err)
			assert.Equal(t, uint64(42), got, "getUint64Arg(%s)", key)
		}

		got, err := getUint64Arg(args, "missing", 10)
		require.NoError(t, err)
		assert.Equal(t, uint64(10), got)
	})

	t.Run("getBoolArg", func(t *testing.T) {
		args := map[string]any{
			"true":  true,
			"false": false,
		}

		got, err := getBoolArg(args, "true", false)
		require.NoError(t, err)
		assert.True(t, got)

		got, err = getBoolArg(args, "false", true)
		require.NoError(t, err)
		assert.False(t, got)

		got, err = getBoolArg(args, "missing", true)
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("getStringSliceArg", func(t *testing.T) {
		args := map[string]any{
			"strings": []string{"a", "b"},
			"anys":    []any{"c", "d"},
		}

		got, err := getStringSliceArg(args, "strings")
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, got)

		got, err = getStringSliceArg(args, "anys")
		require.NoError(t, err)
		assert.Equal(t, []string{"c", "d"}, got)

		got, err = getStringSliceArg(args, "missing")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestHelperFunctions_TypeMismatch(t *testing.T) {
	t.Run("getStringArg returns error on wrong type", func(t *testing.T) {
		args := map[string]any{"id": 12345}
		val, err := getStringArg(args, "id", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a string")
		assert.Contains(t, err.Error(), "id")
		assert.Empty(t, val)
	})

	t.Run("getStringArg returns default when key absent", func(t *testing.T) {
		args := map[string]any{}
		val, err := getStringArg(args, "id", "default")
		require.NoError(t, err)
		assert.Equal(t, "default", val)
	})

	t.Run("getStringArg returns value when type correct", func(t *testing.T) {
		args := map[string]any{"id": "abc"}
		val, err := getStringArg(args, "id", "")
		require.NoError(t, err)
		assert.Equal(t, "abc", val)
	})

	t.Run("getUint64Arg returns error on wrong type", func(t *testing.T) {
		args := map[string]any{"limit": "not-a-number"}
		val, err := getUint64Arg(args, "limit", 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a number")
		assert.Contains(t, err.Error(), "limit")
		assert.Equal(t, uint64(0), val)
	})

	t.Run("getUint64Arg returns default when key absent", func(t *testing.T) {
		args := map[string]any{}
		val, err := getUint64Arg(args, "limit", 10)
		require.NoError(t, err)
		assert.Equal(t, uint64(10), val)
	})

	t.Run("getUint64Arg returns value for float64", func(t *testing.T) {
		args := map[string]any{"limit": float64(42)}
		val, err := getUint64Arg(args, "limit", 0)
		require.NoError(t, err)
		assert.Equal(t, uint64(42), val)
	})

	t.Run("getUint64Arg returns error on negative float64", func(t *testing.T) {
		args := map[string]any{"limit": float64(-1)}
		val, err := getUint64Arg(args, "limit", 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be non-negative")
		assert.Equal(t, uint64(0), val)
	})

	t.Run("getBoolArg returns error on wrong type", func(t *testing.T) {
		args := map[string]any{"flag": "true"}
		val, err := getBoolArg(args, "flag", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be a boolean")
		assert.Contains(t, err.Error(), "flag")
		assert.False(t, val)
	})

	t.Run("getBoolArg returns default when key absent", func(t *testing.T) {
		args := map[string]any{}
		val, err := getBoolArg(args, "flag", true)
		require.NoError(t, err)
		assert.True(t, val)
	})

	t.Run("getBoolArg returns value when type correct", func(t *testing.T) {
		args := map[string]any{"flag": true}
		val, err := getBoolArg(args, "flag", false)
		require.NoError(t, err)
		assert.True(t, val)
	})

	t.Run("getStringSliceArg returns error on non-string elements", func(t *testing.T) {
		args := map[string]any{"tags": []any{"ok", 123, "also-ok"}}
		val, err := getStringSliceArg(args, "tags")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must contain only strings")
		assert.Nil(t, val)
	})

	t.Run("getStringSliceArg returns error on wrong type", func(t *testing.T) {
		args := map[string]any{"tags": "not-a-slice"}
		val, err := getStringSliceArg(args, "tags")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be an array")
		assert.Nil(t, val)
	})

	t.Run("getStringSliceArg returns nil when key absent", func(t *testing.T) {
		args := map[string]any{}
		val, err := getStringSliceArg(args, "tags")
		require.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("getStringSliceArg returns value for []string", func(t *testing.T) {
		args := map[string]any{"tags": []string{"a", "b"}}
		val, err := getStringSliceArg(args, "tags")
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, val)
	})

	t.Run("getStringSliceArg returns value for []any with strings", func(t *testing.T) {
		args := map[string]any{"tags": []any{"c", "d"}}
		val, err := getStringSliceArg(args, "tags")
		require.NoError(t, err)
		assert.Equal(t, []string{"c", "d"}, val)
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

func TestRegisterResources(t *testing.T) {
	server := NewServer("test", "1.0")
	registry := NewResourceRegistry(nil) // nil client is fine for registration test

	RegisterResources(server, registry)

	// Verify all resources from registry are registered with the server
	registryResources := registry.List()
	if len(registryResources) == 0 {
		t.Fatal("expected registry to have default resources")
	}

	server.mu.RLock()
	defer server.mu.RUnlock()

	for _, res := range registryResources {
		if _, ok := server.resources[res.URI]; !ok {
			t.Errorf("expected resource %q to be registered with server", res.URI)
		}
		if _, ok := server.readers[res.URI]; !ok {
			t.Errorf("expected reader for %q to be registered with server", res.URI)
		}
	}

	// Verify resource count matches
	if len(server.resources) != len(registryResources) {
		t.Errorf("expected %d resources, got %d", len(registryResources), len(server.resources))
	}
}

func TestRegisterResourcesWithCalendar(t *testing.T) {
	server := NewServer("test", "1.0")
	cal := &CalendarAdapter{
		ListCalendarsFunc: func(_ context.Context) ([]fastmail.Calendar, error) {
			return nil, nil
		},
	}
	registry := NewResourceRegistry(nil, WithCalendarAdapter(cal))

	RegisterResources(server, registry)

	server.mu.RLock()
	defer server.mu.RUnlock()

	// Should have all default resources including calendar ones
	if _, ok := server.resources["fastmail://calendars"]; !ok {
		t.Error("expected fastmail://calendars resource to be registered")
	}
}

func TestMailFlagHandler_MutualExclusivity(t *testing.T) {
	handler := makeMailFlagHandler(ToolsConfig{})

	t.Run("read and unread are mutually exclusive", func(t *testing.T) {
		_, err := handler(t.Context(), map[string]any{
			"id":     "msg-1",
			"read":   true,
			"unread": true,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutually exclusive")
	})

	t.Run("flagged and unflagged are mutually exclusive", func(t *testing.T) {
		_, err := handler(t.Context(), map[string]any{
			"id":        "msg-1",
			"flagged":   true,
			"unflagged": true,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutually exclusive")
	})

	t.Run("read alone passes validation", func(t *testing.T) {
		// With nil client, handler will panic past validation when calling SetKeywords.
		// A panic (not an error) confirms validation was passed successfully.
		assert.Panics(t, func() {
			_, _ = handler(t.Context(), map[string]any{
				"id":   "msg-1",
				"read": true,
			})
		})
	})

	t.Run("read and flagged together pass validation", func(t *testing.T) {
		// With nil client, handler will panic past validation when calling SetKeywords.
		// A panic (not an error) confirms validation was passed successfully.
		assert.Panics(t, func() {
			_, _ = handler(t.Context(), map[string]any{
				"id":      "msg-1",
				"read":    true,
				"flagged": true,
			})
		})
	})
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

// Package mcp provides MCP (Model Context Protocol) server implementation
// for exposing Fastmail functionality to AI agents.
package mcp

import (
	"context"
	"time"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// ToolsConfig holds clients needed for tool handlers.
type ToolsConfig struct {
	Client   *fastmail.Client
	Contacts *fastmail.ContactsClient
	Calendar *CalendarAdapter
}

// CalendarAdapter wraps calendar operations for MCP tools.
// This is needed because CalendarService requires a dav.Client.
type CalendarAdapter struct {
	ListEventsFunc    func(ctx context.Context, calendarID string, start, end time.Time) ([]fastmail.Event, error)
	CreateEventFunc   func(ctx context.Context, event *fastmail.Event) error
	GetEventFunc      func(ctx context.Context, eventID string) (*fastmail.Event, error)
	UpdateEventFunc   func(ctx context.Context, event *fastmail.Event) error
	DeleteEventFunc   func(ctx context.Context, eventID string) error
	ListCalendarsFunc func(ctx context.Context) ([]fastmail.Calendar, error)
}

// RegisterMailTools registers all mail-related tools on the server.
func RegisterMailTools(s *Server, cfg ToolsConfig) {
	registerMailTools(s, cfg)
	registerMailboxTools(s, cfg)
	registerMaskedEmailTools(s, cfg)
	registerContactTools(s, cfg)
	registerCalendarTools(s, cfg)
}

// registerMailTools registers email tools.
func registerMailTools(s *Server, cfg ToolsConfig) {
	// mail_list - List emails from folder
	s.RegisterTool(
		NewTool("mail_list", "List emails from a mailbox folder").
			WithProperty("folder", "string", "Mailbox folder name (e.g., 'Inbox', 'Sent', 'Archive')").
			WithProperty("limit", "integer", "Maximum number of emails to return (default: 10)"),
		makeMailListHandler(cfg),
	)

	// mail_get - Get email by ID
	s.RegisterTool(
		NewTool("mail_get", "Get a single email by its ID").
			WithProperty("id", "string", "The email ID").
			WithRequired("id"),
		makeMailGetHandler(cfg),
	)

	// mail_search - Search emails
	s.RegisterTool(
		NewTool("mail_search", "Search emails by text query").
			WithProperty("query", "string", "Search query (e.g., 'from:alice subject:meeting')").
			WithProperty("limit", "integer", "Maximum number of results (default: 10)").
			WithRequired("query"),
		makeMailSearchHandler(cfg),
	)

	// mail_send - Send email
	s.RegisterTool(
		NewTool("mail_send", "Compose and send a new email").
			WithProperty("to", "array", "Recipient email addresses").
			WithProperty("cc", "array", "CC recipient addresses (optional)").
			WithProperty("bcc", "array", "BCC recipient addresses (optional)").
			WithProperty("subject", "string", "Email subject").
			WithProperty("body", "string", "Email body text").
			WithRequired("to", "subject", "body"),
		makeMailSendHandler(cfg),
	)

	// mail_reply - Reply to email
	s.RegisterTool(
		NewTool("mail_reply", "Send a reply to an existing email").
			WithProperty("email_id", "string", "ID of the email to reply to").
			WithProperty("body", "string", "Reply body text").
			WithProperty("reply_all", "boolean", "Reply to all recipients (default: false)").
			WithRequired("email_id", "body"),
		makeMailReplyHandler(cfg),
	)

	// mail_move - Move email to folder
	s.RegisterTool(
		NewTool("mail_move", "Move an email to a different folder").
			WithProperty("id", "string", "The email ID").
			WithProperty("folder", "string", "Target folder name").
			WithRequired("id", "folder"),
		makeMailMoveHandler(cfg),
	)

	// mail_flag - Set email flags
	s.RegisterTool(
		NewTool("mail_flag", "Set or remove email flags (read, flagged)").
			WithProperty("id", "string", "The email ID").
			WithProperty("read", "boolean", "Mark as read").
			WithProperty("unread", "boolean", "Mark as unread").
			WithProperty("flagged", "boolean", "Mark as flagged").
			WithProperty("unflagged", "boolean", "Remove flagged").
			WithRequired("id"),
		makeMailFlagHandler(cfg),
	)

	// mail_delete - Delete email
	s.RegisterTool(
		NewTool("mail_delete", "Delete an email (moves to Trash, or permanently deletes if already in Trash)").
			WithProperty("id", "string", "The email ID").
			WithRequired("id"),
		makeMailDeleteHandler(cfg),
	)
}

// Mail tool handlers

func makeMailListHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		folder := getStringArg(args, "folder", "Inbox")
		limit := getUint64Arg(args, "limit", 10)

		emails, err := cfg.Client.Mail().List(ctx, folder, limit)
		if err != nil {
			return nil, oops.Wrapf(err, "listing emails")
		}

		return convertEmailsToMCP(emails), nil
	}
}

func makeMailGetHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		email, err := cfg.Client.Mail().Get(ctx, id)
		if err != nil {
			return nil, oops.Wrapf(err, "getting email")
		}

		return convertEmailToMCP(email), nil
	}
}

func makeMailSearchHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		query := getStringArg(args, "query", "")
		if query == "" {
			return nil, oops.Errorf("query is required")
		}
		limit := getUint64Arg(args, "limit", 10)

		emails, err := cfg.Client.Mail().Search(ctx, query, limit)
		if err != nil {
			return nil, oops.Wrapf(err, "searching emails")
		}

		return convertEmailsToMCP(emails), nil
	}
}

func makeMailSendHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		to := getStringSliceArg(args, "to")
		if len(to) == 0 {
			return nil, oops.Errorf("to is required")
		}
		subject := getStringArg(args, "subject", "")
		if subject == "" {
			return nil, oops.Errorf("subject is required")
		}
		body := getStringArg(args, "body", "")
		if body == "" {
			return nil, oops.Errorf("body is required")
		}

		opts := fastmail.SendOptions{
			To:      parseAddresses(to),
			Cc:      parseAddresses(getStringSliceArg(args, "cc")),
			Bcc:     parseAddresses(getStringSliceArg(args, "bcc")),
			Subject: subject,
			Body:    body,
		}

		emailID, err := cfg.Client.Mail().Send(ctx, opts)
		if err != nil {
			return nil, oops.Wrapf(err, "sending email")
		}

		return map[string]any{
			"id":     emailID,
			"status": "sent",
		}, nil
	}
}

func makeMailReplyHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		emailID := getStringArg(args, "email_id", "")
		if emailID == "" {
			return nil, oops.Errorf("email_id is required")
		}
		body := getStringArg(args, "body", "")
		if body == "" {
			return nil, oops.Errorf("body is required")
		}
		replyAll := getBoolArg(args, "reply_all", false)

		opts := fastmail.ReplyOptions{
			EmailID:  emailID,
			Body:     body,
			ReplyAll: replyAll,
		}

		replyID, err := cfg.Client.Mail().Reply(ctx, opts)
		if err != nil {
			return nil, oops.Wrapf(err, "sending reply")
		}

		return map[string]any{
			"id":     replyID,
			"status": "sent",
		}, nil
	}
}

func makeMailMoveHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}
		folder := getStringArg(args, "folder", "")
		if folder == "" {
			return nil, oops.Errorf("folder is required")
		}

		if err := cfg.Client.Mail().Move(ctx, id, folder); err != nil {
			return nil, oops.Wrapf(err, "moving email")
		}

		return map[string]any{
			"id":     id,
			"folder": folder,
			"status": "moved",
		}, nil
	}
}

func makeMailFlagHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		keywords := make(map[string]bool)
		if getBoolArg(args, "read", false) {
			keywords["$seen"] = true
		}
		if getBoolArg(args, "unread", false) {
			keywords["$seen"] = false
		}
		if getBoolArg(args, "flagged", false) {
			keywords["$flagged"] = true
		}
		if getBoolArg(args, "unflagged", false) {
			keywords["$flagged"] = false
		}

		if len(keywords) == 0 {
			return nil, oops.Errorf("at least one flag (read, unread, flagged, unflagged) is required")
		}

		if err := cfg.Client.Mail().SetKeywords(ctx, id, keywords); err != nil {
			return nil, oops.Wrapf(err, "setting flags")
		}

		return map[string]any{
			"id":     id,
			"status": "updated",
		}, nil
	}
}

func makeMailDeleteHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		if err := cfg.Client.Mail().Delete(ctx, id); err != nil {
			return nil, oops.Wrapf(err, "deleting email")
		}

		return map[string]any{
			"id":     id,
			"status": "deleted",
		}, nil
	}
}

// Masked email tools

func registerMaskedEmailTools(s *Server, cfg ToolsConfig) {
	// masked_email_list - List masked emails
	s.RegisterTool(
		NewTool("masked_email_list", "List all masked email addresses"),
		makeMaskedEmailListHandler(cfg),
	)

	// masked_email_create - Create masked email
	s.RegisterTool(
		NewTool("masked_email_create", "Create a new masked email address").
			WithProperty("domain", "string", "Domain to associate with the masked email (optional)").
			WithProperty("description", "string", "Description for the masked email (optional)"),
		makeMaskedEmailCreateHandler(cfg),
	)
}

func makeMaskedEmailListHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, _ map[string]any) (any, error) {
		emails, err := cfg.Client.MaskedEmail().List(ctx)
		if err != nil {
			return nil, oops.Wrapf(err, "listing masked emails")
		}

		result := make([]map[string]any, len(emails))
		for i, e := range emails {
			result[i] = map[string]any{
				"id":              e.ID,
				"email":           e.Email,
				"state":           string(e.State),
				"for_domain":      e.ForDomain,
				"description":     e.Description,
				"created_at":      e.CreatedAt,
				"last_message_at": e.LastMessageAt,
			}
		}
		return result, nil
	}
}

func makeMaskedEmailCreateHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		opts := fastmail.CreateMaskedEmailOptions{
			ForDomain:   getStringArg(args, "domain", ""),
			Description: getStringArg(args, "description", ""),
		}

		email, err := cfg.Client.MaskedEmail().Create(ctx, opts)
		if err != nil {
			return nil, oops.Wrapf(err, "creating masked email")
		}

		return map[string]any{
			"id":          email.ID,
			"email":       email.Email,
			"state":       string(email.State),
			"for_domain":  email.ForDomain,
			"description": email.Description,
		}, nil
	}
}

// Contact tools

func registerContactTools(s *Server, cfg ToolsConfig) {
	// contacts_list - List contacts
	s.RegisterTool(
		NewTool("contacts_list", "List all contacts from the address book"),
		makeContactsListHandler(cfg),
	)

	// contacts_get - Get contact by ID
	s.RegisterTool(
		NewTool("contacts_get", "Get a single contact by ID").
			WithProperty("id", "string", "The contact ID").
			WithRequired("id"),
		makeContactsGetHandler(cfg),
	)

	// contacts_create - Create contact
	s.RegisterTool(
		NewTool("contacts_create", "Create a new contact").
			WithProperty("name", "string", "Full name").
			WithProperty("email", "string", "Email address").
			WithProperty("phone", "string", "Phone number (optional)").
			WithRequired("name"),
		makeContactsCreateHandler(cfg),
	)
}

func makeContactsListHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, _ map[string]any) (any, error) {
		if cfg.Contacts == nil {
			return nil, oops.Errorf("contacts client not configured")
		}

		contacts, err := cfg.Contacts.List(ctx)
		if err != nil {
			return nil, oops.Wrapf(err, "listing contacts")
		}

		result := make([]map[string]any, len(contacts))
		for i, c := range contacts {
			result[i] = map[string]any{
				"id":    c.ID,
				"name":  c.Name,
				"email": c.Email,
				"phone": c.Phone,
			}
		}
		return result, nil
	}
}

func makeContactsGetHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		if cfg.Contacts == nil {
			return nil, oops.Errorf("contacts client not configured")
		}

		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		contact, err := cfg.Contacts.Get(ctx, id)
		if err != nil {
			return nil, oops.Wrapf(err, "getting contact")
		}

		return map[string]any{
			"id":    contact.ID,
			"name":  contact.Name,
			"email": contact.Email,
			"phone": contact.Phone,
		}, nil
	}
}

func makeContactsCreateHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		if cfg.Contacts == nil {
			return nil, oops.Errorf("contacts client not configured")
		}

		name := getStringArg(args, "name", "")
		if name == "" {
			return nil, oops.Errorf("name is required")
		}

		contact := &fastmail.Contact{
			Name:  name,
			Email: getStringArg(args, "email", ""),
			Phone: getStringArg(args, "phone", ""),
		}

		if err := cfg.Contacts.Create(ctx, contact); err != nil {
			return nil, oops.Wrapf(err, "creating contact")
		}

		return map[string]any{
			"id":     contact.ID,
			"name":   contact.Name,
			"email":  contact.Email,
			"phone":  contact.Phone,
			"status": "created",
		}, nil
	}
}

// Calendar tools

func registerCalendarTools(s *Server, cfg ToolsConfig) {
	// calendar_events - List calendar events
	s.RegisterTool(
		NewTool("calendar_events", "List calendar events within a date range").
			WithProperty("calendar_id", "string", "Calendar ID (optional, uses default if not specified)").
			WithProperty("start", "string", "Start date/time in RFC3339 format (e.g., '2024-01-01T00:00:00Z')").
			WithProperty("end", "string", "End date/time in RFC3339 format").
			WithRequired("start", "end"),
		makeCalendarEventsHandler(cfg),
	)

	// calendar_create - Create calendar event
	s.RegisterTool(
		NewTool("calendar_create", "Create a new calendar event").
			WithProperty("calendar_id", "string", "Calendar ID").
			WithProperty("summary", "string", "Event title/summary").
			WithProperty("description", "string", "Event description (optional)").
			WithProperty("location", "string", "Event location (optional)").
			WithProperty("start", "string", "Start date/time in RFC3339 format").
			WithProperty("end", "string", "End date/time in RFC3339 format").
			WithProperty("all_day", "boolean", "Whether this is an all-day event (default: false)").
			WithRequired("calendar_id", "summary", "start", "end"),
		makeCalendarCreateHandler(cfg),
	)

	// calendar_get - Get calendar event by ID
	s.RegisterTool(
		NewTool("calendar_get", "Get a calendar event by ID").
			WithProperty("id", "string", "Event ID").
			WithRequired("id"),
		makeCalendarGetHandler(cfg),
	)

	// calendar_update - Update calendar event
	s.RegisterTool(
		NewTool("calendar_update", "Update a calendar event").
			WithProperty("id", "string", "Event ID").
			WithProperty("summary", "string", "New event title").
			WithProperty("start", "string", "New start time (RFC3339)").
			WithProperty("end", "string", "New end time (RFC3339)").
			WithProperty("location", "string", "New location").
			WithProperty("description", "string", "New description").
			WithRequired("id"),
		makeCalendarUpdateHandler(cfg),
	)

	// calendar_delete - Delete calendar event
	s.RegisterTool(
		NewTool("calendar_delete", "Delete a calendar event").
			WithProperty("id", "string", "Event ID").
			WithRequired("id"),
		makeCalendarDeleteHandler(cfg),
	)

	// calendar_calendars - List all calendars
	s.RegisterTool(
		NewTool("calendar_calendars", "List all available calendars"),
		makeCalendarCalendarsHandler(cfg),
	)
}

func makeCalendarEventsHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		if cfg.Calendar == nil || cfg.Calendar.ListEventsFunc == nil {
			return nil, oops.Errorf("calendar not configured")
		}

		startStr := getStringArg(args, "start", "")
		if startStr == "" {
			return nil, oops.Errorf("start is required")
		}
		endStr := getStringArg(args, "end", "")
		if endStr == "" {
			return nil, oops.Errorf("end is required")
		}

		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return nil, oops.Wrapf(err, "parsing start time")
		}
		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return nil, oops.Wrapf(err, "parsing end time")
		}

		calendarID := getStringArg(args, "calendar_id", "")

		events, err := cfg.Calendar.ListEventsFunc(ctx, calendarID, start, end)
		if err != nil {
			return nil, oops.Wrapf(err, "listing events")
		}

		return convertEventsToMCP(events), nil
	}
}

func makeCalendarCreateHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		if cfg.Calendar == nil || cfg.Calendar.CreateEventFunc == nil {
			return nil, oops.Errorf("calendar not configured")
		}

		calendarID := getStringArg(args, "calendar_id", "")
		if calendarID == "" {
			return nil, oops.Errorf("calendar_id is required")
		}
		summary := getStringArg(args, "summary", "")
		if summary == "" {
			return nil, oops.Errorf("summary is required")
		}
		startStr := getStringArg(args, "start", "")
		if startStr == "" {
			return nil, oops.Errorf("start is required")
		}
		endStr := getStringArg(args, "end", "")
		if endStr == "" {
			return nil, oops.Errorf("end is required")
		}

		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return nil, oops.Wrapf(err, "parsing start time")
		}
		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return nil, oops.Wrapf(err, "parsing end time")
		}

		event := &fastmail.Event{
			CalendarID:  calendarID,
			Summary:     summary,
			Description: getStringArg(args, "description", ""),
			Location:    getStringArg(args, "location", ""),
			Start:       start,
			End:         end,
			AllDay:      getBoolArg(args, "all_day", false),
		}

		if err := cfg.Calendar.CreateEventFunc(ctx, event); err != nil {
			return nil, oops.Wrapf(err, "creating event")
		}

		return map[string]any{
			"id":          event.ID,
			"calendar_id": event.CalendarID,
			"summary":     event.Summary,
			"start":       event.Start.Format(time.RFC3339),
			"end":         event.End.Format(time.RFC3339),
			"status":      "created",
		}, nil
	}
}

func makeCalendarGetHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		if cfg.Calendar == nil || cfg.Calendar.GetEventFunc == nil {
			return nil, oops.Errorf("calendar not configured")
		}

		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		event, err := cfg.Calendar.GetEventFunc(ctx, id)
		if err != nil {
			return nil, oops.Wrapf(err, "getting event")
		}

		return convertEventToMCP(event), nil
	}
}

func makeCalendarUpdateHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		if cfg.Calendar == nil || cfg.Calendar.GetEventFunc == nil || cfg.Calendar.UpdateEventFunc == nil {
			return nil, oops.Errorf("calendar not configured")
		}

		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		// Fetch existing event
		event, err := cfg.Calendar.GetEventFunc(ctx, id)
		if err != nil {
			return nil, oops.Wrapf(err, "getting event")
		}

		// Apply updates
		if s := getStringArg(args, "summary", ""); s != "" {
			event.Summary = s
		}
		if s := getStringArg(args, "location", ""); s != "" {
			event.Location = s
		}
		if s := getStringArg(args, "description", ""); s != "" {
			event.Description = s
		}
		if s := getStringArg(args, "start", ""); s != "" {
			t, parseErr := time.Parse(time.RFC3339, s)
			if parseErr != nil {
				return nil, oops.Wrapf(parseErr, "invalid start time")
			}
			event.Start = t
		}
		if s := getStringArg(args, "end", ""); s != "" {
			t, parseErr := time.Parse(time.RFC3339, s)
			if parseErr != nil {
				return nil, oops.Wrapf(parseErr, "invalid end time")
			}
			event.End = t
		}

		if err := cfg.Calendar.UpdateEventFunc(ctx, event); err != nil {
			return nil, oops.Wrapf(err, "updating event")
		}

		return map[string]any{"id": id, "status": "updated"}, nil
	}
}

func makeCalendarDeleteHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]any) (any, error) {
		if cfg.Calendar == nil || cfg.Calendar.DeleteEventFunc == nil {
			return nil, oops.Errorf("calendar not configured")
		}

		id := getStringArg(args, "id", "")
		if id == "" {
			return nil, oops.Errorf("id is required")
		}

		if err := cfg.Calendar.DeleteEventFunc(ctx, id); err != nil {
			return nil, oops.Wrapf(err, "deleting event")
		}

		return map[string]any{"id": id, "status": "deleted"}, nil
	}
}

func makeCalendarCalendarsHandler(cfg ToolsConfig) ToolHandler {
	return func(ctx context.Context, _ map[string]any) (any, error) {
		if cfg.Calendar == nil || cfg.Calendar.ListCalendarsFunc == nil {
			return nil, oops.Errorf("calendar not configured")
		}

		return cfg.Calendar.ListCalendarsFunc(ctx)
	}
}

// Event is the MCP representation of a calendar event.
type Event struct {
	ID          string `json:"id"`
	CalendarID  string `json:"calendar_id"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	Start       string `json:"start"`
	End         string `json:"end"`
	AllDay      bool   `json:"all_day"`
	Status      string `json:"status,omitempty"`
}

func convertEventToMCP(e *fastmail.Event) *Event {
	return &Event{
		ID:          e.ID,
		CalendarID:  e.CalendarID,
		Summary:     e.Summary,
		Description: e.Description,
		Location:    e.Location,
		Start:       e.Start.Format(time.RFC3339),
		End:         e.End.Format(time.RFC3339),
		AllDay:      e.AllDay,
		Status:      e.Status,
	}
}

func convertEventsToMCP(events []fastmail.Event) []*Event {
	result := make([]*Event, len(events))
	for i := range events {
		result[i] = convertEventToMCP(&events[i])
	}
	return result
}

// Helper functions

func getStringArg(args map[string]any, key, defaultValue string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}

func getUint64Arg(args map[string]any, key string, defaultValue uint64) uint64 {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			if n >= 0 {
				return uint64(n)
			}
		case int:
			if n >= 0 {
				return uint64(n)
			}
		case int64:
			if n >= 0 {
				return uint64(n)
			}
		case uint64:
			return n
		}
	}
	return defaultValue
}

func getBoolArg(args map[string]any, key string, defaultValue bool) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getStringSliceArg(args map[string]any, key string) []string {
	if v, ok := args[key]; ok {
		switch arr := v.(type) {
		case []string:
			return arr
		case []any:
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

func parseAddresses(addrs []string) []fastmail.EmailAddress {
	result := make([]fastmail.EmailAddress, len(addrs))
	for i, addr := range addrs {
		result[i] = fastmail.EmailAddress{Email: addr}
	}
	return result
}

// Email is the MCP representation of an email.
type Email struct {
	ID         string   `json:"id"`
	ThreadID   string   `json:"thread_id"`
	Subject    string   `json:"subject"`
	Preview    string   `json:"preview"`
	ReceivedAt string   `json:"received_at"`
	Size       uint64   `json:"size"`
	IsRead     bool     `json:"is_read"`
	IsFlagged  bool     `json:"is_flagged"`
	Folders    []string `json:"folders"`
}

func convertEmailToMCP(e *fastmail.Email) *Email {
	return &Email{
		ID:         e.ID,
		ThreadID:   e.ThreadID,
		Subject:    e.Subject,
		Preview:    e.Preview,
		ReceivedAt: e.ReceivedAt.Format(time.RFC3339),
		Size:       e.Size,
		IsRead:     e.IsRead(),
		IsFlagged:  e.IsFlagged(),
		Folders:    e.MailboxIDs,
	}
}

func convertEmailsToMCP(emails []fastmail.Email) []*Email {
	result := make([]*Email, len(emails))
	for i := range emails {
		result[i] = convertEmailToMCP(&emails[i])
	}
	return result
}

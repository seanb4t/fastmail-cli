// Package mcp provides MCP (Model Context Protocol) server implementation
// for exposing Fastmail functionality to AI agents.
package mcp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/samber/oops"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// ResourceHandler fetches content for a resource.
// The params map contains captured URI path parameters.
type ResourceHandler func(ctx context.Context, params map[string]string) (*ResourceContent, error)

// ResourceTemplate represents a parameterized resource URI pattern.
type ResourceTemplate struct {
	URITemplate string `json:"uriTemplate"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// registeredResource holds a resource definition and its handler.
type registeredResource struct {
	resource Resource
	pattern  *regexp.Regexp
	handler  ResourceHandler
}

// ResourceRegistry manages MCP resources.
type ResourceRegistry struct {
	client    *fastmail.Client
	calendar  *CalendarAdapter
	resources []registeredResource
	templates []ResourceTemplate
}

// ResourceRegistryOption configures a ResourceRegistry.
type ResourceRegistryOption func(*ResourceRegistry)

// WithCalendarAdapter sets the calendar adapter for calendar resources.
func WithCalendarAdapter(cal *CalendarAdapter) ResourceRegistryOption {
	return func(r *ResourceRegistry) {
		r.calendar = cal
	}
}

// NewResourceRegistry creates a new resource registry with default resources.
func NewResourceRegistry(client *fastmail.Client, opts ...ResourceRegistryOption) *ResourceRegistry {
	r := &ResourceRegistry{
		client: client,
	}
	for _, opt := range opts {
		opt(r)
	}
	r.registerDefaults()
	return r
}

// List returns all registered resources.
func (r *ResourceRegistry) List() []Resource {
	result := make([]Resource, len(r.resources))
	for i, rr := range r.resources {
		result[i] = rr.resource
	}
	return result
}

// Templates returns all registered resource templates.
func (r *ResourceRegistry) Templates() []ResourceTemplate {
	return r.templates
}

// Read fetches the content of a resource by URI.
func (r *ResourceRegistry) Read(ctx context.Context, uri string) (*ResourceContent, error) {
	for _, rr := range r.resources {
		if rr.pattern != nil {
			matches := rr.pattern.FindStringSubmatch(uri)
			if matches != nil {
				params := extractParams(rr.pattern, matches)
				return rr.handler(ctx, params)
			}
		} else if rr.resource.URI == uri {
			return rr.handler(ctx, nil)
		}
	}
	return nil, oops.Errorf("resource not found: %s", uri)
}

// register adds a static resource to the registry.
func (r *ResourceRegistry) register(res Resource, handler ResourceHandler) {
	r.resources = append(r.resources, registeredResource{
		resource: res,
		handler:  handler,
	})
}

// registerTemplate adds a parameterized resource template to the registry.
func (r *ResourceRegistry) registerTemplate(template ResourceTemplate, pattern *regexp.Regexp, handler ResourceHandler) {
	r.templates = append(r.templates, template)
	r.resources = append(r.resources, registeredResource{
		resource: Resource{
			URI:         template.URITemplate,
			Name:        template.Name,
			Description: template.Description,
			MimeType:    template.MimeType,
		},
		pattern: pattern,
		handler: handler,
	})
}

// extractParams extracts named groups from a regexp match.
func extractParams(pattern *regexp.Regexp, matches []string) map[string]string {
	params := make(map[string]string)
	for i, name := range pattern.SubexpNames() {
		if i > 0 && name != "" && i < len(matches) {
			params[name] = matches[i]
		}
	}
	return params
}

// registerDefaults registers all default Fastmail resources.
func (r *ResourceRegistry) registerDefaults() {
	// fastmail://inbox - Recent inbox messages
	r.register(
		*NewResource("fastmail://inbox", "Recent Inbox").
			WithDescription("Recent emails from the inbox folder").
			WithMimeType("text/plain"),
		r.handleInbox,
	)

	// fastmail://mail/{id} - Single email content (template)
	r.registerTemplate(
		ResourceTemplate{
			URITemplate: "fastmail://mail/{id}",
			Name:        "Email Message",
			Description: "Content of a specific email message",
			MimeType:    "text/plain",
		},
		regexp.MustCompile(`^fastmail://mail/(?P<id>[^/]+)$`),
		r.handleMail,
	)

	// fastmail://contacts - Contact list
	r.register(
		*NewResource("fastmail://contacts", "Contacts").
			WithDescription("List of contacts from the address book").
			WithMimeType("text/plain"),
		r.handleContacts,
	)

	// fastmail://calendar/today - Today's events
	r.register(
		*NewResource("fastmail://calendar/today", "Today's Events").
			WithDescription("Calendar events for today").
			WithMimeType("text/plain"),
		r.handleCalendarToday,
	)

	// fastmail://masked-emails - Masked email list
	r.register(
		*NewResource("fastmail://masked-emails", "Masked Emails").
			WithDescription("List of masked email addresses").
			WithMimeType("text/plain"),
		r.handleMaskedEmails,
	)

	// fastmail://mailboxes - Mailbox list
	r.register(
		*NewResource("fastmail://mailboxes", "Mailboxes").
			WithDescription("List of email mailboxes (folders) with message counts").
			WithMimeType("text/plain"),
		r.handleMailboxes,
	)

	// fastmail://identities - Identity list
	r.register(
		*NewResource("fastmail://identities", "Identities").
			WithDescription("List of sender identities").
			WithMimeType("text/plain"),
		r.handleIdentities,
	)

	// fastmail://vacation - Vacation status
	r.register(
		*NewResource("fastmail://vacation", "Vacation Status").
			WithDescription("Current vacation/auto-reply settings").
			WithMimeType("text/plain"),
		r.handleVacation,
	)

	// fastmail://calendars - Calendar list
	r.register(
		*NewResource("fastmail://calendars", "Calendars").
			WithDescription("List of calendars").
			WithMimeType("text/plain"),
		r.handleCalendars,
	)
}

// handleInbox returns recent inbox messages.
func (r *ResourceRegistry) handleInbox(ctx context.Context, _ map[string]string) (*ResourceContent, error) {
	if r.client == nil {
		return nil, oops.Errorf("client not configured")
	}

	emails, err := r.client.Mail().List(ctx, "Inbox", 10)
	if err != nil {
		return nil, oops.Wrapf(err, "listing inbox")
	}

	content := formatEmailList(emails)
	return &ResourceContent{
		URI:      "fastmail://inbox",
		MimeType: "text/plain",
		Text:     content,
	}, nil
}

// handleMail returns a single email by ID.
func (r *ResourceRegistry) handleMail(ctx context.Context, params map[string]string) (*ResourceContent, error) {
	if r.client == nil {
		return nil, oops.Errorf("client not configured")
	}

	id := params["id"]
	if id == "" {
		return nil, oops.Errorf("email ID required")
	}

	email, err := r.client.Mail().Get(ctx, id)
	if err != nil {
		return nil, oops.Wrapf(err, "getting email %s", id)
	}

	content := formatEmail(email)
	return &ResourceContent{
		URI:      fmt.Sprintf("fastmail://mail/%s", id),
		MimeType: "text/plain",
		Text:     content,
	}, nil
}

// handleContacts returns the contact list.
func (r *ResourceRegistry) handleContacts(_ context.Context, _ map[string]string) (*ResourceContent, error) {
	// Contacts require CardDAV client which is not configured via JMAP
	return nil, oops.Errorf("contacts access requires CardDAV configuration")
}

// handleCalendarToday returns today's calendar events.
func (r *ResourceRegistry) handleCalendarToday(_ context.Context, _ map[string]string) (*ResourceContent, error) {
	// Calendar service requires DAV client which is not configured
	return nil, oops.Errorf("calendar access requires CalDAV configuration")
}

// handleMaskedEmails returns the masked email list.
func (r *ResourceRegistry) handleMaskedEmails(ctx context.Context, _ map[string]string) (*ResourceContent, error) {
	if r.client == nil {
		return nil, oops.Errorf("client not configured")
	}

	maskedEmails, err := r.client.MaskedEmail().List(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "listing masked emails")
	}

	content := formatMaskedEmailList(maskedEmails)
	return &ResourceContent{
		URI:      "fastmail://masked-emails",
		MimeType: "text/plain",
		Text:     content,
	}, nil
}

// handleMailboxes returns the mailbox list.
func (r *ResourceRegistry) handleMailboxes(ctx context.Context, _ map[string]string) (*ResourceContent, error) {
	if r.client == nil {
		return nil, oops.Errorf("client not configured")
	}

	mailboxes, err := r.client.Mailbox().List(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "listing mailboxes")
	}

	content := formatMailboxList(mailboxes)
	return &ResourceContent{
		URI:      "fastmail://mailboxes",
		MimeType: "text/plain",
		Text:     content,
	}, nil
}

// handleIdentities returns the identity list.
func (r *ResourceRegistry) handleIdentities(ctx context.Context, _ map[string]string) (*ResourceContent, error) {
	if r.client == nil {
		return nil, oops.Errorf("client not configured")
	}

	identities, err := r.client.Identity().List(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "listing identities")
	}

	content := formatIdentityList(identities)
	return &ResourceContent{
		URI:      "fastmail://identities",
		MimeType: "text/plain",
		Text:     content,
	}, nil
}

// handleVacation returns the vacation status.
func (r *ResourceRegistry) handleVacation(ctx context.Context, _ map[string]string) (*ResourceContent, error) {
	if r.client == nil {
		return nil, oops.Errorf("client not configured")
	}

	vacation, err := r.client.Vacation().Get(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "getting vacation status")
	}

	content := formatVacation(vacation)
	return &ResourceContent{
		URI:      "fastmail://vacation",
		MimeType: "text/plain",
		Text:     content,
	}, nil
}

// handleCalendars returns the calendar list.
func (r *ResourceRegistry) handleCalendars(ctx context.Context, _ map[string]string) (*ResourceContent, error) {
	if r.calendar == nil || r.calendar.ListCalendarsFunc == nil {
		return &ResourceContent{
			URI:      "fastmail://calendars",
			MimeType: "text/plain",
			Text:     "Calendar access requires CalDAV configuration. Use the calendar_list tool instead.",
		}, nil
	}

	calendars, err := r.calendar.ListCalendarsFunc(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "listing calendars")
	}

	content := formatCalendarList(calendars)
	return &ResourceContent{
		URI:      "fastmail://calendars",
		MimeType: "text/plain",
		Text:     content,
	}, nil
}

// formatEmailList formats a list of emails for LLM readability.
func formatEmailList(emails []fastmail.Email) string {
	if len(emails) == 0 {
		return "No emails found."
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Inbox (%d messages)\n\n", len(emails)))

	for i, email := range emails {
		b.WriteString(fmt.Sprintf("## %d. %s\n", i+1, email.Subject))
		b.WriteString(fmt.Sprintf("- ID: %s\n", email.ID))
		if !email.ReceivedAt.IsZero() {
			b.WriteString(fmt.Sprintf("- Received: %s\n", email.ReceivedAt.Format(time.RFC3339)))
		}
		if email.Preview != "" {
			b.WriteString(fmt.Sprintf("- Preview: %s\n", truncate(email.Preview, 100)))
		}
		status := formatEmailStatus(email)
		if status != "" {
			b.WriteString(fmt.Sprintf("- Status: %s\n", status))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// formatEmail formats a single email for LLM readability.
func formatEmail(email *fastmail.Email) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", email.Subject))
	b.WriteString(fmt.Sprintf("- ID: %s\n", email.ID))
	b.WriteString(fmt.Sprintf("- Thread: %s\n", email.ThreadID))
	if !email.ReceivedAt.IsZero() {
		b.WriteString(fmt.Sprintf("- Received: %s\n", email.ReceivedAt.Format(time.RFC3339)))
	}
	if email.Size > 0 {
		b.WriteString(fmt.Sprintf("- Size: %d bytes\n", email.Size))
	}
	status := formatEmailStatus(*email)
	if status != "" {
		b.WriteString(fmt.Sprintf("- Status: %s\n", status))
	}

	b.WriteString("\n## Preview\n\n")
	if email.Preview != "" {
		b.WriteString(email.Preview)
	} else {
		b.WriteString("(No preview available)")
	}
	b.WriteString("\n")

	return b.String()
}

// formatEmailStatus returns a human-readable status string.
func formatEmailStatus(email fastmail.Email) string {
	var statuses []string
	if email.IsRead() {
		statuses = append(statuses, "read")
	} else {
		statuses = append(statuses, "unread")
	}
	if email.IsFlagged() {
		statuses = append(statuses, "flagged")
	}
	if email.IsDraft() {
		statuses = append(statuses, "draft")
	}
	return strings.Join(statuses, ", ")
}

// formatContactList formats a list of contacts for LLM readability.
func formatContactList(contacts []fastmail.Contact) string {
	if len(contacts) == 0 {
		return "No contacts found."
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Contacts (%d total)\n\n", len(contacts)))

	for i, contact := range contacts {
		b.WriteString(fmt.Sprintf("## %d. %s\n", i+1, contact.Name))
		b.WriteString(fmt.Sprintf("- ID: %s\n", contact.ID))
		if contact.Email != "" {
			b.WriteString(fmt.Sprintf("- Email: %s\n", contact.Email))
		}
		if contact.Phone != "" {
			b.WriteString(fmt.Sprintf("- Phone: %s\n", contact.Phone))
		}
		if contact.Address != "" {
			b.WriteString(fmt.Sprintf("- Address: %s\n", contact.Address))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// formatMaskedEmailList formats a list of masked emails for LLM readability.
func formatMaskedEmailList(maskedEmails []fastmail.MaskedEmail) string {
	if len(maskedEmails) == 0 {
		return "No masked emails found."
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Masked Emails (%d total)\n\n", len(maskedEmails)))

	for i, me := range maskedEmails {
		b.WriteString(fmt.Sprintf("## %d. %s\n", i+1, me.Email))
		b.WriteString(fmt.Sprintf("- ID: %s\n", me.ID))
		b.WriteString(fmt.Sprintf("- State: %s\n", me.State))
		if me.ForDomain != "" {
			b.WriteString(fmt.Sprintf("- Domain: %s\n", me.ForDomain))
		}
		if me.Description != "" {
			b.WriteString(fmt.Sprintf("- Description: %s\n", me.Description))
		}
		if me.LastMessageAt != "" {
			b.WriteString(fmt.Sprintf("- Last Message: %s\n", me.LastMessageAt))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// formatMailboxList formats a list of mailboxes for LLM readability.
func formatMailboxList(mailboxes []fastmail.Mailbox) string {
	if len(mailboxes) == 0 {
		return "No mailboxes found."
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Mailboxes (%d total)\n\n", len(mailboxes)))

	for i, mb := range mailboxes {
		b.WriteString(fmt.Sprintf("## %d. %s\n", i+1, mb.Name))
		b.WriteString(fmt.Sprintf("- ID: %s\n", mb.ID))
		if mb.Role != "" {
			b.WriteString(fmt.Sprintf("- Role: %s\n", mb.Role))
		}
		b.WriteString(fmt.Sprintf("- Total: %d emails, %d unread\n", mb.TotalEmails, mb.UnreadEmails))
		if mb.ParentID != "" {
			b.WriteString(fmt.Sprintf("- Parent: %s\n", mb.ParentID))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// formatIdentityList formats a list of identities for LLM readability.
func formatIdentityList(identities []fastmail.Identity) string {
	if len(identities) == 0 {
		return "No identities found."
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Identities (%d total)\n\n", len(identities)))

	for i, id := range identities {
		b.WriteString(fmt.Sprintf("## %d. %s\n", i+1, id.Name))
		b.WriteString(fmt.Sprintf("- ID: %s\n", id.ID))
		b.WriteString(fmt.Sprintf("- Email: %s\n", id.Email))
		if id.TextSignature != "" {
			b.WriteString(fmt.Sprintf("- Signature: %s\n", truncate(id.TextSignature, 80)))
		}
		if id.MayDelete {
			b.WriteString("- May delete: yes\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

// formatVacation formats a vacation response for LLM readability.
func formatVacation(v *fastmail.Vacation) string {
	var b strings.Builder
	b.WriteString("# Vacation Response\n\n")

	if v.IsEnabled {
		b.WriteString("- Status: **enabled**\n")
	} else {
		b.WriteString("- Status: disabled\n")
	}
	if v.Subject != "" {
		b.WriteString(fmt.Sprintf("- Subject: %s\n", v.Subject))
	}
	if v.FromDate != "" {
		b.WriteString(fmt.Sprintf("- From: %s\n", v.FromDate))
	}
	if v.ToDate != "" {
		b.WriteString(fmt.Sprintf("- To: %s\n", v.ToDate))
	}
	if v.TextBody != "" {
		b.WriteString(fmt.Sprintf("\n## Message\n\n%s\n", v.TextBody))
	}

	return b.String()
}

// formatCalendarList formats a list of calendars for LLM readability.
func formatCalendarList(calendars []fastmail.Calendar) string {
	if len(calendars) == 0 {
		return "No calendars found."
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Calendars (%d total)\n\n", len(calendars)))

	for i, cal := range calendars {
		b.WriteString(fmt.Sprintf("## %d. %s\n", i+1, cal.Name))
		b.WriteString(fmt.Sprintf("- ID: %s\n", cal.ID))
		if cal.Description != "" {
			b.WriteString(fmt.Sprintf("- Description: %s\n", cal.Description))
		}
		if cal.Color != "" {
			b.WriteString(fmt.Sprintf("- Color: %s\n", cal.Color))
		}
		if cal.IsDefaultCalendar {
			b.WriteString("- Default: yes\n")
		}
		if cal.ReadOnly {
			b.WriteString("- Read-only: yes\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

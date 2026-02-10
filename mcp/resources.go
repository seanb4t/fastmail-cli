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
	resources []registeredResource
	templates []ResourceTemplate
}

// NewResourceRegistry creates a new resource registry with default resources.
func NewResourceRegistry(client *fastmail.Client) *ResourceRegistry {
	r := &ResourceRegistry{
		client: client,
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

	// fastmail://contact/{id} - Single contact (template)
	r.registerTemplate(
		ResourceTemplate{
			URITemplate: "fastmail://contact/{id}",
			Name:        "Contact",
			Description: "Details of a specific contact",
			MimeType:    "text/plain",
		},
		regexp.MustCompile(`^fastmail://contact/(?P<id>[^/]+)$`),
		r.handleContact,
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

	// fastmail://masked-email/{id} - Single masked email (template)
	r.registerTemplate(
		ResourceTemplate{
			URITemplate: "fastmail://masked-email/{id}",
			Name:        "Masked Email",
			Description: "Details of a specific masked email address",
			MimeType:    "text/plain",
		},
		regexp.MustCompile(`^fastmail://masked-email/(?P<id>[^/]+)$`),
		r.handleMaskedEmail,
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

	email, err := r.client.Mail().GetWithBody(ctx, id)
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
	return &ResourceContent{
		URI:      "fastmail://contacts",
		MimeType: "text/plain",
		Text:     "Contacts access requires CardDAV configuration. Use the contacts_list tool instead.",
	}, nil
}

// handleContact returns a single contact by ID.
func (r *ResourceRegistry) handleContact(_ context.Context, params map[string]string) (*ResourceContent, error) {
	id := params["id"]
	if id == "" {
		return nil, oops.Errorf("contact ID required")
	}

	// Contacts require CardDAV client — direct the agent to use the tool instead
	return &ResourceContent{
		URI:      fmt.Sprintf("fastmail://contact/%s", id),
		MimeType: "text/plain",
		Text:     fmt.Sprintf("Contact lookup requires CardDAV configuration. Use the contacts_get tool with id=%q instead.", id),
	}, nil
}

// handleCalendarToday returns today's calendar events.
func (r *ResourceRegistry) handleCalendarToday(_ context.Context, _ map[string]string) (*ResourceContent, error) {
	// Calendar service requires DAV client which may not be configured
	// Return informative message if not available
	return &ResourceContent{
		URI:      "fastmail://calendar/today",
		MimeType: "text/plain",
		Text:     "Calendar access requires CalDAV configuration. No events available.",
	}, nil
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

// handleMaskedEmail returns a single masked email by ID.
func (r *ResourceRegistry) handleMaskedEmail(ctx context.Context, params map[string]string) (*ResourceContent, error) {
	if r.client == nil {
		return nil, oops.Errorf("client not configured")
	}

	id := params["id"]
	if id == "" {
		return nil, oops.Errorf("masked email ID required")
	}

	me, err := r.client.MaskedEmail().Get(ctx, id)
	if err != nil {
		return nil, oops.Wrapf(err, "getting masked email %s", id)
	}

	content := formatMaskedEmail(me)
	return &ResourceContent{
		URI:      fmt.Sprintf("fastmail://masked-email/%s", id),
		MimeType: "text/plain",
		Text:     content,
	}, nil
}

// formatMaskedEmail formats a single masked email for LLM readability.
func formatMaskedEmail(me *fastmail.MaskedEmail) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", me.Email))
	b.WriteString(fmt.Sprintf("- ID: %s\n", me.ID))
	b.WriteString(fmt.Sprintf("- State: %s\n", me.State))
	if me.ForDomain != "" {
		b.WriteString(fmt.Sprintf("- Domain: %s\n", me.ForDomain))
	}
	if me.Description != "" {
		b.WriteString(fmt.Sprintf("- Description: %s\n", me.Description))
	}
	if me.URL != "" {
		b.WriteString(fmt.Sprintf("- URL: %s\n", me.URL))
	}
	if me.CreatedBy != "" {
		b.WriteString(fmt.Sprintf("- Created By: %s\n", me.CreatedBy))
	}
	if me.CreatedAt != "" {
		b.WriteString(fmt.Sprintf("- Created: %s\n", me.CreatedAt))
	}
	if me.LastMessageAt != "" {
		b.WriteString(fmt.Sprintf("- Last Message: %s\n", me.LastMessageAt))
	}

	return b.String()
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

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Package dav provides CalDAV protocol client functionality.
package dav

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

// Event represents a calendar event with iCalendar properties.
// This is the internal representation used by the DAV layer.
type Event struct {
	// ID is the CalDAV resource path/identifier.
	ID string

	// CalendarID is the ID of the calendar containing this event.
	CalendarID string

	// UID is the iCalendar UID.
	UID string

	// Summary is the event title.
	Summary string

	// Description is the event description.
	Description string

	// Location is where the event takes place.
	Location string

	// Start is when the event begins.
	Start time.Time

	// End is when the event ends.
	End time.Time

	// AllDay indicates if this is an all-day event.
	AllDay bool

	// RecurrenceRule is the RRULE value.
	RecurrenceRule string

	// Status is the event status.
	Status string

	// Created is when the event was created.
	Created time.Time

	// LastModified is when the event was last modified.
	LastModified time.Time
}

// IsRecurring reports whether this event has a recurrence rule.
func (e *Event) IsRecurring() bool {
	return e.RecurrenceRule != ""
}

// Calendar represents a CalDAV calendar collection.
type Calendar struct {
	// ID is the CalDAV calendar path/identifier.
	ID string

	// Name is the display name.
	Name string

	// Description is the calendar description.
	Description string

	// Color is the calendar color.
	Color string

	// IsDefault indicates if this is the default calendar.
	IsDefault bool

	// ReadOnly indicates if the calendar is read-only.
	ReadOnly bool
}

// Client is a CalDAV protocol client.
type Client struct {
	endpoint   string
	username   string
	password   string
	httpClient webdav.HTTPClient
	caldav     *caldav.Client
}

// NewClient creates a new CalDAV client.
func NewClient(endpoint, username, password string) *Client {
	httpClient := &http.Client{}
	return &Client{
		endpoint:   endpoint,
		username:   username,
		password:   password,
		httpClient: webdav.HTTPClientWithBasicAuth(httpClient, username, password),
	}
}

// ListCalendars returns all calendars for the user.
func (c *Client) ListCalendars(ctx context.Context) ([]Calendar, error) {
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting caldav client: %w", err)
	}

	principal, err := client.FindCurrentUserPrincipal(ctx)
	if err != nil {
		return nil, fmt.Errorf("finding principal: %w", err)
	}

	homeSet, err := client.FindCalendarHomeSet(ctx, principal)
	if err != nil {
		return nil, fmt.Errorf("finding calendar home set: %w", err)
	}

	cals, err := client.FindCalendars(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("finding calendars: %w", err)
	}

	result := make([]Calendar, 0, len(cals))
	for _, cal := range cals {
		result = append(result, Calendar{
			ID:          cal.Path,
			Name:        cal.Name,
			Description: cal.Description,
		})
	}

	return result, nil
}

// ListEvents returns events in the given calendar within the date range.
func (c *Client) ListEvents(ctx context.Context, calendarID string, start, end time.Time) ([]Event, error) {
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting caldav client: %w", err)
	}

	query := &caldav.CalendarQuery{
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []caldav.CompFilter{{
				Name:  "VEVENT",
				Start: start,
				End:   end,
			}},
		},
	}

	objects, err := client.QueryCalendar(ctx, calendarID, query)
	if err != nil {
		return nil, fmt.Errorf("querying calendar: %w", err)
	}

	result := make([]Event, 0, len(objects))
	for _, obj := range objects {
		if obj.Data == nil {
			continue
		}
		event := convertCalendarObject(obj, calendarID)
		if event != nil {
			result = append(result, *event)
		}
	}

	return result, nil
}

// GetEvent returns a single event by ID.
func (c *Client) GetEvent(ctx context.Context, calendarID, eventID string) (*Event, error) {
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting caldav client: %w", err)
	}

	obj, err := client.GetCalendarObject(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("getting calendar object: %w", err)
	}

	return convertCalendarObject(*obj, calendarID), nil
}

// CreateEvent creates a new event in the calendar.
func (c *Client) CreateEvent(ctx context.Context, calendarID string, event *Event) error {
	client, err := c.getClient(ctx)
	if err != nil {
		return fmt.Errorf("getting caldav client: %w", err)
	}

	cal := createICalendar(event)

	_, err = client.PutCalendarObject(ctx, calendarID+event.UID+".ics", cal)
	if err != nil {
		return fmt.Errorf("creating event: %w", err)
	}

	return nil
}

// UpdateEvent updates an existing event.
func (c *Client) UpdateEvent(ctx context.Context, calendarID string, event *Event) error {
	client, err := c.getClient(ctx)
	if err != nil {
		return fmt.Errorf("getting caldav client: %w", err)
	}

	cal := createICalendar(event)

	_, err = client.PutCalendarObject(ctx, event.ID, cal)
	if err != nil {
		return fmt.Errorf("updating event: %w", err)
	}

	return nil
}

// DeleteEvent deletes an event by ID.
func (c *Client) DeleteEvent(ctx context.Context, eventID string) error {
	client, err := c.getClient(ctx)
	if err != nil {
		return fmt.Errorf("getting caldav client: %w", err)
	}

	err = client.RemoveAll(ctx, eventID)
	if err != nil {
		return fmt.Errorf("deleting event: %w", err)
	}

	return nil
}

func (c *Client) getClient(ctx context.Context) (*caldav.Client, error) {
	if c.caldav != nil {
		return c.caldav, nil
	}

	client, err := caldav.NewClient(c.httpClient, c.endpoint)
	if err != nil {
		return nil, err
	}

	c.caldav = client
	return client, nil
}

// ParseICalendarEvent parses iCalendar data into an Event.
func ParseICalendarEvent(data []byte) (*Event, error) {
	dec := ical.NewDecoder(bytes.NewReader(data))
	cal, err := dec.Decode()
	if err != nil {
		return nil, fmt.Errorf("decoding icalendar: %w", err)
	}

	for _, child := range cal.Children {
		if child.Name == ical.CompEvent {
			return parseVEvent(child), nil
		}
	}

	return nil, fmt.Errorf("no VEVENT found in calendar data")
}

// SerializeEventToICalendar converts an Event to iCalendar format.
func SerializeEventToICalendar(event *Event) ([]byte, error) {
	cal := createICalendar(event)

	var buf bytes.Buffer
	enc := ical.NewEncoder(&buf)
	if err := enc.Encode(cal); err != nil {
		return nil, fmt.Errorf("encoding icalendar: %w", err)
	}

	return buf.Bytes(), nil
}

func parseVEvent(comp *ical.Component) *Event {
	event := &Event{
		UID:         comp.Props.Get(ical.PropUID).Value,
		Summary:     getPropValue(comp, ical.PropSummary),
		Description: getPropValue(comp, ical.PropDescription),
		Location:    getPropValue(comp, ical.PropLocation),
		Status:      getPropValue(comp, ical.PropStatus),
	}

	// Parse DTSTART
	if dtstart := comp.Props.Get(ical.PropDateTimeStart); dtstart != nil {
		if t, err := dtstart.DateTime(time.UTC); err == nil {
			event.Start = t
		}
		// Check if all-day event (DATE value type)
		if dtstart.Params.Get(ical.ParamValue) == "DATE" {
			event.AllDay = true
		}
	}

	// Parse DTEND
	if dtend := comp.Props.Get(ical.PropDateTimeEnd); dtend != nil {
		if t, err := dtend.DateTime(time.UTC); err == nil {
			event.End = t
		}
	}

	// Parse RRULE
	if rrule := comp.Props.Get(ical.PropRecurrenceRule); rrule != nil {
		event.RecurrenceRule = rrule.Value
	}

	// Parse CREATED
	if created := comp.Props.Get(ical.PropCreated); created != nil {
		if t, err := created.DateTime(time.UTC); err == nil {
			event.Created = t
		}
	}

	// Parse LAST-MODIFIED
	if lastmod := comp.Props.Get(ical.PropLastModified); lastmod != nil {
		if t, err := lastmod.DateTime(time.UTC); err == nil {
			event.LastModified = t
		}
	}

	return event
}

func getPropValue(comp *ical.Component, name string) string {
	if prop := comp.Props.Get(name); prop != nil {
		return prop.Value
	}
	return ""
}

func createICalendar(event *Event) *ical.Calendar {
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//fastmail-cli//EN")

	vevent := ical.NewComponent(ical.CompEvent)
	vevent.Props.SetText(ical.PropUID, event.UID)
	vevent.Props.SetText(ical.PropSummary, event.Summary)

	// DTSTAMP is required by iCalendar spec
	vevent.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())

	if event.Description != "" {
		vevent.Props.SetText(ical.PropDescription, event.Description)
	}
	if event.Location != "" {
		vevent.Props.SetText(ical.PropLocation, event.Location)
	}
	if event.Status != "" {
		vevent.Props.SetText(ical.PropStatus, event.Status)
	}

	// Set DTSTART and DTEND
	if event.AllDay {
		vevent.Props.SetDate(ical.PropDateTimeStart, event.Start)
		vevent.Props.SetDate(ical.PropDateTimeEnd, event.End)
	} else {
		vevent.Props.SetDateTime(ical.PropDateTimeStart, event.Start)
		vevent.Props.SetDateTime(ical.PropDateTimeEnd, event.End)
	}

	if event.RecurrenceRule != "" {
		vevent.Props.SetText(ical.PropRecurrenceRule, event.RecurrenceRule)
	}

	cal.Children = append(cal.Children, vevent)
	return cal
}

func convertCalendarObject(obj caldav.CalendarObject, calendarID string) *Event {
	if obj.Data == nil {
		return nil
	}

	for _, child := range obj.Data.Children {
		if child.Name == ical.CompEvent {
			event := parseVEvent(child)
			event.ID = obj.Path
			event.CalendarID = calendarID
			return event
		}
	}

	return nil
}

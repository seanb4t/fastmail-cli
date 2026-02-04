package fastmail

import (
	"context"
	"time"

	"github.com/samber/oops"
	"github.com/seanb4t/fastmail-cli/internal/dav"
)

// Calendar represents a calendar collection.
//
// This is a domain type designed for ease of use, not a direct mapping
// of the CalDAV calendar object. Protocol-specific details are hidden.
type Calendar struct {
	// ID is the unique identifier for this calendar.
	ID string

	// Name is the display name of the calendar.
	Name string

	// Color is the calendar's display color (hex format, e.g., "#FF5733").
	Color string

	// Description is an optional description of the calendar.
	Description string

	// IsDefaultCalendar indicates if this is the user's default calendar.
	IsDefaultCalendar bool

	// ReadOnly indicates if the calendar is read-only.
	ReadOnly bool
}

// IsDefault reports whether this is the user's default calendar.
func (c *Calendar) IsDefault() bool {
	return c.IsDefaultCalendar
}

// IsReadOnly reports whether this calendar is read-only.
func (c *Calendar) IsReadOnly() bool {
	return c.ReadOnly
}

// CalendarService provides calendar operations.
type CalendarService struct {
	dav *dav.Client
}

// ListCalendars returns all calendars for the user.
func (s *CalendarService) ListCalendars(ctx context.Context) ([]Calendar, error) {
	calendars, err := s.dav.ListCalendars(ctx)
	if err != nil {
		return nil, oops.Wrapf(err, "listing calendars")
	}

	result := make([]Calendar, 0, len(calendars))
	for _, cal := range calendars {
		result = append(result, Calendar{
			ID:          cal.ID,
			Name:        cal.Name,
			Description: cal.Description,
			Color:       cal.Color,
		})
	}

	return result, nil
}

// ListEvents returns events in the given calendar within the date range.
func (s *CalendarService) ListEvents(ctx context.Context, calendarID string, start, end time.Time) ([]Event, error) {
	events, err := s.dav.ListEvents(ctx, calendarID, start, end)
	if err != nil {
		return nil, oops.Wrapf(err, "listing events")
	}

	return convertDAVEvents(events), nil
}

// GetEvent returns a single event by ID.
func (s *CalendarService) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	event, err := s.dav.GetEvent(ctx, "", eventID)
	if err != nil {
		return nil, oops.Wrapf(err, "getting event %s", eventID)
	}

	return convertDAVEvent(event), nil
}

// CreateEvent creates a new event in the default calendar.
func (s *CalendarService) CreateEvent(ctx context.Context, event *Event) error {
	davEvent := toDavEvent(event)
	if err := s.dav.CreateEvent(ctx, event.CalendarID, davEvent); err != nil {
		return oops.Wrapf(err, "creating event")
	}
	return nil
}

// UpdateEvent updates an existing event.
func (s *CalendarService) UpdateEvent(ctx context.Context, event *Event) error {
	davEvent := toDavEvent(event)
	if err := s.dav.UpdateEvent(ctx, event.CalendarID, davEvent); err != nil {
		return oops.Wrapf(err, "updating event")
	}
	return nil
}

// DeleteEvent deletes an event by ID.
func (s *CalendarService) DeleteEvent(ctx context.Context, eventID string) error {
	if err := s.dav.DeleteEvent(ctx, eventID); err != nil {
		return oops.Wrapf(err, "deleting event %s", eventID)
	}
	return nil
}

func convertDAVEvents(davEvents []dav.Event) []Event {
	result := make([]Event, 0, len(davEvents))
	for _, e := range davEvents {
		result = append(result, *convertDAVEvent(&e))
	}
	return result
}

func convertDAVEvent(e *dav.Event) *Event {
	return &Event{
		ID:             e.ID,
		CalendarID:     e.CalendarID,
		UID:            e.UID,
		Summary:        e.Summary,
		Description:    e.Description,
		Location:       e.Location,
		Start:          e.Start,
		End:            e.End,
		AllDay:         e.AllDay,
		RecurrenceRule: e.RecurrenceRule,
		Status:         e.Status,
		Created:        e.Created,
		LastModified:   e.LastModified,
	}
}

func toDavEvent(e *Event) *dav.Event {
	return &dav.Event{
		ID:             e.ID,
		CalendarID:     e.CalendarID,
		UID:            e.UID,
		Summary:        e.Summary,
		Description:    e.Description,
		Location:       e.Location,
		Start:          e.Start,
		End:            e.End,
		AllDay:         e.AllDay,
		RecurrenceRule: e.RecurrenceRule,
		Status:         e.Status,
		Created:        e.Created,
		LastModified:   e.LastModified,
	}
}

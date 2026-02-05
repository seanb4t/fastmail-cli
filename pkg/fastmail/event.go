package fastmail

import "time"

// Event represents a calendar event.
//
// This is a domain type designed for ease of use, not a direct mapping
// of the iCalendar VEVENT component. Protocol-specific details are hidden.
type Event struct {
	// ID is the unique identifier for this event.
	ID string

	// CalendarID is the ID of the calendar containing this event.
	CalendarID string

	// UID is the iCalendar UID for this event.
	UID string

	// Summary is the event title/name.
	Summary string

	// Description is the event description/notes.
	Description string

	// Location is where the event takes place.
	Location string

	// Start is when the event begins.
	Start time.Time

	// End is when the event ends.
	End time.Time

	// AllDay indicates if this is an all-day event.
	AllDay bool

	// RecurrenceRule is the iCalendar RRULE (e.g., "FREQ=WEEKLY;BYDAY=MO").
	// Empty if the event does not recur.
	RecurrenceRule string

	// Status is the event status (e.g., "CONFIRMED", "TENTATIVE", "CANCELED").
	Status string

	// Created is when the event was created.
	Created time.Time

	// LastModified is when the event was last modified.
	LastModified time.Time
}

// Duration returns the duration of the event.
func (e *Event) Duration() time.Duration {
	return e.End.Sub(e.Start)
}

// IsAllDay reports whether this is an all-day event.
func (e *Event) IsAllDay() bool {
	return e.AllDay
}

// IsRecurring reports whether this event has a recurrence rule.
func (e *Event) IsRecurring() bool {
	return e.RecurrenceRule != ""
}

// HasLocation reports whether this event has a location set.
func (e *Event) HasLocation() bool {
	return e.Location != ""
}

// InDateRange reports whether this event overlaps with the given date range.
// An event is considered in range if any part of it falls within the range.
func (e *Event) InDateRange(start, end time.Time) bool {
	// Event overlaps if it starts before range ends AND ends after range starts
	return e.Start.Before(end) && e.End.After(start)
}

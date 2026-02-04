package fastmail

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalendar_IsDefault(t *testing.T) {
	tests := []struct {
		name     string
		calendar Calendar
		want     bool
	}{
		{
			name:     "default calendar returns true",
			calendar: Calendar{IsDefaultCalendar: true},
			want:     true,
		},
		{
			name:     "non-default calendar returns false",
			calendar: Calendar{IsDefaultCalendar: false},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.calendar.IsDefault())
		})
	}
}

func TestCalendar_IsReadOnly(t *testing.T) {
	tests := []struct {
		name     string
		calendar Calendar
		want     bool
	}{
		{
			name:     "read-only calendar returns true",
			calendar: Calendar{ReadOnly: true},
			want:     true,
		},
		{
			name:     "writable calendar returns false",
			calendar: Calendar{ReadOnly: false},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.calendar.IsReadOnly())
		})
	}
}

// CalendarServiceInterface defines the contract that CalendarService must implement.
// This test ensures the interface is satisfied.
func TestCalendarService_ImplementsInterface(_ *testing.T) {
	// This is a compile-time check that CalendarService implements the expected methods
	var _ interface {
		ListCalendars(ctx context.Context) ([]Calendar, error)
		ListEvents(ctx context.Context, calendarID string, start, end time.Time) ([]Event, error)
		GetEvent(ctx context.Context, eventID string) (*Event, error)
		CreateEvent(ctx context.Context, event *Event) error
		UpdateEvent(ctx context.Context, event *Event) error
		DeleteEvent(ctx context.Context, eventID string) error
	} = (*CalendarService)(nil)
}

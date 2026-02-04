package fastmail

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvent_Duration(t *testing.T) {
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC)

	event := Event{
		Start: start,
		End:   end,
	}

	assert.Equal(t, 90*time.Minute, event.Duration())
}

func TestEvent_IsAllDay(t *testing.T) {
	tests := []struct {
		name  string
		event Event
		want  bool
	}{
		{
			name:  "all-day event returns true",
			event: Event{AllDay: true},
			want:  true,
		},
		{
			name:  "timed event returns false",
			event: Event{AllDay: false},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.event.IsAllDay())
		})
	}
}

func TestEvent_IsRecurring(t *testing.T) {
	tests := []struct {
		name  string
		event Event
		want  bool
	}{
		{
			name:  "event with recurrence rule returns true",
			event: Event{RecurrenceRule: "FREQ=WEEKLY;BYDAY=MO,WE,FR"},
			want:  true,
		},
		{
			name:  "event without recurrence returns false",
			event: Event{RecurrenceRule: ""},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.event.IsRecurring())
		})
	}
}

func TestEvent_HasLocation(t *testing.T) {
	tests := []struct {
		name  string
		event Event
		want  bool
	}{
		{
			name:  "event with location returns true",
			event: Event{Location: "Conference Room A"},
			want:  true,
		},
		{
			name:  "event without location returns false",
			event: Event{Location: ""},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.event.HasLocation())
		})
	}
}

func TestEvent_InDateRange(t *testing.T) {
	event := Event{
		Start: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name       string
		rangeStart time.Time
		rangeEnd   time.Time
		want       bool
	}{
		{
			name:       "event fully within range",
			rangeStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			want:       true,
		},
		{
			name:       "event starts before range",
			rangeStart: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			rangeEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			want:       true, // overlaps
		},
		{
			name:       "event ends after range",
			rangeStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			want:       true, // overlaps
		},
		{
			name:       "event entirely before range",
			rangeStart: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			want:       false,
		},
		{
			name:       "event entirely after range",
			rangeStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			rangeEnd:   time.Date(2024, 1, 14, 23, 59, 59, 0, time.UTC),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, event.InDateRange(tt.rangeStart, tt.rangeEnd))
		})
	}
}

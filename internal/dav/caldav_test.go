package dav

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseICalendarEvent(t *testing.T) {
	ical := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:test-uid-123@example.com
SUMMARY:Team Meeting
DESCRIPTION:Weekly sync
LOCATION:Conference Room A
DTSTART:20240115T100000Z
DTEND:20240115T110000Z
STATUS:CONFIRMED
CREATED:20240101T120000Z
LAST-MODIFIED:20240110T090000Z
END:VEVENT
END:VCALENDAR`

	event, err := ParseICalendarEvent([]byte(ical))
	require.NoError(t, err)

	assert.Equal(t, "test-uid-123@example.com", event.UID)
	assert.Equal(t, "Team Meeting", event.Summary)
	assert.Equal(t, "Weekly sync", event.Description)
	assert.Equal(t, "Conference Room A", event.Location)
	assert.Equal(t, "CONFIRMED", event.Status)
	assert.Equal(t, time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), event.Start)
	assert.Equal(t, time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC), event.End)
}

func TestParseICalendarEvent_AllDay(t *testing.T) {
	ical := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:allday-123@example.com
SUMMARY:Holiday
DTSTART;VALUE=DATE:20240115
DTEND;VALUE=DATE:20240116
END:VEVENT
END:VCALENDAR`

	event, err := ParseICalendarEvent([]byte(ical))
	require.NoError(t, err)

	assert.Equal(t, "allday-123@example.com", event.UID)
	assert.Equal(t, "Holiday", event.Summary)
	assert.True(t, event.AllDay)
}

func TestParseICalendarEvent_Recurring(t *testing.T) {
	ical := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:recurring-123@example.com
SUMMARY:Daily Standup
DTSTART:20240115T090000Z
DTEND:20240115T091500Z
RRULE:FREQ=DAILY;BYDAY=MO,TU,WE,TH,FR
END:VEVENT
END:VCALENDAR`

	event, err := ParseICalendarEvent([]byte(ical))
	require.NoError(t, err)

	assert.Equal(t, "recurring-123@example.com", event.UID)
	assert.Equal(t, "Daily Standup", event.Summary)
	assert.Equal(t, "FREQ=DAILY;BYDAY=MO,TU,WE,TH,FR", event.RecurrenceRule)
	assert.True(t, event.IsRecurring())
}

func TestSerializeEventToICalendar(t *testing.T) {
	event := &Event{
		UID:         "test-uid-456@example.com",
		Summary:     "Project Review",
		Description: "Quarterly review meeting",
		Location:    "Main Office",
		Start:       time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC),
		End:         time.Date(2024, 2, 20, 15, 30, 0, 0, time.UTC),
		Status:      "CONFIRMED",
	}

	data, err := SerializeEventToICalendar(event)
	require.NoError(t, err)

	// Parse it back to verify
	parsed, err := ParseICalendarEvent(data)
	require.NoError(t, err)

	assert.Equal(t, event.UID, parsed.UID)
	assert.Equal(t, event.Summary, parsed.Summary)
	assert.Equal(t, event.Description, parsed.Description)
	assert.Equal(t, event.Location, parsed.Location)
	assert.Equal(t, event.Start.UTC(), parsed.Start.UTC())
	assert.Equal(t, event.End.UTC(), parsed.End.UTC())
}

func TestNewClient(t *testing.T) {
	client := NewClient("https://caldav.example.com", "user", "pass")
	assert.NotNil(t, client)
}

func TestClient_HTTPClient(t *testing.T) {
	// Test that client is configured correctly
	client := NewClient("https://caldav.example.com", "user", "pass")
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, "https://caldav.example.com", client.endpoint)
	assert.Equal(t, "user", client.username)
	assert.Equal(t, "pass", client.password)
}

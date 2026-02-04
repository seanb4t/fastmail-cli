# CalendarService

Calendar and event operations via CalDAV.

## Methods

### ListCalendars

Returns all calendars for the user.

```go
func (s *CalendarService) ListCalendars(ctx context.Context) ([]Calendar, error)
```

#### Example

```go
calendars, err := calendarService.ListCalendars(ctx)
for _, cal := range calendars {
    fmt.Printf("%s (%s)\n", cal.Name, cal.ID)
}
```

### ListEvents

Returns events in a calendar within a date range.

```go
func (s *CalendarService) ListEvents(ctx context.Context, calendarID string, start, end time.Time) ([]Event, error)
```

#### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `calendarID` | `string` | Calendar ID from `ListCalendars` |
| `start` | `time.Time` | Range start (inclusive) |
| `end` | `time.Time` | Range end (exclusive) |

#### Example

```go
// Get events for this week
now := time.Now()
start := now.Truncate(24 * time.Hour)
end := start.AddDate(0, 0, 7)

events, err := calendarService.ListEvents(ctx, calendarID, start, end)
for _, e := range events {
    fmt.Printf("%s: %s\n", e.Start.Format("Mon 3:04 PM"), e.Summary)
}
```

### GetEvent

Returns a single event by ID.

```go
func (s *CalendarService) GetEvent(ctx context.Context, eventID string) (*Event, error)
```

#### Example

```go
event, err := calendarService.GetEvent(ctx, eventID)
fmt.Printf("%s at %s\n", event.Summary, event.Location)
```

### CreateEvent

Creates a new event in a calendar.

```go
func (s *CalendarService) CreateEvent(ctx context.Context, event *Event) error
```

#### Example

```go
event := &fastmail.Event{
    CalendarID:  calendarID,
    Summary:     "Team Meeting",
    Description: "Weekly sync",
    Location:    "Conference Room A",
    Start:       time.Date(2026, 2, 5, 14, 0, 0, 0, time.Local),
    End:         time.Date(2026, 2, 5, 15, 0, 0, 0, time.Local),
}

err := calendarService.CreateEvent(ctx, event)
```

### UpdateEvent

Updates an existing event.

```go
func (s *CalendarService) UpdateEvent(ctx context.Context, event *Event) error
```

#### Example

```go
event, _ := calendarService.GetEvent(ctx, eventID)
event.Location = "Conference Room B"
err := calendarService.UpdateEvent(ctx, event)
```

### DeleteEvent

Deletes an event by ID.

```go
func (s *CalendarService) DeleteEvent(ctx context.Context, eventID string) error
```

#### Example

```go
err := calendarService.DeleteEvent(ctx, eventID)
```

## Types

### Calendar

Represents a calendar collection.

```go
type Calendar struct {
    ID                string
    Name              string
    Color             string
    Description       string
    IsDefaultCalendar bool
    ReadOnly          bool
}
```

| Field | Description |
|-------|-------------|
| `ID` | Unique identifier |
| `Name` | Display name |
| `Color` | Hex color (e.g., `"#FF5733"`) |
| `Description` | Optional description |
| `IsDefaultCalendar` | Whether this is the default calendar |
| `ReadOnly` | Whether calendar is read-only |

#### Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `IsDefault()` | `bool` | Check if default calendar |
| `IsReadOnly()` | `bool` | Check if read-only |

### Event

Represents a calendar event.

```go
type Event struct {
    ID             string
    CalendarID     string
    UID            string
    Summary        string
    Description    string
    Location       string
    Start          time.Time
    End            time.Time
    AllDay         bool
    RecurrenceRule string
    Status         string
    Created        time.Time
    LastModified   time.Time
}
```

| Field | Description |
|-------|-------------|
| `ID` | Unique identifier |
| `CalendarID` | Parent calendar ID |
| `UID` | iCalendar UID |
| `Summary` | Event title |
| `Description` | Event notes |
| `Location` | Event location |
| `Start` | Event start time |
| `End` | Event end time |
| `AllDay` | All-day event flag |
| `RecurrenceRule` | RRULE (e.g., `"FREQ=WEEKLY;BYDAY=MO"`) |
| `Status` | Event status (`"CONFIRMED"`, `"TENTATIVE"`, `"CANCELLED"`) |
| `Created` | Creation timestamp |
| `LastModified` | Last modification timestamp |

#### Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `Duration()` | `time.Duration` | Event duration |
| `IsAllDay()` | `bool` | Check if all-day event |
| `IsRecurring()` | `bool` | Check if has recurrence rule |
| `HasLocation()` | `bool` | Check if location is set |
| `InDateRange(start, end)` | `bool` | Check if overlaps date range |

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func main() {
    ctx := context.Background()

    // Assume calendarService is obtained from Client
    var calendarService *fastmail.CalendarService

    // List calendars
    calendars, err := calendarService.ListCalendars(ctx)
    if err != nil {
        log.Fatal(err)
    }

    if len(calendars) == 0 {
        log.Fatal("No calendars found")
    }

    primaryCal := calendars[0]
    fmt.Printf("Using calendar: %s\n", primaryCal.Name)

    // List upcoming events
    now := time.Now()
    events, _ := calendarService.ListEvents(ctx, primaryCal.ID, now, now.AddDate(0, 1, 0))

    fmt.Printf("Found %d events in the next month\n", len(events))
    for _, e := range events {
        if e.IsAllDay() {
            fmt.Printf("  %s (all day): %s\n", e.Start.Format("Jan 2"), e.Summary)
        } else {
            fmt.Printf("  %s: %s\n", e.Start.Format("Jan 2 3:04 PM"), e.Summary)
        }
    }

    // Create a new event
    meeting := &fastmail.Event{
        CalendarID: primaryCal.ID,
        Summary:    "Project Review",
        Location:   "Zoom",
        Start:      time.Now().Add(24 * time.Hour).Truncate(time.Hour),
        End:        time.Now().Add(25 * time.Hour).Truncate(time.Hour),
    }

    if err := calendarService.CreateEvent(ctx, meeting); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Event created")
}
```

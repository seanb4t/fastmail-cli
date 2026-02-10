//go:build integration

package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"

	"github.com/seanb4t/fastmail-cli/internal/dav"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func registerCalendarSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the following calendars exist:$`, theFollowingCalendarsExist)
	sc.Step(`^the following events exist in calendar "([^"]*)":$`, theFollowingEventsExistInCalendar)

	// When steps
	sc.Step(`^I list calendars$`, iListCalendars)
	sc.Step(`^I list events in calendar "([^"]*)" from "([^"]*)" to "([^"]*)"$`, iListEventsInCalendar)
	sc.Step(`^I get event "([^"]*)"$`, iGetEvent)
	sc.Step(`^I create an event in calendar "([^"]*)" with:$`, iCreateEvent)
	sc.Step(`^I update event "([^"]*)" in calendar "([^"]*)" with summary "([^"]*)"$`, iUpdateEvent)
	sc.Step(`^I delete event "([^"]*)"$`, iDeleteEvent)

	// Then steps
	sc.Step(`^I should receive (\d+) calendars?$`, iShouldReceiveNCalendars)
	sc.Step(`^calendar (\d+) should have name "([^"]*)"$`, calendarShouldHaveName)
	sc.Step(`^calendar (\d+) should have description "([^"]*)"$`, calendarShouldHaveDescription)
	sc.Step(`^I should receive (\d+) events?$`, iShouldReceiveNEvents)
	sc.Step(`^event (\d+) should have summary "([^"]*)"$`, eventShouldHaveSummary)
	sc.Step(`^event (\d+) should have location "([^"]*)"$`, eventShouldHaveLocation)
	sc.Step(`^the operation should succeed$`, theCalendarOperationShouldSucceed)
	sc.Step(`^the event should have summary "([^"]*)"$`, theEventShouldHaveSummary)
	sc.Step(`^the event should have location "([^"]*)"$`, theEventShouldHaveLocation)
	sc.Step(`^the event should have description "([^"]*)"$`, theEventShouldHaveDescription)
}

// Given steps

func theFollowingCalendarsExist(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getCalendarData(w)
	data.Calendars = parseCalendarTable(table)
	return ctx, nil
}

func theFollowingEventsExistInCalendar(ctx context.Context, calendarID string, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getCalendarData(w)
	events := parseEventTable(table, calendarID)
	data.Events = append(data.Events, events...)
	return ctx, nil
}

// When steps

func iListCalendars(ctx context.Context) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getCalendarData(w)
	result := getCalendarResult(w)

	server := newMockCalDAVServer(data)
	defer server.Close()

	davClient := dav.NewClient(server.URL, "test", "test-token")
	calService := fastmail.NewCalendarService(davClient)

	calendars, err := calService.ListCalendars(ctx)
	result.OperationErr = err
	if err == nil {
		result.Calendars = make([]calendarResultItem, len(calendars))
		for i, cal := range calendars {
			result.Calendars[i] = calendarResultItem{
				ID:          cal.ID,
				Name:        cal.Name,
				Description: cal.Description,
			}
		}
	}
	return ctx, nil
}

func iListEventsInCalendar(ctx context.Context, calendarID, startStr, endStr string) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getCalendarData(w)
	result := getCalendarResult(w)

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return ctx, fmt.Errorf("parsing start time %q: %w", startStr, err)
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return ctx, fmt.Errorf("parsing end time %q: %w", endStr, err)
	}

	server := newMockCalDAVServer(data)
	defer server.Close()

	davClient := dav.NewClient(server.URL, "test", "test-token")
	calService := fastmail.NewCalendarService(davClient)

	events, err := calService.ListEvents(ctx, calendarID, start, end)
	result.OperationErr = err
	if err == nil {
		result.Events = make([]eventResultItem, len(events))
		for i, evt := range events {
			result.Events[i] = eventResultItem{
				UID:         evt.UID,
				Summary:     evt.Summary,
				Description: evt.Description,
				Location:    evt.Location,
				Start:       evt.Start,
				End:         evt.End,
			}
		}
	}
	return ctx, nil
}

func iGetEvent(ctx context.Context, eventID string) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getCalendarData(w)
	result := getCalendarResult(w)

	server := newMockCalDAVServer(data)
	defer server.Close()

	davClient := dav.NewClient(server.URL, "test", "test-token")
	calService := fastmail.NewCalendarService(davClient)

	event, err := calService.GetEvent(ctx, eventID)
	result.OperationErr = err
	if err == nil && event != nil {
		result.SingleEvent = &eventResultItem{
			UID:         event.UID,
			Summary:     event.Summary,
			Description: event.Description,
			Location:    event.Location,
			Start:       event.Start,
			End:         event.End,
		}
	}
	return ctx, nil
}

func iCreateEvent(ctx context.Context, calendarID string, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getCalendarData(w)
	result := getCalendarResult(w)

	events := parseEventTable(table, calendarID)
	if len(events) == 0 {
		return ctx, fmt.Errorf("no event data provided in table")
	}

	me := events[0]

	server := newMockCalDAVServer(data)
	defer server.Close()

	davClient := dav.NewClient(server.URL, "test", "test-token")
	calService := fastmail.NewCalendarService(davClient)

	event := &fastmail.Event{
		CalendarID:  calendarID,
		UID:         me.UID,
		Summary:     me.Summary,
		Description: me.Description,
		Location:    me.Location,
		Start:       me.Start,
		End:         me.End,
	}

	result.OperationErr = calService.CreateEvent(ctx, event)
	return ctx, nil
}

func iUpdateEvent(ctx context.Context, eventID, calendarID, summary string) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getCalendarData(w)
	result := getCalendarResult(w)

	server := newMockCalDAVServer(data)
	defer server.Close()

	davClient := dav.NewClient(server.URL, "test", "test-token")
	calService := fastmail.NewCalendarService(davClient)

	event := &fastmail.Event{
		ID:         eventID,
		CalendarID: calendarID,
		UID:        "updated-event",
		Summary:    summary,
		Start:      time.Now().UTC(),
		End:        time.Now().UTC().Add(time.Hour),
	}

	result.OperationErr = calService.UpdateEvent(ctx, event)
	return ctx, nil
}

func iDeleteEvent(ctx context.Context, eventID string) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getCalendarData(w)
	result := getCalendarResult(w)

	server := newMockCalDAVServer(data)
	defer server.Close()

	davClient := dav.NewClient(server.URL, "test", "test-token")
	calService := fastmail.NewCalendarService(davClient)

	result.OperationErr = calService.DeleteEvent(ctx, eventID)
	return ctx, nil
}

// Then steps

func theCalendarOperationShouldSucceed(ctx context.Context) error {
	w := WorldFromContext(ctx)
	result := getCalendarResult(w)
	if result.OperationErr != nil {
		return fmt.Errorf("expected operation to succeed, got error: %w", result.OperationErr)
	}
	return nil
}

func theEventShouldHaveSummary(ctx context.Context, summary string) error {
	w := WorldFromContext(ctx)
	result := getCalendarResult(w)
	if result.SingleEvent == nil {
		return fmt.Errorf("no single event result available")
	}
	if result.SingleEvent.Summary != summary {
		return fmt.Errorf("expected event summary %q, got %q", summary, result.SingleEvent.Summary)
	}
	return nil
}

func theEventShouldHaveLocation(ctx context.Context, location string) error {
	w := WorldFromContext(ctx)
	result := getCalendarResult(w)
	if result.SingleEvent == nil {
		return fmt.Errorf("no single event result available")
	}
	if result.SingleEvent.Location != location {
		return fmt.Errorf("expected event location %q, got %q", location, result.SingleEvent.Location)
	}
	return nil
}

func theEventShouldHaveDescription(ctx context.Context, description string) error {
	w := WorldFromContext(ctx)
	result := getCalendarResult(w)
	if result.SingleEvent == nil {
		return fmt.Errorf("no single event result available")
	}
	if result.SingleEvent.Description != description {
		return fmt.Errorf("expected event description %q, got %q", description, result.SingleEvent.Description)
	}
	return nil
}

func iShouldReceiveNCalendars(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	result := getCalendarResult(w)
	if result.OperationErr != nil {
		return fmt.Errorf("unexpected error: %w", result.OperationErr)
	}
	if len(result.Calendars) != count {
		return fmt.Errorf("expected %d calendars, got %d", count, len(result.Calendars))
	}
	return nil
}

func calendarShouldHaveName(ctx context.Context, index int, name string) error {
	w := WorldFromContext(ctx)
	result := getCalendarResult(w)
	i := index - 1
	if i < 0 || i >= len(result.Calendars) {
		return fmt.Errorf("calendar index %d out of range (have %d)", index, len(result.Calendars))
	}
	actual := result.Calendars[i].Name
	if actual != name {
		return fmt.Errorf("expected calendar name %q, got %q", name, actual)
	}
	return nil
}

func calendarShouldHaveDescription(ctx context.Context, index int, description string) error {
	w := WorldFromContext(ctx)
	result := getCalendarResult(w)
	i := index - 1
	if i < 0 || i >= len(result.Calendars) {
		return fmt.Errorf("calendar index %d out of range (have %d)", index, len(result.Calendars))
	}
	actual := result.Calendars[i].Description
	if actual != description {
		return fmt.Errorf("expected calendar description %q, got %q", description, actual)
	}
	return nil
}

func iShouldReceiveNEvents(ctx context.Context, count int) error {
	w := WorldFromContext(ctx)
	result := getCalendarResult(w)
	if result.OperationErr != nil {
		return fmt.Errorf("unexpected error: %w", result.OperationErr)
	}
	if len(result.Events) != count {
		return fmt.Errorf("expected %d events, got %d", count, len(result.Events))
	}
	return nil
}

func eventShouldHaveSummary(ctx context.Context, index int, summary string) error {
	w := WorldFromContext(ctx)
	result := getCalendarResult(w)
	i := index - 1
	if i < 0 || i >= len(result.Events) {
		return fmt.Errorf("event index %d out of range (have %d)", index, len(result.Events))
	}
	actual := result.Events[i].Summary
	if actual != summary {
		return fmt.Errorf("expected event summary %q, got %q", summary, actual)
	}
	return nil
}

func eventShouldHaveLocation(ctx context.Context, index int, location string) error {
	w := WorldFromContext(ctx)
	result := getCalendarResult(w)
	i := index - 1
	if i < 0 || i >= len(result.Events) {
		return fmt.Errorf("event index %d out of range (have %d)", index, len(result.Events))
	}
	actual := result.Events[i].Location
	if actual != location {
		return fmt.Errorf("expected event location %q, got %q", location, actual)
	}
	return nil
}

// Helpers

func parseCalendarTable(table *godog.Table) []MockCalendar {
	if len(table.Rows) < 2 {
		return nil
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	var calendars []MockCalendar
	for _, row := range table.Rows[1:] {
		calendars = append(calendars, MockCalendar{
			ID:          calCellValue(row, headers, "id"),
			Name:        calCellValue(row, headers, "name"),
			Description: calCellValue(row, headers, "description"),
		})
	}
	return calendars
}

func parseEventTable(table *godog.Table, calendarID string) []MockEvent {
	if len(table.Rows) < 2 {
		return nil
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	var events []MockEvent
	for _, row := range table.Rows[1:] {
		start, _ := time.Parse(time.RFC3339, calCellValue(row, headers, "start"))
		end, _ := time.Parse(time.RFC3339, calCellValue(row, headers, "end"))

		events = append(events, MockEvent{
			CalendarID:  calendarID,
			UID:         calCellValue(row, headers, "uid"),
			Summary:     calCellValue(row, headers, "summary"),
			Description: calCellValue(row, headers, "description"),
			Location:    calCellValue(row, headers, "location"),
			Start:       start,
			End:         end,
		})
	}
	return events
}

func calCellValue(row *messages.PickleTableRow, headers map[string]int, key string) string {
	idx, ok := headers[key]
	if !ok || idx >= len(row.Cells) {
		return ""
	}
	return row.Cells[idx].Value
}

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

	// Then steps
	sc.Step(`^I should receive (\d+) calendars?$`, iShouldReceiveNCalendars)
	sc.Step(`^calendar (\d+) should have name "([^"]*)"$`, calendarShouldHaveName)
	sc.Step(`^calendar (\d+) should have description "([^"]*)"$`, calendarShouldHaveDescription)
	sc.Step(`^I should receive (\d+) events?$`, iShouldReceiveNEvents)
	sc.Step(`^event (\d+) should have summary "([^"]*)"$`, eventShouldHaveSummary)
	sc.Step(`^event (\d+) should have location "([^"]*)"$`, eventShouldHaveLocation)
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

// Then steps

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

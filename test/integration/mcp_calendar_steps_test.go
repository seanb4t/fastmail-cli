//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"

	"github.com/seanb4t/fastmail-cli/mcp"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func registerMCPCalendarSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the following MCP calendars exist:$`, theFollowingMCPCalendarsExist)
	sc.Step(`^the following MCP calendar events exist:$`, theFollowingMCPCalendarEventsExist)
	sc.Step(`^I call the "([^"]*)" calendar tool$`, iCallCalendarToolNoArgs)
	sc.Step(`^I call the "([^"]*)" calendar tool with:$`, iCallCalendarToolWith)
}

// mcpCalendarData holds calendar test data for MCP calendar tool scenarios.
// Uses a separate DomainData key ("mcp_calendar") to avoid collisions with
// the service-layer calendar tests that use "calendar".
type mcpCalendarData struct {
	Calendars []fastmail.Calendar
	Events    []fastmail.Event
}

func getMCPCalendarData(w *World) *mcpCalendarData {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	data, ok := w.DomainData["mcp_calendar"].(*mcpCalendarData)
	if !ok {
		data = &mcpCalendarData{}
		w.DomainData["mcp_calendar"] = data
	}
	return data
}

func theFollowingMCPCalendarsExist(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getMCPCalendarData(w)

	if len(table.Rows) < 2 {
		return ctx, nil
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	for _, row := range table.Rows[1:] {
		cal := fastmail.Calendar{
			ID:          mcpCalCellValue(row, headers, "id"),
			Name:        mcpCalCellValue(row, headers, "name"),
			Color:       mcpCalCellValue(row, headers, "color"),
			Description: mcpCalCellValue(row, headers, "description"),
		}
		if mcpCalCellValue(row, headers, "is_default") == "true" {
			cal.IsDefaultCalendar = true
		}
		if mcpCalCellValue(row, headers, "read_only") == "true" {
			cal.ReadOnly = true
		}
		data.Calendars = append(data.Calendars, cal)
	}
	return ctx, nil
}

func theFollowingMCPCalendarEventsExist(ctx context.Context, table *godog.Table) (context.Context, error) {
	w := WorldFromContext(ctx)
	data := getMCPCalendarData(w)

	if len(table.Rows) < 2 {
		return ctx, nil
	}

	headers := make(map[string]int)
	for i, cell := range table.Rows[0].Cells {
		headers[cell.Value] = i
	}

	for _, row := range table.Rows[1:] {
		start, _ := time.Parse(time.RFC3339, mcpCalCellValue(row, headers, "start"))
		end, _ := time.Parse(time.RFC3339, mcpCalCellValue(row, headers, "end"))

		evt := fastmail.Event{
			ID:          mcpCalCellValue(row, headers, "id"),
			CalendarID:  mcpCalCellValue(row, headers, "calendar_id"),
			Summary:     mcpCalCellValue(row, headers, "summary"),
			Description: mcpCalCellValue(row, headers, "description"),
			Location:    mcpCalCellValue(row, headers, "location"),
			Start:       start,
			End:         end,
		}
		data.Events = append(data.Events, evt)
	}
	return ctx, nil
}

func iCallCalendarToolNoArgs(ctx context.Context, toolName string) (context.Context, error) {
	return callCalendarTool(ctx, toolName, map[string]any{})
}

func iCallCalendarToolWith(ctx context.Context, toolName string, table *godog.Table) (context.Context, error) {
	args := tableToArgs(table)
	return callCalendarTool(ctx, toolName, args)
}

func callCalendarTool(ctx context.Context, toolName string, args map[string]any) (context.Context, error) {
	w := WorldFromContext(ctx)
	calData := getMCPCalendarData(w)

	client := fastmail.NewClient(w.SessionServer.URL, "test-token")

	cfg := mcp.ToolsConfig{
		Client: client,
		Calendar: &mcp.CalendarAdapter{
			ListCalendarsFunc: func(_ context.Context) ([]fastmail.Calendar, error) {
				return calData.Calendars, nil
			},
			ListEventsFunc: func(_ context.Context, calID string, _, _ time.Time) ([]fastmail.Event, error) {
				var result []fastmail.Event
				for _, e := range calData.Events {
					if calID == "" || e.CalendarID == calID {
						result = append(result, e)
					}
				}
				return result, nil
			},
			GetEventFunc: func(_ context.Context, eventID string) (*fastmail.Event, error) {
				for i := range calData.Events {
					if calData.Events[i].ID == eventID {
						e := calData.Events[i]
						return &e, nil
					}
				}
				return nil, fmt.Errorf("event not found: %s", eventID)
			},
			CreateEventFunc: func(_ context.Context, _ *fastmail.Event) error {
				return nil
			},
			UpdateEventFunc: func(_ context.Context, _ *fastmail.Event) error {
				return nil
			},
			DeleteEventFunc: func(_ context.Context, _ string) error {
				return nil
			},
		},
	}

	handler := resolveToolHandler(toolName, cfg)
	if handler == nil {
		return ctx, fmt.Errorf("unknown tool: %s", toolName)
	}

	result, err := handler(ctx, args)
	// Normalize struct results (e.g., *mcp.Event, []*mcp.Event) to
	// map[string]any / []map[string]any via JSON round-trip so the
	// generic assertion steps can inspect them.
	w.ToolResult = normalizeToolResult(result)
	w.ToolError = err
	return ctx, nil
}

// normalizeToolResult converts typed results (structs, struct slices) to
// map[string]any / []map[string]any via JSON round-trip.
func normalizeToolResult(result any) any {
	if result == nil {
		return nil
	}
	switch result.(type) {
	case map[string]any, []map[string]any:
		return result
	}
	b, err := json.Marshal(result)
	if err != nil {
		return result
	}
	// Try slice first.
	var sliceResult []map[string]any
	if err := json.Unmarshal(b, &sliceResult); err == nil {
		return sliceResult
	}
	// Try single map.
	var mapResult map[string]any
	if err := json.Unmarshal(b, &mapResult); err == nil {
		return mapResult
	}
	return result
}

func mcpCalCellValue(row *messages.PickleTableRow, headers map[string]int, key string) string {
	idx, ok := headers[key]
	if !ok || idx >= len(row.Cells) {
		return ""
	}
	return row.Cells[idx].Value
}

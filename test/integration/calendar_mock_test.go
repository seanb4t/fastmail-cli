//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http/httptest"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

// MockCalendar holds calendar test data for the CalDAV mock server.
type MockCalendar struct {
	ID          string
	Name        string
	Description string
}

// MockEvent holds event test data for the CalDAV mock server.
type MockEvent struct {
	CalendarID  string
	UID         string
	Summary     string
	Description string
	Location    string
	Start       time.Time
	End         time.Time
}

// calendarDomainData holds typed state for calendar scenarios.
type calendarDomainData struct {
	Calendars []MockCalendar
	Events    []MockEvent
}

// calendarResult holds captured results from calendar When steps.
type calendarResult struct {
	Calendars    []calendarResultItem
	Events       []eventResultItem
	OperationErr error
}

type calendarResultItem struct {
	ID          string
	Name        string
	Description string
}

type eventResultItem struct {
	UID         string
	Summary     string
	Description string
	Location    string
	Start       time.Time
	End         time.Time
}

func getCalendarData(w *World) *calendarDomainData {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	data, ok := w.DomainData["calendar"].(*calendarDomainData)
	if !ok {
		data = &calendarDomainData{}
		w.DomainData["calendar"] = data
	}
	return data
}

func getCalendarResult(w *World) *calendarResult {
	if w.DomainData == nil {
		w.DomainData = make(map[string]any)
	}
	result, ok := w.DomainData["calendar-result"].(*calendarResult)
	if !ok {
		result = &calendarResult{}
		w.DomainData["calendar-result"] = result
	}
	return result
}

// mockCalDAVBackend implements caldav.Backend for testing.
type mockCalDAVBackend struct {
	calendars []MockCalendar
	events    []MockEvent
}

func (b *mockCalDAVBackend) CurrentUserPrincipal(_ context.Context) (string, error) {
	return "/user/", nil
}

func (b *mockCalDAVBackend) CalendarHomeSetPath(_ context.Context) (string, error) {
	return "/user/calendars/", nil
}

func (b *mockCalDAVBackend) CreateCalendar(_ context.Context, _ *caldav.Calendar) error {
	return nil
}

func (b *mockCalDAVBackend) ListCalendars(_ context.Context) ([]caldav.Calendar, error) {
	cals := make([]caldav.Calendar, 0, len(b.calendars))
	for _, mc := range b.calendars {
		cals = append(cals, caldav.Calendar{
			Path:                  mc.ID,
			Name:                  mc.Name,
			Description:           mc.Description,
			SupportedComponentSet: []string{"VEVENT"},
		})
	}
	return cals, nil
}

func (b *mockCalDAVBackend) GetCalendar(_ context.Context, path string) (*caldav.Calendar, error) {
	for _, mc := range b.calendars {
		if mc.ID == path {
			return &caldav.Calendar{
				Path:                  mc.ID,
				Name:                  mc.Name,
				Description:           mc.Description,
				SupportedComponentSet: []string{"VEVENT"},
			}, nil
		}
	}
	return nil, fmt.Errorf("calendar not found: %s", path)
}

func (b *mockCalDAVBackend) GetCalendarObject(_ context.Context, path string, _ *caldav.CalendarCompRequest) (*caldav.CalendarObject, error) {
	for _, me := range b.events {
		objPath := me.CalendarID + "/" + me.UID + ".ics"
		if objPath == path {
			return &caldav.CalendarObject{
				Path: objPath,
				Data: buildICalendar(&me),
			}, nil
		}
	}
	return nil, fmt.Errorf("calendar object not found: %s", path)
}

func (b *mockCalDAVBackend) ListCalendarObjects(_ context.Context, path string, _ *caldav.CalendarCompRequest) ([]caldav.CalendarObject, error) {
	return b.objectsForCalendar(path), nil
}

func (b *mockCalDAVBackend) QueryCalendarObjects(_ context.Context, path string, query *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	objects := b.objectsForCalendar(path)
	if query != nil {
		filtered, err := caldav.Filter(query, objects)
		if err != nil {
			return nil, err
		}
		return filtered, nil
	}
	return objects, nil
}

func (b *mockCalDAVBackend) PutCalendarObject(_ context.Context, _ string, _ *ical.Calendar, _ *caldav.PutCalendarObjectOptions) (*caldav.CalendarObject, error) {
	return &caldav.CalendarObject{}, nil //nolint:exhaustruct // test
}

func (b *mockCalDAVBackend) DeleteCalendarObject(_ context.Context, _ string) error {
	return nil
}

func (b *mockCalDAVBackend) objectsForCalendar(calendarPath string) []caldav.CalendarObject {
	var objects []caldav.CalendarObject
	for i := range b.events {
		if b.events[i].CalendarID == calendarPath {
			objects = append(objects, caldav.CalendarObject{
				Path: calendarPath + "/" + b.events[i].UID + ".ics",
				Data: buildICalendar(&b.events[i]),
			})
		}
	}
	return objects
}

func buildICalendar(me *MockEvent) *ical.Calendar {
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//test//EN")

	vevent := ical.NewComponent(ical.CompEvent)
	vevent.Props.SetText(ical.PropUID, me.UID)
	vevent.Props.SetText(ical.PropSummary, me.Summary)
	vevent.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
	vevent.Props.SetDateTime(ical.PropDateTimeStart, me.Start)
	vevent.Props.SetDateTime(ical.PropDateTimeEnd, me.End)

	if me.Description != "" {
		vevent.Props.SetText(ical.PropDescription, me.Description)
	}
	if me.Location != "" {
		vevent.Props.SetText(ical.PropLocation, me.Location)
	}

	cal.Children = append(cal.Children, vevent)
	return cal
}

// newMockCalDAVServer creates an httptest.Server backed by the caldav.Handler
// and a mock Backend populated from the given domain data.
func newMockCalDAVServer(data *calendarDomainData) *httptest.Server {
	backend := &mockCalDAVBackend{
		calendars: data.Calendars,
		events:    data.Events,
	}
	handler := &caldav.Handler{
		Backend: backend,
		Prefix:  "",
	}
	return httptest.NewServer(handler)
}

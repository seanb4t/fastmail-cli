package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/internal/auth"
	"github.com/seanb4t/fastmail-cli/internal/config"
	"github.com/seanb4t/fastmail-cli/internal/dav"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// newCalendarCommand creates the calendar command with subcommands.
func newCalendarCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Calendar operations",
		Long:  "Commands for managing Fastmail calendar events via CalDAV.",
	}

	cmd.AddCommand(newCalendarListCommand())
	cmd.AddCommand(newCalendarShowCommand())
	cmd.AddCommand(newCalendarCreateCommand())
	cmd.AddCommand(newCalendarUpdateCommand())
	cmd.AddCommand(newCalendarDeleteCommand())
	cmd.AddCommand(newCalendarCalendarsCommand())

	return cmd
}

// newCalendarListCommand creates the calendar list command.
func newCalendarListCommand() *cobra.Command {
	var startStr, endStr, calendarID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List events",
		Long:  "List calendar events within a date range. Defaults to today.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			svc, err := createCalendarService()
			if err != nil {
				return err
			}

			start, end, err := parseDateRange(startStr, endStr)
			if err != nil {
				return err
			}

			ctx := context.Background()

			if calendarID == "" {
				calendars, err := svc.ListCalendars(ctx)
				if err != nil {
					return fmt.Errorf("listing calendars: %w", err)
				}
				if len(calendars) == 0 {
					return fmt.Errorf("no calendars found")
				}
				calendarID = calendars[0].ID
			}

			events, err := svc.ListEvents(ctx, calendarID, start, end)
			if err != nil {
				return fmt.Errorf("listing events: %w", err)
			}

			return outputEvents(cmd, events)
		},
	}

	cmd.Flags().StringVar(&startStr, "start", "", "start date (RFC3339 or YYYY-MM-DD, default: today 00:00)")
	cmd.Flags().StringVar(&endStr, "end", "", "end date (RFC3339 or YYYY-MM-DD, default: today 23:59)")
	cmd.Flags().StringVar(&calendarID, "calendar", "", "calendar ID (default: first calendar)")

	return cmd
}

// newCalendarShowCommand creates the calendar show command.
func newCalendarShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show ID",
		Short: "Show event details",
		Long:  "Display detailed information about a single calendar event.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			svc, err := createCalendarService()
			if err != nil {
				return err
			}

			ctx := context.Background()
			event, err := svc.GetEvent(ctx, id)
			if err != nil {
				return fmt.Errorf("getting event: %w", err)
			}

			return outputEvent(cmd, event)
		},
	}

	return cmd
}

// newCalendarCreateCommand creates the calendar create command.
func newCalendarCreateCommand() *cobra.Command {
	var summary, description, location, calendarID, startStr, endStr string
	var allDay bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an event",
		Long:  "Create a new calendar event.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			start, err := parseDateTime(startStr)
			if err != nil {
				return fmt.Errorf("parsing --start: %w", err)
			}

			end, err := parseDateTime(endStr)
			if err != nil {
				return fmt.Errorf("parsing --end: %w", err)
			}

			svc, err := createCalendarService()
			if err != nil {
				return err
			}

			event := &fastmail.Event{
				CalendarID:  calendarID,
				UID:         uuid.New().String(),
				Summary:     summary,
				Description: description,
				Location:    location,
				Start:       start,
				End:         end,
				AllDay:      allDay,
				Status:      "CONFIRMED",
			}

			ctx := context.Background()
			if err := svc.CreateEvent(ctx, event); err != nil {
				return fmt.Errorf("creating event: %w", err)
			}

			return outputCalendarStatus(cmd, event.UID, "Created")
		},
	}

	cmd.Flags().StringVar(&summary, "summary", "", "event summary/title (required)")
	cmd.Flags().StringVar(&startStr, "start", "", "start time (RFC3339 or YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&endStr, "end", "", "end time (RFC3339 or YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&calendarID, "calendar", "", "calendar ID (required)")
	cmd.Flags().StringVar(&description, "description", "", "event description")
	cmd.Flags().StringVar(&location, "location", "", "event location")
	cmd.Flags().BoolVar(&allDay, "all-day", false, "create as all-day event")
	_ = cmd.MarkFlagRequired("summary")
	_ = cmd.MarkFlagRequired("start")
	_ = cmd.MarkFlagRequired("end")
	_ = cmd.MarkFlagRequired("calendar")

	return cmd
}

// newCalendarUpdateCommand creates the calendar update command.
func newCalendarUpdateCommand() *cobra.Command {
	var summary, description, location, startStr, endStr string

	cmd := &cobra.Command{
		Use:   "update ID",
		Short: "Update an event",
		Long:  "Update an existing calendar event. Only provided fields are changed.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			svc, err := createCalendarService()
			if err != nil {
				return err
			}

			ctx := context.Background()

			existing, err := svc.GetEvent(ctx, id)
			if err != nil {
				return fmt.Errorf("getting event: %w", err)
			}

			if cmd.Flags().Changed("summary") {
				existing.Summary = summary
			}
			if cmd.Flags().Changed("description") {
				existing.Description = description
			}
			if cmd.Flags().Changed("location") {
				existing.Location = location
			}
			if cmd.Flags().Changed("start") {
				start, err := parseDateTime(startStr)
				if err != nil {
					return fmt.Errorf("parsing --start: %w", err)
				}
				existing.Start = start
			}
			if cmd.Flags().Changed("end") {
				end, err := parseDateTime(endStr)
				if err != nil {
					return fmt.Errorf("parsing --end: %w", err)
				}
				existing.End = end
			}

			if err := svc.UpdateEvent(ctx, existing); err != nil {
				return fmt.Errorf("updating event: %w", err)
			}

			return outputCalendarStatus(cmd, id, "Updated")
		},
	}

	cmd.Flags().StringVar(&summary, "summary", "", "new event summary/title")
	cmd.Flags().StringVar(&description, "description", "", "new event description")
	cmd.Flags().StringVar(&location, "location", "", "new event location")
	cmd.Flags().StringVar(&startStr, "start", "", "new start time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringVar(&endStr, "end", "", "new end time (RFC3339 or YYYY-MM-DD)")

	return cmd
}

// newCalendarDeleteCommand creates the calendar delete command.
func newCalendarDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete ID",
		Short: "Delete an event",
		Long:  "Permanently delete a calendar event.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !force {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete event %s? Use --force to confirm.\n", id)
				return fmt.Errorf("deletion canceled: use --force to confirm")
			}

			svc, err := createCalendarService()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := svc.DeleteEvent(ctx, id); err != nil {
				return fmt.Errorf("deleting event: %w", err)
			}

			return outputCalendarStatus(cmd, id, "Deleted")
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")

	return cmd
}

// newCalendarCalendarsCommand creates the calendar calendars command.
func newCalendarCalendarsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calendars",
		Short: "List calendars",
		Long:  "List all available calendars.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			svc, err := createCalendarService()
			if err != nil {
				return err
			}

			ctx := context.Background()
			calendars, err := svc.ListCalendars(ctx)
			if err != nil {
				return fmt.Errorf("listing calendars: %w", err)
			}

			return outputCalendars(cmd, calendars)
		},
	}

	return cmd
}

// createCalendarService creates a CalendarService from config.
func createCalendarService() (*fastmail.CalendarService, error) {
	configPath := GetConfigPath()
	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	store := auth.NewStore(configPath)
	setStoreWarningWriter(store, os.Stderr)

	token, err := store.GetToken()
	if err != nil {
		return nil, fmt.Errorf("getting token: %w (run 'fastmail auth login' first)", err)
	}

	username := cfg.CardDAVUsername
	if username == "" {
		return nil, fmt.Errorf("carddav_username not configured (set FASTMAIL_CARDDAV_USERNAME or add to config)")
	}

	davClient := dav.NewClient(cfg.CalDAVEndpoint, username, token)
	return fastmail.NewCalendarService(davClient), nil
}

// parseDateRange parses start/end date strings, defaulting to today.
func parseDateRange(startStr, endStr string) (time.Time, time.Time, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	start := todayStart
	end := todayEnd

	if startStr != "" {
		var err error
		start, err = parseDateTime(startStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parsing --start: %w", err)
		}
	}

	if endStr != "" {
		var err error
		end, err = parseDateTime(endStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parsing --end: %w", err)
		}
	}

	return start, end, nil
}

// parseDateTime parses a date/time string in RFC3339 or YYYY-MM-DD format.
func parseDateTime(s string) (time.Time, error) {
	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid date format %q (use RFC3339 or YYYY-MM-DD)", s)
}

// outputEvents writes the event list to output.
func outputEvents(cmd *cobra.Command, events []fastmail.Event) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(events)
	}

	for _, e := range events {
		startFmt := e.Start.Format("2006-01-02 15:04")
		endFmt := e.End.Format("15:04")
		if e.AllDay {
			startFmt = e.Start.Format("2006-01-02")
			endFmt = "all-day"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s-%s  %s\n",
			e.ID, startFmt, endFmt, e.Summary)
	}

	return nil
}

// outputEvent writes a single event to output.
func outputEvent(cmd *cobra.Command, event *fastmail.Event) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(event)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", event.ID)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Summary:     %s\n", event.Summary)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Calendar:    %s\n", event.CalendarID)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Start:       %s\n", event.Start.Format(time.RFC3339))
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "End:         %s\n", event.End.Format(time.RFC3339))
	if event.AllDay {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "All Day:     yes\n")
	}
	if event.Location != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Location:    %s\n", event.Location)
	}
	if event.Description != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", event.Description)
	}
	if event.Status != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status:      %s\n", event.Status)
	}
	if event.RecurrenceRule != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Recurrence:  %s\n", event.RecurrenceRule)
	}

	return nil
}

// outputCalendars writes the calendar list to output.
func outputCalendars(cmd *cobra.Command, calendars []fastmail.Calendar) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(calendars)
	}

	for _, c := range calendars {
		line := fmt.Sprintf("%s  %s", c.ID, c.Name)
		if c.Description != "" {
			line += fmt.Sprintf("  (%s)", c.Description)
		}
		if c.ReadOnly {
			line += "  [read-only]"
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), line)
	}

	return nil
}

// outputCalendarStatus writes a status update to output.
func outputCalendarStatus(cmd *cobra.Command, id, status string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"id": id, "status": status}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", status, id)
	return nil
}

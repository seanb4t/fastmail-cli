package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/internal/auth"
	"github.com/seanb4t/fastmail-cli/internal/config"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// newCalendarCommand creates the calendar command with subcommands.
func newCalendarCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Calendar operations",
		Long:  "Commands for managing calendar events.",
	}

	cmd.AddCommand(newCalendarListCommand())
	cmd.AddCommand(newCalendarShowCommand())
	cmd.AddCommand(newCalendarCreateCommand())
	cmd.AddCommand(newCalendarUpdateCommand())
	cmd.AddCommand(newCalendarDeleteCommand())
	cmd.AddCommand(newCalendarCalendarsCommand())

	return cmd
}

// createCalendarClient creates a calendar service from config.
func createCalendarClient() (*fastmail.CalendarService, error) {
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

	username := cfg.CalDAVUsername
	if username == "" {
		// Fall back to CardDAV username
		username = cfg.CardDAVUsername
	}
	if username == "" {
		return nil, fmt.Errorf("caldav_username not configured (set FASTMAIL_CALDAV_USERNAME or add to config)")
	}

	return fastmail.NewCalendarService(cfg.CalDAVEndpoint, username, token), nil
}

// newCalendarListCommand creates the calendar list command.
func newCalendarListCommand() *cobra.Command {
	var fromStr, toStr, calendarID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List events",
		Long:  "List calendar events within a date range.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createCalendarClient()
			if err != nil {
				return err
			}

			// Default: last 7 days to next 30 days
			now := time.Now()
			from := now.AddDate(0, 0, -7)
			to := now.AddDate(0, 0, 30)

			if fromStr != "" {
				from, err = time.Parse("2006-01-02", fromStr)
				if err != nil {
					return fmt.Errorf("invalid --from date: %w", err)
				}
			}
			if toStr != "" {
				to, err = time.Parse("2006-01-02", toStr)
				if err != nil {
					return fmt.Errorf("invalid --to date: %w", err)
				}
			}

			ctx := context.Background()
			events, err := client.ListEvents(ctx, calendarID, from, to)
			if err != nil {
				return fmt.Errorf("listing events: %w", err)
			}

			return outputEvents(cmd, events)
		},
	}

	cmd.Flags().StringVar(&fromStr, "from", "", "start date (YYYY-MM-DD, default: 7 days ago)")
	cmd.Flags().StringVar(&toStr, "to", "", "end date (YYYY-MM-DD, default: 30 days from now)")
	cmd.Flags().StringVar(&calendarID, "calendar", "", "calendar ID (default: all calendars)")

	return cmd
}

// newCalendarShowCommand creates the calendar show command.
func newCalendarShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show EVENT_ID",
		Short: "Show an event",
		Long:  "Display details of a calendar event.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createCalendarClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			event, err := client.GetEvent(ctx, args[0])
			if err != nil {
				return fmt.Errorf("getting event: %w", err)
			}

			return outputEventDetail(cmd, event)
		},
	}
}

// newCalendarCreateCommand creates the calendar create command.
func newCalendarCreateCommand() *cobra.Command {
	var summary, start, end, location, description, calendarID string
	var allDay bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an event",
		Long:  "Create a new calendar event.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if summary == "" {
				return fmt.Errorf("--summary is required")
			}
			if start == "" {
				return fmt.Errorf("--start is required")
			}
			if end == "" {
				return fmt.Errorf("--end is required")
			}

			layout := "2006-01-02T15:04"
			if allDay {
				layout = "2006-01-02"
			}

			startTime, err := time.Parse(layout, start)
			if err != nil {
				return fmt.Errorf("invalid --start: %w", err)
			}
			endTime, err := time.Parse(layout, end)
			if err != nil {
				return fmt.Errorf("invalid --end: %w", err)
			}

			client, err := createCalendarClient()
			if err != nil {
				return err
			}

			event := &fastmail.Event{
				CalendarID:  calendarID,
				Summary:     summary,
				Start:       startTime,
				End:         endTime,
				Location:    location,
				Description: description,
				AllDay:      allDay,
			}

			ctx := context.Background()
			if err := client.CreateEvent(ctx, event); err != nil {
				return fmt.Errorf("creating event: %w", err)
			}

			return outputEventCreated(cmd, event)
		},
	}

	cmd.Flags().StringVar(&summary, "summary", "", "event title")
	cmd.Flags().StringVar(&start, "start", "", "start time (YYYY-MM-DDTHH:MM or YYYY-MM-DD for all-day)")
	cmd.Flags().StringVar(&end, "end", "", "end time")
	cmd.Flags().StringVar(&location, "location", "", "event location")
	cmd.Flags().StringVar(&description, "description", "", "event description")
	cmd.Flags().StringVar(&calendarID, "calendar", "", "target calendar ID")
	cmd.Flags().BoolVar(&allDay, "all-day", false, "create as all-day event")

	_ = cmd.MarkFlagRequired("summary")
	_ = cmd.MarkFlagRequired("start")
	_ = cmd.MarkFlagRequired("end")

	return cmd
}

// newCalendarUpdateCommand creates the calendar update command.
func newCalendarUpdateCommand() *cobra.Command {
	var summary, start, end, location, description string

	cmd := &cobra.Command{
		Use:   "update EVENT_ID",
		Short: "Update an event",
		Long:  "Update an existing calendar event.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := args[0]

			client, err := createCalendarClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			event, err := client.GetEvent(ctx, eventID)
			if err != nil {
				return fmt.Errorf("getting event: %w", err)
			}

			if summary != "" {
				event.Summary = summary
			}
			if location != "" {
				event.Location = location
			}
			if description != "" {
				event.Description = description
			}
			if start != "" {
				t, err := time.Parse("2006-01-02T15:04", start)
				if err != nil {
					return fmt.Errorf("invalid --start: %w", err)
				}
				event.Start = t
			}
			if end != "" {
				t, err := time.Parse("2006-01-02T15:04", end)
				if err != nil {
					return fmt.Errorf("invalid --end: %w", err)
				}
				event.End = t
			}

			if err := client.UpdateEvent(ctx, event); err != nil {
				return fmt.Errorf("updating event: %w", err)
			}

			return outputEventStatus(cmd, eventID, "updated")
		},
	}

	cmd.Flags().StringVar(&summary, "summary", "", "new event title")
	cmd.Flags().StringVar(&start, "start", "", "new start time (YYYY-MM-DDTHH:MM)")
	cmd.Flags().StringVar(&end, "end", "", "new end time")
	cmd.Flags().StringVar(&location, "location", "", "new location")
	cmd.Flags().StringVar(&description, "description", "", "new description")

	return cmd
}

// newCalendarDeleteCommand creates the calendar delete command.
func newCalendarDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete EVENT_ID",
		Short: "Delete an event",
		Long:  "Delete a calendar event. Requires --force to confirm.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventID := args[0]
			if !force {
				return fmt.Errorf("use --force to confirm deletion")
			}

			client, err := createCalendarClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.DeleteEvent(ctx, eventID); err != nil {
				return fmt.Errorf("deleting event: %w", err)
			}

			return outputEventStatus(cmd, eventID, "deleted")
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "confirm deletion")

	return cmd
}

// newCalendarCalendarsCommand creates the calendar calendars command.
func newCalendarCalendarsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "calendars",
		Short: "List calendars",
		Long:  "List all available calendars.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createCalendarClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			calendars, err := client.ListCalendars(ctx)
			if err != nil {
				return fmt.Errorf("listing calendars: %w", err)
			}

			return outputCalendars(cmd, calendars)
		},
	}
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
		dateStr := e.Start.Format("2006-01-02 15:04")
		if e.AllDay {
			dateStr = e.Start.Format("2006-01-02") + " (all day)"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n", e.ID, dateStr, e.Summary)
	}

	return nil
}

// outputEventDetail writes a single event to output.
func outputEventDetail(cmd *cobra.Command, event *fastmail.Event) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(event)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Summary:  %s\n", event.Summary)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Start:    %s\n", event.Start.Format("2006-01-02 15:04"))
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "End:      %s\n", event.End.Format("2006-01-02 15:04"))
	if event.Location != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Location: %s\n", event.Location)
	}
	if event.Description != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", event.Description)
	}

	return nil
}

// outputEventCreated writes the created event to output.
func outputEventCreated(cmd *cobra.Command, event *fastmail.Event) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"summary": event.Summary, "status": "created"}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Event created: %s\n", event.Summary)

	return nil
}

// outputEventStatus writes a status update to output.
func outputEventStatus(cmd *cobra.Command, eventID, status string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"id": eventID, "status": status}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Event %s: %s\n", status, eventID)

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
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n", c.ID, c.Name)
	}

	return nil
}

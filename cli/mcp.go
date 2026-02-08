package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/mcp"
)

func newMCPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run MCP server",
		Long:  "Run the Model Context Protocol server on stdin/stdout for AI agent integration.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMCPServer(cmd)
		},
	}
	return cmd
}

func runMCPServer(_ *cobra.Command) error {
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		cancel()
	}()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connecting to Fastmail: %w", err)
	}

	// Create optional DAV clients (nil if not configured)
	contactsClient, contactsErr := createContactsClient()
	if contactsErr != nil {
		fmt.Fprintf(os.Stderr, "warning: contacts unavailable: %v\n", contactsErr)
	}
	calendarAdapter := createCalendarAdapter()

	server := mcp.NewServer("fastmail-cli", Version)

	mcp.RegisterMailTools(server, mcp.ToolsConfig{
		Client:   client,
		Contacts: contactsClient,
		Calendar: calendarAdapter,
	})

	registry := mcp.NewResourceRegistry(client, mcp.WithCalendarAdapter(calendarAdapter))
	mcp.RegisterResources(server, registry)

	return server.Run(ctx, os.Stdin, os.Stdout)
}

// createCalendarAdapter creates a CalendarAdapter from the calendar service.
// Returns nil if CalDAV is not configured.
func createCalendarAdapter() *mcp.CalendarAdapter {
	calService, err := createCalendarClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: calendar unavailable: %v\n", err)
		return nil
	}

	return &mcp.CalendarAdapter{
		ListEventsFunc:    calService.ListEvents,
		CreateEventFunc:   calService.CreateEvent,
		GetEventFunc:      calService.GetEvent,
		UpdateEventFunc:   calService.UpdateEvent,
		DeleteEventFunc:   calService.DeleteEvent,
		ListCalendarsFunc: calService.ListCalendars,
	}
}

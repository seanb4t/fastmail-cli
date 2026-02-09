package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/mcp"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// newMCPCommand creates the mcp command that runs the MCP server.
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

// runMCPServer initializes and runs the MCP server.
func runMCPServer(_ *cobra.Command) error {
	// Create fastmail client using existing auth
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// Create context with signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGTERM and SIGINT for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		cancel()
	}()

	// Connect to verify credentials before starting server
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connecting to Fastmail: %w", err)
	}

	// Create MCP server
	server := mcp.NewServer("fastmail-cli", Version)

	// Build tools config
	cfg := mcp.ToolsConfig{
		Client: client,
	}

	// Try to set up contacts (CardDAV) — optional, requires config
	if contactsClient, err := createContactsClient(); err == nil {
		cfg.Contacts = contactsClient
	}

	// Try to set up calendar (CalDAV) — optional, requires config
	if calSvc, err := createCalendarService(); err == nil {
		cfg.Calendar = buildCalendarAdapter(calSvc)
	}

	// Register all tools via the centralized registration
	mcp.RegisterMailTools(server, cfg)

	// Register resources
	registry := mcp.NewResourceRegistry(client)
	for _, res := range registry.List() {
		server.RegisterResource(&res, registry.Read)
	}

	// Run server on stdin/stdout
	return server.Run(ctx, os.Stdin, os.Stdout)
}

// buildCalendarAdapter wraps a CalendarService into a CalendarAdapter for MCP tools.
func buildCalendarAdapter(svc *fastmail.CalendarService) *mcp.CalendarAdapter {
	return &mcp.CalendarAdapter{
		ListCalendarsFunc: svc.ListCalendars,
		ListEventsFunc:    svc.ListEvents,
		GetEventFunc:      svc.GetEvent,
		CreateEventFunc:   svc.CreateEvent,
		UpdateEventFunc:   svc.UpdateEvent,
		DeleteEventFunc:   svc.DeleteEvent,
	}
}

package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/seanb4t/fastmail-cli/mcp"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
	"github.com/spf13/cobra"
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

	// Register tools
	registerMailTools(server, client)
	registerMaskedEmailTools(server, client)

	// Run server on stdin/stdout
	return server.Run(ctx, os.Stdin, os.Stdout)
}

// registerMailTools adds email-related tools to the server.
func registerMailTools(server *mcp.Server, client *fastmail.Client) {
	// mail_list - List emails from a folder
	listTool := mcp.NewTool("mail_list", "List emails from a mailbox folder").
		WithProperty("folder", "string", "Folder name (e.g., 'Inbox', 'Sent', 'Archive')").
		WithProperty("limit", "integer", "Maximum number of emails to return (default: 10)")
	server.RegisterTool(listTool, func(ctx context.Context, args map[string]any) (any, error) {
		folder := "Inbox"
		if f, ok := args["folder"].(string); ok && f != "" {
			folder = f
		}
		var limit uint64 = 10
		if l, ok := args["limit"].(float64); ok && l > 0 {
			limit = uint64(l)
		}
		return client.Mail().List(ctx, folder, limit)
	})

	// mail_search - Search emails
	searchTool := mcp.NewTool("mail_search", "Search emails by query text").
		WithProperty("query", "string", "Search query (e.g., 'from:alice subject:meeting')").
		WithProperty("limit", "integer", "Maximum number of results (default: 10)").
		WithRequired("query")
	server.RegisterTool(searchTool, func(ctx context.Context, args map[string]any) (any, error) {
		query, ok := args["query"].(string)
		if !ok || query == "" {
			return nil, fmt.Errorf("query is required")
		}
		var limit uint64 = 10
		if l, ok := args["limit"].(float64); ok && l > 0 {
			limit = uint64(l)
		}
		return client.Mail().Search(ctx, query, limit)
	})

	// mail_get - Get a single email by ID
	getTool := mcp.NewTool("mail_get", "Get a single email by ID").
		WithProperty("id", "string", "Email ID").
		WithRequired("id")
	server.RegisterTool(getTool, func(ctx context.Context, args map[string]any) (any, error) {
		id, ok := args["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("id is required")
		}
		return client.Mail().Get(ctx, id)
	})

	// mail_send - Send a new email
	sendTool := mcp.NewTool("mail_send", "Send a new email").
		WithProperty("to", "string", "Recipient email address").
		WithProperty("subject", "string", "Email subject").
		WithProperty("body", "string", "Email body text").
		WithRequired("to", "subject", "body")
	server.RegisterTool(sendTool, func(ctx context.Context, args map[string]any) (any, error) {
		to, _ := args["to"].(string)
		subject, _ := args["subject"].(string)
		body, _ := args["body"].(string)

		if to == "" || subject == "" || body == "" {
			return nil, fmt.Errorf("to, subject, and body are required")
		}

		opts := fastmail.SendOptions{
			To:      []fastmail.EmailAddress{{Email: to}},
			Subject: subject,
			Body:    body,
		}
		emailID, err := client.Mail().Send(ctx, opts)
		if err != nil {
			return nil, err
		}
		return map[string]string{"id": emailID, "status": "sent"}, nil
	})

	// mail_move - Move an email to a folder
	moveTool := mcp.NewTool("mail_move", "Move an email to a different folder").
		WithProperty("id", "string", "Email ID").
		WithProperty("folder", "string", "Target folder name").
		WithRequired("id", "folder")
	server.RegisterTool(moveTool, func(ctx context.Context, args map[string]any) (any, error) {
		id, _ := args["id"].(string)
		folder, _ := args["folder"].(string)

		if id == "" || folder == "" {
			return nil, fmt.Errorf("id and folder are required")
		}

		if err := client.Mail().Move(ctx, id, folder); err != nil {
			return nil, err
		}
		return map[string]string{"status": "moved", "folder": folder}, nil
	})

	// mail_delete - Delete an email
	deleteTool := mcp.NewTool("mail_delete", "Delete an email (moves to Trash or permanently deletes if already in Trash)").
		WithProperty("id", "string", "Email ID").
		WithRequired("id")
	server.RegisterTool(deleteTool, func(ctx context.Context, args map[string]any) (any, error) {
		id, _ := args["id"].(string)
		if id == "" {
			return nil, fmt.Errorf("id is required")
		}

		if err := client.Mail().Delete(ctx, id); err != nil {
			return nil, err
		}
		return map[string]string{"status": "deleted"}, nil
	})
}

// registerMaskedEmailTools adds masked email tools to the server.
func registerMaskedEmailTools(server *mcp.Server, client *fastmail.Client) {
	// masked_email_list - List all masked emails
	listTool := mcp.NewTool("masked_email_list", "List all masked email addresses")
	server.RegisterTool(listTool, func(ctx context.Context, _ map[string]any) (any, error) {
		return client.MaskedEmail().List(ctx)
	})

	// masked_email_create - Create a new masked email
	createTool := mcp.NewTool("masked_email_create", "Create a new masked email address").
		WithProperty("domain", "string", "Domain to associate the masked email with").
		WithProperty("description", "string", "Description for the masked email")
	server.RegisterTool(createTool, func(ctx context.Context, args map[string]any) (any, error) {
		domain, _ := args["domain"].(string)
		description, _ := args["description"].(string)

		opts := fastmail.CreateMaskedEmailOptions{
			ForDomain:   domain,
			Description: description,
		}
		return client.MaskedEmail().Create(ctx, opts)
	})

	// masked_email_enable - Enable a masked email
	enableTool := mcp.NewTool("masked_email_enable", "Enable a masked email address").
		WithProperty("id", "string", "Masked email ID").
		WithRequired("id")
	server.RegisterTool(enableTool, func(ctx context.Context, args map[string]any) (any, error) {
		id, _ := args["id"].(string)
		if id == "" {
			return nil, fmt.Errorf("id is required")
		}

		if err := client.MaskedEmail().Enable(ctx, id); err != nil {
			return nil, err
		}
		return map[string]string{"status": "enabled"}, nil
	})

	// masked_email_disable - Disable a masked email
	disableTool := mcp.NewTool("masked_email_disable", "Disable a masked email address").
		WithProperty("id", "string", "Masked email ID").
		WithRequired("id")
	server.RegisterTool(disableTool, func(ctx context.Context, args map[string]any) (any, error) {
		id, _ := args["id"].(string)
		if id == "" {
			return nil, fmt.Errorf("id is required")
		}

		if err := client.MaskedEmail().Disable(ctx, id); err != nil {
			return nil, err
		}
		return map[string]string{"status": "disabled"}, nil
	})

	// masked_email_delete - Delete a masked email
	deleteToolMasked := mcp.NewTool("masked_email_delete", "Permanently delete a masked email address").
		WithProperty("id", "string", "Masked email ID").
		WithRequired("id")
	server.RegisterTool(deleteToolMasked, func(ctx context.Context, args map[string]any) (any, error) {
		id, _ := args["id"].(string)
		if id == "" {
			return nil, fmt.Errorf("id is required")
		}

		if err := client.MaskedEmail().Delete(ctx, id); err != nil {
			return nil, err
		}
		return map[string]string{"status": "deleted"}, nil
	})
}

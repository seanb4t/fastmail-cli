package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// newMailboxCommand creates the mailbox command with list/create/rename/delete subcommands.
func newMailboxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mailbox",
		Short: "Mailbox management",
		Long:  "Commands for listing, creating, renaming, and deleting mailbox folders.",
	}

	cmd.AddCommand(newMailboxListCommand())
	cmd.AddCommand(newMailboxCreateCommand())
	cmd.AddCommand(newMailboxRenameCommand())
	cmd.AddCommand(newMailboxDeleteCommand())

	return cmd
}

// newMailboxListCommand creates the mailbox list command.
func newMailboxListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mailboxes",
		Long:  "List all mailbox folders with unread and total email counts.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			mailboxes, err := client.Mailbox().List(ctx)
			if err != nil {
				return fmt.Errorf("listing mailboxes: %w", err)
			}

			return outputMailboxes(cmd, mailboxes)
		},
	}

	return cmd
}

// newMailboxCreateCommand creates the mailbox create command.
func newMailboxCreateCommand() *cobra.Command {
	var parentID string

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a mailbox",
		Long:  "Create a new mailbox folder with the given name.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			mailbox, err := client.Mailbox().Create(ctx, name, parentID)
			if err != nil {
				return fmt.Errorf("creating mailbox: %w", err)
			}

			return outputMailboxCreated(cmd, mailbox)
		},
	}

	cmd.Flags().StringVar(&parentID, "parent", "", "parent mailbox ID for nested folders")

	return cmd
}

// newMailboxRenameCommand creates the mailbox rename command.
func newMailboxRenameCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "rename ID",
		Short: "Rename a mailbox",
		Long:  "Rename an existing mailbox folder.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if name == "" {
				return fmt.Errorf("--name is required")
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if err := client.Mailbox().Rename(ctx, id, name); err != nil {
				return fmt.Errorf("renaming mailbox: %w", err)
			}

			return outputMailboxRenamed(cmd, id, name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new mailbox name")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// newMailboxDeleteCommand creates the mailbox delete command.
func newMailboxDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete ID",
		Short: "Delete a mailbox",
		Long:  "Delete a mailbox folder by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if err := client.Mailbox().Delete(ctx, id); err != nil {
				return fmt.Errorf("deleting mailbox: %w", err)
			}

			return outputMailboxDeleted(cmd, id)
		},
	}

	return cmd
}

// outputMailboxes writes the mailbox list to output.
func outputMailboxes(cmd *cobra.Command, mailboxes []fastmail.Mailbox) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(mailboxes)
	}

	for _, mb := range mailboxes {
		role := ""
		if mb.Role != "" {
			role = string(mb.Role)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s  %d/%d\n",
			mb.ID, mb.Name, role, mb.UnreadEmails, mb.TotalEmails)
	}

	return nil
}

// outputMailboxCreated writes the create result to output.
func outputMailboxCreated(cmd *cobra.Command, mailbox *fastmail.Mailbox) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]any{
			"id":     mailbox.ID,
			"name":   mailbox.Name,
			"status": "created",
		})
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Mailbox created: %s (%s)\n", mailbox.Name, mailbox.ID)
	return nil
}

// outputMailboxRenamed writes the rename result to output.
func outputMailboxRenamed(cmd *cobra.Command, id, name string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"id": id, "name": name, "status": "renamed"}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Mailbox %s renamed to %s\n", id, name)
	return nil
}

// outputMailboxDeleted writes the delete result to output.
func outputMailboxDeleted(cmd *cobra.Command, id string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"id": id, "status": "deleted"}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Mailbox %s deleted\n", id)
	return nil
}

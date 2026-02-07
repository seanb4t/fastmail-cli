package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// newMailboxCommand creates the mailbox command with subcommands.
func newMailboxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mailbox",
		Short: "Mailbox operations",
		Long:  "Commands for managing email mailboxes (folders).",
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
		Long:  "List all mailboxes with message counts.",
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
	var name string
	var parentID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a mailbox",
		Long:  "Create a new mailbox (folder). Optionally nest under a parent.",
		RunE: func(cmd *cobra.Command, _ []string) error {
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

	cmd.Flags().StringVar(&name, "name", "", "mailbox name (required)")
	cmd.Flags().StringVar(&parentID, "parent", "", "parent mailbox ID for nesting")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// newMailboxRenameCommand creates the mailbox rename command.
func newMailboxRenameCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "rename ID",
		Short: "Rename a mailbox",
		Long:  "Change the display name of a mailbox.",
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

			if err := client.Mailbox().Rename(ctx, id, name); err != nil {
				return fmt.Errorf("renaming mailbox: %w", err)
			}

			return outputMailboxStatus(cmd, id, "Renamed")
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new mailbox name (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// newMailboxDeleteCommand creates the mailbox delete command.
func newMailboxDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete ID",
		Short: "Delete a mailbox",
		Long:  "Permanently delete a mailbox and its contents.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !force {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to delete mailbox %s? Use --force to confirm.\n", id)
				return nil
			}

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

			return outputMailboxStatus(cmd, id, "Deleted")
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation")

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
		role := string(mb.Role)
		if role == "" {
			role = "-"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s  %-20s  %-10s  %d/%d\n",
			mb.ID, mb.Name, role, mb.UnreadEmails, mb.TotalEmails)
	}

	return nil
}

// outputMailboxCreated writes the created mailbox to output.
func outputMailboxCreated(cmd *cobra.Command, mailbox *fastmail.Mailbox) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(mailbox)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created: %s (%s)\n", mailbox.Name, mailbox.ID)
	return nil
}

// outputMailboxStatus writes a status update to output.
func outputMailboxStatus(cmd *cobra.Command, id, status string) error {
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

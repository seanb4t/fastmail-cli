package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// newMaskedEmailCommand creates the masked-email command with subcommands.
func newMaskedEmailCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "masked-email",
		Short: "Masked email operations",
		Long:  "Commands for managing Fastmail masked email addresses.",
	}

	cmd.AddCommand(newMaskedEmailListCommand())
	cmd.AddCommand(newMaskedEmailCreateCommand())
	cmd.AddCommand(newMaskedEmailEnableCommand())
	cmd.AddCommand(newMaskedEmailDisableCommand())
	cmd.AddCommand(newMaskedEmailDeleteCommand())

	return cmd
}

// newMaskedEmailListCommand creates the masked-email list command.
func newMaskedEmailListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List masked emails",
		Long:  "List all masked email addresses.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			emails, err := client.MaskedEmail().List(ctx)
			if err != nil {
				return fmt.Errorf("listing masked emails: %w", err)
			}

			return outputMaskedEmails(cmd, emails)
		},
	}

	return cmd
}

// newMaskedEmailCreateCommand creates the masked-email create command.
func newMaskedEmailCreateCommand() *cobra.Command {
	var forDomain string
	var description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a masked email",
		Long:  "Create a new masked email address.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			opts := fastmail.CreateMaskedEmailOptions{
				ForDomain:   forDomain,
				Description: description,
			}

			maskedEmail, err := client.MaskedEmail().Create(ctx, opts)
			if err != nil {
				return fmt.Errorf("creating masked email: %w", err)
			}

			return outputMaskedEmailCreated(cmd, maskedEmail)
		},
	}

	cmd.Flags().StringVar(&forDomain, "for-domain", "", "domain this masked email is for")
	cmd.Flags().StringVar(&description, "description", "", "description of the masked email")

	return cmd
}

// newMaskedEmailEnableCommand creates the masked-email enable command.
func newMaskedEmailEnableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable ID",
		Short: "Enable a masked email",
		Long:  "Enable a disabled masked email address.",
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

			if err := client.MaskedEmail().Enable(ctx, id); err != nil {
				return fmt.Errorf("enabling masked email: %w", err)
			}

			return outputMaskedEmailStatus(cmd, id, "Enabled")
		},
	}

	return cmd
}

// newMaskedEmailDisableCommand creates the masked-email disable command.
func newMaskedEmailDisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable ID",
		Short: "Disable a masked email",
		Long:  "Disable an enabled masked email address.",
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

			if err := client.MaskedEmail().Disable(ctx, id); err != nil {
				return fmt.Errorf("disabling masked email: %w", err)
			}

			return outputMaskedEmailStatus(cmd, id, "Disabled")
		},
	}

	return cmd
}

// newMaskedEmailDeleteCommand creates the masked-email delete command.
func newMaskedEmailDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete ID",
		Short: "Delete a masked email",
		Long:  "Permanently delete a masked email address.",
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

			if err := client.MaskedEmail().Delete(ctx, id); err != nil {
				return fmt.Errorf("deleting masked email: %w", err)
			}

			return outputMaskedEmailStatus(cmd, id, "Deleted")
		},
	}

	return cmd
}

// outputMaskedEmails writes the masked email list to output.
func outputMaskedEmails(cmd *cobra.Command, emails []fastmail.MaskedEmail) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(emails)
	}

	// Simple text output
	for _, e := range emails {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  [%s]  %s\n",
			e.ID, e.Email, e.State, e.ForDomain)
	}

	return nil
}

// outputMaskedEmailCreated writes the created masked email to output.
func outputMaskedEmailCreated(cmd *cobra.Command, email *fastmail.MaskedEmail) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(email)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", email.Email)
	return nil
}

// outputMaskedEmailStatus writes a status update to output.
func outputMaskedEmailStatus(cmd *cobra.Command, id, status string) error {
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

package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// newIdentityCommand creates the identity command with subcommands.
func newIdentityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Sender identity operations",
		Long:  "Commands for managing sender identities (name, email, reply-to, signature).",
	}

	cmd.AddCommand(newIdentityListCommand())
	cmd.AddCommand(newIdentitySetCommand())

	return cmd
}

// newIdentityListCommand creates the identity list command.
func newIdentityListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all sender identities",
		Long:  "Display all configured sender identities for the account.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			identities, err := client.Identity().List(ctx)
			if err != nil {
				return fmt.Errorf("listing identities: %w", err)
			}

			return outputIdentityList(cmd, identities)
		},
	}
}

// newIdentitySetCommand creates the identity set command.
func newIdentitySetCommand() *cobra.Command {
	var (
		name      string
		replyTo   string
		signature string
	)

	cmd := &cobra.Command{
		Use:   "set ID",
		Short: "Update a sender identity",
		Long:  "Update the name, reply-to address, or text signature of a sender identity.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			opts := fastmail.UpdateIdentityOptions{}
			hasChange := false

			if cmd.Flags().Changed("name") {
				opts.Name = &name
				hasChange = true
			}
			if cmd.Flags().Changed("reply-to") {
				opts.ReplyTo = []fastmail.EmailAddress{{Email: replyTo}}
				hasChange = true
			}
			if cmd.Flags().Changed("signature") {
				opts.TextSignature = &signature
				hasChange = true
			}

			if !hasChange {
				return fmt.Errorf("at least one of --name, --reply-to, or --signature must be specified")
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if err := client.Identity().Update(ctx, id, opts); err != nil {
				return fmt.Errorf("updating identity: %w", err)
			}

			return outputIdentityUpdated(cmd, id)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "display name for the identity")
	cmd.Flags().StringVar(&replyTo, "reply-to", "", "reply-to email address")
	cmd.Flags().StringVar(&signature, "signature", "", "text signature")

	return cmd
}

// outputIdentityList writes the identity list to output.
func outputIdentityList(cmd *cobra.Command, identities []fastmail.Identity) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := make([]map[string]any, len(identities))
		for i, id := range identities {
			entry := map[string]any{
				"id":             id.ID,
				"name":           id.Name,
				"email":          id.Email,
				"text_signature": id.TextSignature,
				"html_signature": id.HTMLSignature,
				"may_delete":     id.MayDelete,
			}
			if len(id.ReplyTo) > 0 {
				replyTo := make([]map[string]string, len(id.ReplyTo))
				for j, addr := range id.ReplyTo {
					replyTo[j] = map[string]string{"name": addr.Name, "email": addr.Email}
				}
				entry["reply_to"] = replyTo
			}
			if len(id.Bcc) > 0 {
				bcc := make([]map[string]string, len(id.Bcc))
				for j, addr := range id.Bcc {
					bcc[j] = map[string]string{"name": addr.Name, "email": addr.Email}
				}
				entry["bcc"] = bcc
			}
			result[i] = entry
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	w := cmd.OutOrStdout()
	for _, id := range identities {
		_, _ = fmt.Fprintf(w, "%-20s %-25s %s", id.ID, id.Name, id.Email)
		if id.TextSignature != "" {
			sig := truncate(id.TextSignature, 40)
			_, _ = fmt.Fprintf(w, "  [sig: %s]", sig)
		}
		_, _ = fmt.Fprintln(w)
	}

	return nil
}

// outputIdentityUpdated writes the update confirmation to output.
func outputIdentityUpdated(cmd *cobra.Command, id string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"id": id, "status": "updated"}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Identity updated: %s\n", id)
	return nil
}

// truncate shortens a string to maxLen, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

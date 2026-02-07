package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func newIdentityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Identity operations",
		Long:  "Commands for managing sender identities.",
	}

	cmd.AddCommand(newIdentityListCommand())
	return cmd
}

func newIdentityListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List sender identities",
		Long:  "List all configured sender identities.",
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

			return outputIdentities(cmd, identities)
		},
	}
}

func outputIdentities(cmd *cobra.Command, identities []fastmail.Identity) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(identities)
	}

	for _, id := range identities {
		deletable := ""
		if !id.MayDelete {
			deletable = " [locked]"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s <%s>%s\n", id.ID, id.Name, id.Email, deletable)
	}
	return nil
}

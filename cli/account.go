package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// newAccountCommand creates the account command with subcommands.
func newAccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Account operations",
		Long:  "Commands for viewing account information such as storage quota.",
	}

	cmd.AddCommand(newAccountQuotaCommand())

	return cmd
}

// newAccountQuotaCommand creates the account quota command.
func newAccountQuotaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quota",
		Short: "Show storage quota",
		Long:  "Display the current storage quota usage for the account.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			quota, err := client.Quota().Get(ctx)
			if err != nil {
				return fmt.Errorf("getting quota: %w", err)
			}

			return outputQuota(cmd, quota)
		},
	}

	return cmd
}

// outputQuota writes the quota information to output.
func outputQuota(cmd *cobra.Command, quota *fastmail.QuotaInfo) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(quota)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Storage: %s / %s (%.1f%%)\n",
		fastmail.FormatSize(quota.Used),
		fastmail.FormatSize(quota.Limit),
		quota.UsedPercent,
	)

	return nil
}

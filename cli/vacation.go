package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func newVacationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vacation",
		Short: "Vacation response operations",
		Long:  "Commands for managing vacation auto-reply settings.",
	}

	cmd.AddCommand(newVacationShowCommand())
	cmd.AddCommand(newVacationSetCommand())

	return cmd
}

func newVacationShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show vacation status",
		Long:  "Display current vacation auto-reply settings.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			vacation, err := client.Vacation().Get(ctx)
			if err != nil {
				return fmt.Errorf("getting vacation status: %w", err)
			}

			return outputVacation(cmd, vacation)
		},
	}
}

func newVacationSetCommand() *cobra.Command {
	var enable, disable bool
	var fromDate, toDate, subject, textBody string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set vacation response",
		Long:  "Configure vacation auto-reply settings.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := fastmail.SetVacationOptions{
				FromDate: fromDate,
				ToDate:   toDate,
				Subject:  subject,
				TextBody: textBody,
			}

			if enable {
				t := true
				opts.IsEnabled = &t
			}
			if disable {
				f := false
				opts.IsEnabled = &f
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if err := client.Vacation().Set(ctx, opts); err != nil {
				return fmt.Errorf("setting vacation: %w", err)
			}

			if IsQuiet() {
				return nil
			}
			if IsJSONOutput() {
				result := map[string]string{"status": "updated"}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Vacation response updated")
			return nil
		},
	}

	cmd.Flags().BoolVar(&enable, "enable", false, "enable vacation response")
	cmd.Flags().BoolVar(&disable, "disable", false, "disable vacation response")
	cmd.Flags().StringVar(&fromDate, "from", "", "start date (RFC3339)")
	cmd.Flags().StringVar(&toDate, "to", "", "end date (RFC3339)")
	cmd.Flags().StringVar(&subject, "subject", "", "auto-reply subject")
	cmd.Flags().StringVar(&textBody, "body", "", "auto-reply body text")

	return cmd
}

func outputVacation(cmd *cobra.Command, v *fastmail.Vacation) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}

	status := "disabled"
	if v.IsEnabled {
		status = "enabled"
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status:  %s\n", status)
	if v.FromDate != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "From:    %s\n", v.FromDate)
	}
	if v.ToDate != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "To:      %s\n", v.ToDate)
	}
	if v.Subject != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Subject: %s\n", v.Subject)
	}
	if v.TextBody != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", v.TextBody)
	}
	return nil
}

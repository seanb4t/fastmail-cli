package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// newVacationCommand creates the vacation command with subcommands.
func newVacationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vacation",
		Short: "Vacation response (out-of-office) operations",
		Long:  "Commands for managing vacation/out-of-office auto-reply settings.",
	}

	cmd.AddCommand(newVacationStatusCommand())
	cmd.AddCommand(newVacationEnableCommand())
	cmd.AddCommand(newVacationDisableCommand())

	return cmd
}

// newVacationStatusCommand creates the vacation status command.
func newVacationStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show vacation response status",
		Long:  "Display the current vacation/out-of-office auto-reply settings.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			status, err := client.Vacation().GetStatus(ctx)
			if err != nil {
				return fmt.Errorf("getting vacation status: %w", err)
			}

			return outputVacationStatus(cmd, status)
		},
	}
}

// newVacationEnableCommand creates the vacation enable command.
func newVacationEnableCommand() *cobra.Command {
	var (
		subject string
		body    string
		from    string
		to      string
	)

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable vacation response",
		Long:  "Enable the vacation/out-of-office auto-reply with a subject and message body.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			var fromTime, toTime *time.Time
			if from != "" {
				t, err := parseFlexibleDate(from)
				if err != nil {
					return fmt.Errorf("parsing --from date: %w", err)
				}
				fromTime = &t
			}
			if to != "" {
				t, err := parseFlexibleDate(to)
				if err != nil {
					return fmt.Errorf("parsing --to date: %w", err)
				}
				toTime = &t
			}

			if fromTime != nil && toTime != nil && fromTime.After(*toTime) {
				return fmt.Errorf("--from date must be before --to date")
			}

			if err := client.Vacation().Enable(ctx, subject, body, fromTime, toTime); err != nil {
				return fmt.Errorf("enabling vacation response: %w", err)
			}

			return outputVacationEnabled(cmd)
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "auto-reply subject line (required)")
	cmd.Flags().StringVar(&body, "body", "", "auto-reply message body (required)")
	cmd.Flags().StringVar(&from, "from", "", "start date (RFC3339 or YYYY-MM-DD, optional)")
	cmd.Flags().StringVar(&to, "to", "", "end date (RFC3339 or YYYY-MM-DD, optional)")

	_ = cmd.MarkFlagRequired("subject")
	_ = cmd.MarkFlagRequired("body")

	return cmd
}

// newVacationDisableCommand creates the vacation disable command.
func newVacationDisableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable vacation response",
		Long:  "Disable the vacation/out-of-office auto-reply.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if err := client.Vacation().Disable(ctx); err != nil {
				return fmt.Errorf("disabling vacation response: %w", err)
			}

			return outputVacationDisabled(cmd)
		},
	}
}

// parseFlexibleDate parses a date string in either RFC3339 or YYYY-MM-DD format.
func parseFlexibleDate(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected RFC3339 (e.g. 2024-01-15T00:00:00Z) or YYYY-MM-DD (e.g. 2024-01-15): %w", err)
	}
	return t, nil
}

// outputVacationStatus writes the vacation status to output.
func outputVacationStatus(cmd *cobra.Command, status *fastmail.VacationStatus) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]any{
			"is_enabled": status.IsEnabled,
			"subject":    status.Subject,
			"text_body":  status.TextBody,
		}
		if status.FromDate != nil {
			result["from_date"] = status.FromDate.Format(time.RFC3339)
		}
		if status.ToDate != nil {
			result["to_date"] = status.ToDate.Format(time.RFC3339)
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	w := cmd.OutOrStdout()
	if status.IsEnabled {
		_, _ = fmt.Fprintf(w, "Vacation response: ENABLED\n")
		if status.Subject != "" {
			_, _ = fmt.Fprintf(w, "Subject: %s\n", status.Subject)
		}
		if status.TextBody != "" {
			_, _ = fmt.Fprintf(w, "Message: %s\n", status.TextBody)
		}
		if status.FromDate != nil {
			_, _ = fmt.Fprintf(w, "From:    %s\n", status.FromDate.Format("2006-01-02"))
		}
		if status.ToDate != nil {
			_, _ = fmt.Fprintf(w, "To:      %s\n", status.ToDate.Format("2006-01-02"))
		}
	} else {
		_, _ = fmt.Fprintf(w, "Vacation response: DISABLED\n")
	}

	return nil
}

// outputVacationEnabled writes the enable confirmation to output.
func outputVacationEnabled(cmd *cobra.Command) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"status": "enabled"}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Vacation response enabled")
	return nil
}

// outputVacationDisabled writes the disable confirmation to output.
func outputVacationDisabled(cmd *cobra.Command) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"status": "disabled"}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Vacation response disabled")
	return nil
}

package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

func newThreadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thread",
		Short: "Thread operations",
		Long:  "Commands for viewing email threads/conversations.",
	}

	cmd.AddCommand(newThreadShowCommand())
	return cmd
}

func newThreadShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show THREAD_ID",
		Short: "Show a thread",
		Long:  "Display all emails in a thread/conversation.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			threadID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			emails, err := client.Thread().Get(ctx, threadID)
			if err != nil {
				return fmt.Errorf("getting thread: %w", err)
			}

			return outputThread(cmd, threadID, emails)
		},
	}
}

func outputThread(cmd *cobra.Command, threadID string, emails []fastmail.Email) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]any{
			"threadId": threadID,
			"emails":   emails,
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Thread %s (%d emails)\n\n", threadID, len(emails))
	for i, e := range emails {
		from := e.From.Email
		if e.From.Name != "" {
			from = e.From.Name
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%d. %s  %s  %s\n",
			i+1, e.ReceivedAt.Format("2006-01-02 15:04"), from, e.Subject)
	}
	return nil
}

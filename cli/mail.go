package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/internal/auth"
	"github.com/seanb4t/fastmail-cli/internal/config"
	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// newMailCommand creates the mail command with list/send/reply subcommands.
func newMailCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mail",
		Short: "Email operations",
		Long:  "Commands for reading, sending, and managing email.",
	}

	cmd.AddCommand(newMailListCommand())
	cmd.AddCommand(newMailSendCommand())
	cmd.AddCommand(newMailReplyCommand())

	return cmd
}

// newMailListCommand creates the mail list command.
func newMailListCommand() *cobra.Command {
	var limit uint64
	var folder string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List emails",
		Long:  "List emails from a mailbox folder.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			emails, err := client.Mail().List(ctx, folder, limit)
			if err != nil {
				return fmt.Errorf("listing emails: %w", err)
			}

			return outputEmails(cmd, emails)
		},
	}

	cmd.Flags().Uint64VarP(&limit, "limit", "n", 10, "maximum emails to return")
	cmd.Flags().StringVarP(&folder, "folder", "f", "Inbox", "mailbox folder name")

	return cmd
}

// newMailSendCommand creates the mail send command.
func newMailSendCommand() *cobra.Command {
	var to []string
	var cc []string
	var bcc []string
	var subject string
	var body string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send an email",
		Long:  "Compose and send a new email.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(to) == 0 {
				return fmt.Errorf("at least one --to recipient is required")
			}
			if subject == "" {
				return fmt.Errorf("--subject is required")
			}
			if body == "" {
				return fmt.Errorf("--body is required")
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			opts := fastmail.SendOptions{
				To:      parseAddresses(to),
				Cc:      parseAddresses(cc),
				Bcc:     parseAddresses(bcc),
				Subject: subject,
				Body:    body,
			}

			emailID, err := client.Mail().Send(ctx, opts)
			if err != nil {
				return fmt.Errorf("sending email: %w", err)
			}

			return outputSendResult(cmd, emailID)
		},
	}

	cmd.Flags().StringArrayVar(&to, "to", nil, "recipient email address (can be repeated)")
	cmd.Flags().StringArrayVar(&cc, "cc", nil, "CC recipient (can be repeated)")
	cmd.Flags().StringArrayVar(&bcc, "bcc", nil, "BCC recipient (can be repeated)")
	cmd.Flags().StringVar(&subject, "subject", "", "email subject")
	cmd.Flags().StringVar(&body, "body", "", "email body text")

	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("subject")
	_ = cmd.MarkFlagRequired("body")

	return cmd
}

// newMailReplyCommand creates the mail reply command.
func newMailReplyCommand() *cobra.Command {
	var body string
	var replyAll bool

	cmd := &cobra.Command{
		Use:   "reply EMAIL_ID",
		Short: "Reply to an email",
		Long:  "Send a reply to an existing email.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			emailID := args[0]

			if body == "" {
				return fmt.Errorf("--body is required")
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			opts := fastmail.ReplyOptions{
				EmailID:  emailID,
				Body:     body,
				ReplyAll: replyAll,
			}

			replyID, err := client.Mail().Reply(ctx, opts)
			if err != nil {
				return fmt.Errorf("sending reply: %w", err)
			}

			return outputSendResult(cmd, replyID)
		},
	}

	cmd.Flags().StringVar(&body, "body", "", "reply body text")
	cmd.Flags().BoolVar(&replyAll, "all", false, "reply to all recipients")

	_ = cmd.MarkFlagRequired("body")

	return cmd
}

// createClient creates and configures a Fastmail client.
func createClient() (*fastmail.Client, error) {
	configPath := GetConfigPath()
	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// Get token from auth store
	store := auth.NewStore(configPath)
	store.DisableKeychain()

	token, err := store.GetToken()
	if err != nil {
		return nil, fmt.Errorf("getting token: %w (run 'fastmail auth login' first)", err)
	}

	return fastmail.NewClient(cfg.Endpoint, token), nil
}

// parseAddresses converts string addresses to EmailAddress structs.
func parseAddresses(addrs []string) []fastmail.EmailAddress {
	result := make([]fastmail.EmailAddress, len(addrs))
	for i, addr := range addrs {
		// Support "Name <email>" format
		if idx := strings.LastIndex(addr, "<"); idx != -1 && strings.HasSuffix(addr, ">") {
			name := strings.TrimSpace(addr[:idx])
			email := addr[idx+1 : len(addr)-1]
			result[i] = fastmail.EmailAddress{Name: name, Email: email}
		} else {
			result[i] = fastmail.EmailAddress{Email: addr}
		}
	}
	return result
}

// outputEmails writes the email list to output.
func outputEmails(cmd *cobra.Command, emails []fastmail.Email) error {
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
		status := ""
		if e.IsRead() {
			status = " [read]"
		}
		if e.IsFlagged() {
			status += " [flagged]"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s%s\n", e.ID, e.Subject, status)
	}

	return nil
}

// outputSendResult writes the send result to output.
func outputSendResult(cmd *cobra.Command, emailID string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"id": emailID, "status": "sent"}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Email sent: %s\n", emailID)
	return nil
}

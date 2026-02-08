package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	cmd.AddCommand(newMailShowCommand())
	cmd.AddCommand(newMailSearchCommand())
	cmd.AddCommand(newMailMoveCommand())
	cmd.AddCommand(newMailDeleteCommand())
	cmd.AddCommand(newMailFlagCommand())
	cmd.AddCommand(newMailThreadCommand())
	cmd.AddCommand(newMailAttachmentsCommand())
	cmd.AddCommand(newMailDownloadCommand())
	cmd.AddCommand(newMailUploadCommand())
	cmd.AddCommand(newMailImportCommand())
	cmd.AddCommand(newMailScheduledCommand())

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
	var schedule string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send an email",
		Long:  "Compose and send a new email. Use --schedule to delay delivery.",
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

			if schedule != "" {
				t, perr := time.Parse(time.RFC3339, schedule)
				if perr != nil {
					return fmt.Errorf("invalid --schedule time (must be RFC3339, e.g. 2024-06-15T14:00:00Z): %w", perr)
				}
				opts.Schedule = &t
			}

			emailID, err := client.Mail().Send(ctx, opts)
			if err != nil {
				return fmt.Errorf("sending email: %w", err)
			}

			if schedule != "" {
				return outputScheduledResult(cmd, emailID, schedule)
			}
			return outputSendResult(cmd, emailID)
		},
	}

	cmd.Flags().StringArrayVar(&to, "to", nil, "recipient email address (can be repeated)")
	cmd.Flags().StringArrayVar(&cc, "cc", nil, "CC recipient (can be repeated)")
	cmd.Flags().StringArrayVar(&bcc, "bcc", nil, "BCC recipient (can be repeated)")
	cmd.Flags().StringVar(&subject, "subject", "", "email subject")
	cmd.Flags().StringVar(&body, "body", "", "email body text")
	cmd.Flags().StringVar(&schedule, "schedule", "", "schedule delivery time (RFC3339, e.g. 2024-06-15T14:00:00Z)")

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
	setStoreWarningWriter(store, os.Stderr)

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

// newMailShowCommand creates the mail show command.
func newMailShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show EMAIL_ID",
		Short: "Show a single email",
		Long:  "Display the details of a single email by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			emailID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			email, err := client.Mail().Get(ctx, emailID)
			if err != nil {
				return fmt.Errorf("getting email: %w", err)
			}

			return outputEmail(cmd, email)
		},
	}

	return cmd
}

// newMailSearchCommand creates the mail search command.
func newMailSearchCommand() *cobra.Command {
	var limit uint64
	var snippets bool

	cmd := &cobra.Command{
		Use:   "search QUERY",
		Short: "Search emails",
		Long:  "Search emails using a text query.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if snippets {
				results, err := client.Mail().SearchWithSnippets(ctx, query, limit)
				if err != nil {
					return fmt.Errorf("searching emails: %w", err)
				}
				return outputSearchResults(cmd, results)
			}

			emails, err := client.Mail().Search(ctx, query, limit)
			if err != nil {
				return fmt.Errorf("searching emails: %w", err)
			}

			return outputEmails(cmd, emails)
		},
	}

	cmd.Flags().Uint64VarP(&limit, "limit", "n", 10, "maximum emails to return")
	cmd.Flags().BoolVarP(&snippets, "snippets", "s", false, "include highlighted search snippets")

	return cmd
}

// outputSearchResults writes search results with snippets to output.
func outputSearchResults(cmd *cobra.Command, results []fastmail.SearchResult) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		type jsonResult struct {
			ID             string `json:"id"`
			Subject        string `json:"subject"`
			Preview        string `json:"preview"`
			SubjectSnippet string `json:"subject_snippet,omitempty"`
			PreviewSnippet string `json:"preview_snippet,omitempty"`
		}

		out := make([]jsonResult, len(results))
		for i, r := range results {
			out[i] = jsonResult{
				ID:             r.Email.ID,
				Subject:        r.Email.Subject,
				Preview:        r.Email.Preview,
				SubjectSnippet: r.SubjectSnippet,
				PreviewSnippet: r.PreviewSnippet,
			}
		}

		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	w := cmd.OutOrStdout()
	for _, r := range results {
		status := ""
		if r.Email.IsRead() {
			status = " [read]"
		}
		if r.Email.IsFlagged() {
			status += " [flagged]"
		}
		_, _ = fmt.Fprintf(w, "%s  %s%s\n", r.Email.ID, r.Email.Subject, status)
		if r.SubjectSnippet != "" {
			_, _ = fmt.Fprintf(w, "  Match: %s\n", r.SubjectSnippet)
		}
		if r.PreviewSnippet != "" {
			_, _ = fmt.Fprintf(w, "  Match: %s\n", r.PreviewSnippet)
		}
	}

	return nil
}

// newMailMoveCommand creates the mail move command.
func newMailMoveCommand() *cobra.Command {
	var folder string

	cmd := &cobra.Command{
		Use:   "move EMAIL_ID",
		Short: "Move an email to a folder",
		Long:  "Move an email to the specified mailbox folder.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			emailID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if err := client.Mail().Move(ctx, emailID, folder); err != nil {
				return fmt.Errorf("moving email: %w", err)
			}

			return outputMoveResult(cmd, emailID, folder)
		},
	}

	cmd.Flags().StringVarP(&folder, "folder", "f", "", "destination mailbox folder")
	_ = cmd.MarkFlagRequired("folder")

	return cmd
}

// newMailDeleteCommand creates the mail delete command.
func newMailDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete EMAIL_ID",
		Short: "Delete an email",
		Long:  "Delete an email by moving it to Trash, or permanently if already in Trash.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			emailID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if err := client.Mail().Delete(ctx, emailID); err != nil {
				return fmt.Errorf("deleting email: %w", err)
			}

			return outputDeleteResult(cmd, emailID)
		},
	}

	return cmd
}

// outputEmail writes a single email's details to output.
func outputEmail(cmd *cobra.Command, email *fastmail.Email) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(email)
	}

	w := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(w, "ID:      %s\n", email.ID)
	_, _ = fmt.Fprintf(w, "Subject: %s\n", email.Subject)
	_, _ = fmt.Fprintf(w, "From:    %s\n", email.From.String())
	if len(email.To) > 0 {
		toStrs := make([]string, len(email.To))
		for i, addr := range email.To {
			toStrs[i] = addr.String()
		}
		_, _ = fmt.Fprintf(w, "To:      %s\n", strings.Join(toStrs, ", "))
	}
	_, _ = fmt.Fprintf(w, "Date:    %s\n", email.ReceivedAt.Format("2006-01-02 15:04:05"))
	if email.Preview != "" {
		_, _ = fmt.Fprintf(w, "Preview: %s\n", email.Preview)
	}
	if len(email.Attachments) > 0 {
		_, _ = fmt.Fprintf(w, "\nAttachments (%d):\n", len(email.Attachments))
		for _, att := range email.Attachments {
			_, _ = fmt.Fprintf(w, "  - %s (%s, %s)\n", att.Name, att.Type, formatSize(att.Size))
		}
	}

	return nil
}

// outputMoveResult writes the move result to output.
func outputMoveResult(cmd *cobra.Command, emailID, folder string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"id": emailID, "status": "moved", "folder": folder}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Email %s moved to %s\n", emailID, folder)
	return nil
}

// outputDeleteResult writes the delete result to output.
func outputDeleteResult(cmd *cobra.Command, emailID string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"id": emailID, "status": "deleted"}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Email %s deleted\n", emailID)
	return nil
}

// newMailThreadCommand creates the mail thread command.
func newMailThreadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thread THREAD_ID",
		Short: "Show all emails in a thread",
		Long:  "Display all emails in a conversation thread, ordered chronologically.",
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

			emails, err := client.Mail().GetThread(ctx, threadID)
			if err != nil {
				return fmt.Errorf("getting thread: %w", err)
			}

			return outputThread(cmd, emails)
		},
	}

	return cmd
}

// outputThread writes the thread emails to output.
func outputThread(cmd *cobra.Command, emails []fastmail.Email) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(emails)
	}

	w := cmd.OutOrStdout()
	for i, e := range emails {
		_, _ = fmt.Fprintf(w, "--- Email %d of %d ---\n", i+1, len(emails))
		_, _ = fmt.Fprintf(w, "ID:      %s\n", e.ID)
		_, _ = fmt.Fprintf(w, "Subject: %s\n", e.Subject)
		_, _ = fmt.Fprintf(w, "Date:    %s\n", e.ReceivedAt.Format("2006-01-02 15:04:05"))
		if e.Preview != "" {
			_, _ = fmt.Fprintf(w, "Preview: %s\n", e.Preview)
		}
		_, _ = fmt.Fprintln(w)
	}

	return nil
}

// newMailFlagCommand creates the mail flag command.
func newMailFlagCommand() *cobra.Command {
	var (
		markRead   bool
		markUnread bool
		star       bool
		unstar     bool
		flagKeys   []string
		unflagKeys []string
	)

	cmd := &cobra.Command{
		Use:   "flag EMAIL_ID",
		Short: "Set or remove flags on an email",
		Long:  "Set or remove keyword flags (read, starred, custom) on an email.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			emailID := args[0]

			if markRead && markUnread {
				return fmt.Errorf("cannot specify both --read and --unread")
			}
			if star && unstar {
				return fmt.Errorf("cannot specify both --star and --unstar")
			}

			actions := buildKeywordActions(markRead, markUnread, star, unstar, flagKeys, unflagKeys)
			if len(actions) == 0 {
				return fmt.Errorf("at least one flag option is required (--read, --unread, --star, --unstar, --flag, --unflag)")
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if err := client.Mail().SetKeywords(ctx, emailID, actions); err != nil {
				return fmt.Errorf("setting flags: %w", err)
			}

			return outputFlagResult(cmd, emailID, actions)
		},
	}

	cmd.Flags().BoolVar(&markRead, "read", false, "mark as read ($seen)")
	cmd.Flags().BoolVar(&markUnread, "unread", false, "mark as unread (remove $seen)")
	cmd.Flags().BoolVar(&star, "star", false, "star the email ($flagged)")
	cmd.Flags().BoolVar(&unstar, "unstar", false, "unstar the email (remove $flagged)")
	cmd.Flags().StringArrayVar(&flagKeys, "flag", nil, "add a custom keyword")
	cmd.Flags().StringArrayVar(&unflagKeys, "unflag", nil, "remove a custom keyword")

	return cmd
}

// buildKeywordActions converts CLI flags into KeywordAction slice.
func buildKeywordActions(read, unread, star, unstar bool, flagKeys, unflagKeys []string) []fastmail.KeywordAction {
	var actions []fastmail.KeywordAction

	if read {
		actions = append(actions, fastmail.KeywordAction{Keyword: fastmail.KeywordSeen, Set: true})
	}
	if unread {
		actions = append(actions, fastmail.KeywordAction{Keyword: fastmail.KeywordSeen, Set: false})
	}
	if star {
		actions = append(actions, fastmail.KeywordAction{Keyword: fastmail.KeywordFlagged, Set: true})
	}
	if unstar {
		actions = append(actions, fastmail.KeywordAction{Keyword: fastmail.KeywordFlagged, Set: false})
	}
	for _, key := range flagKeys {
		actions = append(actions, fastmail.KeywordAction{Keyword: key, Set: true})
	}
	for _, key := range unflagKeys {
		actions = append(actions, fastmail.KeywordAction{Keyword: key, Set: false})
	}

	return actions
}

// outputFlagResult writes the flag result to output.
func outputFlagResult(cmd *cobra.Command, emailID string, actions []fastmail.KeywordAction) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		keywords := make(map[string]bool, len(actions))
		for _, a := range actions {
			keywords[a.Keyword] = a.Set
		}
		result := map[string]any{"id": emailID, "status": "updated", "keywords": keywords}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Email %s flags updated\n", emailID)
	return nil
}

// newMailAttachmentsCommand creates the mail attachments command.
func newMailAttachmentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachments EMAIL_ID",
		Short: "List attachments on an email",
		Long:  "Display the attachments on an email message.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			emailID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			attachments, err := client.Mail().Attachments(ctx, emailID)
			if err != nil {
				return fmt.Errorf("getting attachments: %w", err)
			}

			return outputAttachments(cmd, attachments)
		},
	}

	return cmd
}

// newMailDownloadCommand creates the mail download command.
func newMailDownloadCommand() *cobra.Command {
	var (
		attachmentName string
		blobID         string
		output         string
	)

	cmd := &cobra.Command{
		Use:   "download EMAIL_ID",
		Short: "Download an attachment",
		Long:  "Download an email attachment by name or blob ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			emailID := args[0]

			if attachmentName == "" && blobID == "" {
				return fmt.Errorf("either --attachment or --blob-id is required")
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			// Find the attachment
			attachments, err := client.Mail().Attachments(ctx, emailID)
			if err != nil {
				return fmt.Errorf("getting attachments: %w", err)
			}

			att, err := findAttachment(attachments, attachmentName, blobID)
			if err != nil {
				return err
			}

			// Download the blob
			reader, err := client.Mail().DownloadAttachment(ctx, att.BlobID, att.Name)
			if err != nil {
				return fmt.Errorf("downloading attachment: %w", err)
			}
			defer func() { _ = reader.Close() }()

			// Write to file or stdout
			return writeAttachment(cmd, reader, att, output)
		},
	}

	cmd.Flags().StringVar(&attachmentName, "attachment", "", "attachment filename to download")
	cmd.Flags().StringVar(&blobID, "blob-id", "", "blob ID to download")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file path (default: stdout)")

	return cmd
}

// newMailUploadCommand creates the mail upload command.
func newMailUploadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload FILE",
		Short: "Upload a file as a blob",
		Long:  "Upload a file to the server for use in email drafts.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			f, err := os.Open(filePath) // #nosec G304 -- user-provided file path is expected
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			defer func() { _ = f.Close() }()

			contentType := detectContentType(filePath)

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			blobID, size, err := client.Mail().UploadBlob(ctx, f, contentType)
			if err != nil {
				return fmt.Errorf("uploading file: %w", err)
			}

			return outputUploadResult(cmd, blobID, size, filePath)
		},
	}

	return cmd
}

// findAttachment finds an attachment by name or blob ID.
func findAttachment(attachments []fastmail.Attachment, name, blobID string) (*fastmail.Attachment, error) {
	for i, att := range attachments {
		if blobID != "" && att.BlobID == blobID {
			return &attachments[i], nil
		}
		if name != "" && strings.EqualFold(att.Name, name) {
			return &attachments[i], nil
		}
	}

	if blobID != "" {
		return nil, fmt.Errorf("attachment with blob ID %q not found", blobID)
	}
	return nil, fmt.Errorf("attachment %q not found", name)
}

// writeAttachment writes the downloaded blob to a file or stdout.
func writeAttachment(cmd *cobra.Command, reader io.Reader, att *fastmail.Attachment, output string) (err error) {
	var w io.Writer
	if output == "" {
		w = cmd.OutOrStdout()
	} else {
		f, ferr := os.Create(output) // #nosec G304 -- user-provided output path is expected
		if ferr != nil {
			return fmt.Errorf("creating output file: %w", ferr)
		}
		defer func() {
			if cerr := f.Close(); cerr != nil && err == nil {
				err = fmt.Errorf("closing output file: %w", cerr)
			}
		}()
		w = f
	}

	written, err := io.Copy(w, reader)
	if err != nil {
		return fmt.Errorf("writing attachment: %w", err)
	}

	if output != "" && !IsQuiet() {
		var size uint64
		if written > 0 {
			size = uint64(written) // #nosec G115 -- written is non-negative from io.Copy
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Downloaded %s (%s) to %s\n", att.Name, formatSize(size), output)
	}
	return nil
}

// detectContentType guesses the MIME type from a file extension.
func detectContentType(path string) string {
	ext := filepath.Ext(path)
	if ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
	}
	return "application/octet-stream"
}

// outputAttachments writes attachment list to output.
func outputAttachments(cmd *cobra.Command, attachments []fastmail.Attachment) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(attachments)
	}

	w := cmd.OutOrStdout()
	if len(attachments) == 0 {
		_, _ = fmt.Fprintln(w, "No attachments")
		return nil
	}

	for _, att := range attachments {
		_, _ = fmt.Fprintf(w, "%-40s  %-30s  %8s  %s\n", att.Name, att.Type, formatSize(att.Size), att.BlobID)
	}
	return nil
}

// outputUploadResult writes the upload result to output.
func outputUploadResult(cmd *cobra.Command, blobID string, size uint64, filename string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]any{
			"blob_id":  blobID,
			"size":     size,
			"filename": filepath.Base(filename),
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Uploaded: blob_id=%s size=%s\n", blobID, formatSize(size))
	return nil
}

// newMailImportCommand creates the mail import command.
func newMailImportCommand() *cobra.Command {
	var (
		folder  string
		seen    bool
		flagged bool
	)

	cmd := &cobra.Command{
		Use:   "import FILE",
		Short: "Import an RFC 5322 email message",
		Long:  `Import an .eml file (RFC 5322 message) into a mailbox. Use "-" to read from stdin.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			var reader io.Reader
			if filePath == "-" {
				reader = cmd.InOrStdin()
			} else {
				f, err := os.Open(filePath) // #nosec G304 -- user-provided file path is expected
				if err != nil {
					return fmt.Errorf("opening file: %w", err)
				}
				defer func() { _ = f.Close() }()
				reader = f
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			keywords := make(map[string]bool)
			if seen {
				keywords["$seen"] = true
			}
			if flagged {
				keywords["$flagged"] = true
			}

			opts := fastmail.ImportOptions{
				Folder:   folder,
				Keywords: keywords,
			}

			result, err := client.Mail().Import(ctx, reader, opts)
			if err != nil {
				return fmt.Errorf("importing email: %w", err)
			}

			return outputImportResult(cmd, result)
		},
	}

	cmd.Flags().StringVarP(&folder, "folder", "f", "Inbox", "target mailbox folder")
	cmd.Flags().BoolVar(&seen, "seen", false, "mark as read ($seen)")
	cmd.Flags().BoolVar(&flagged, "flagged", false, "mark as flagged ($flagged)")

	return cmd
}

// outputImportResult writes the import result to output.
func outputImportResult(cmd *cobra.Command, result *fastmail.ImportResult) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		out := map[string]any{
			"id":        result.ID,
			"blob_id":   result.BlobID,
			"thread_id": result.ThreadID,
			"size":      result.Size,
			"status":    "imported",
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Imported: id=%s blob_id=%s size=%s\n", result.ID, result.BlobID, formatSize(result.Size))
	return nil
}

// outputScheduledResult writes the scheduled send result to output.
func outputScheduledResult(cmd *cobra.Command, emailID, schedule string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"id": emailID, "status": "scheduled", "send_at": schedule}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Email scheduled: %s (delivery at %s)\n", emailID, schedule)
	return nil
}

// newMailScheduledCommand creates the mail scheduled command group.
func newMailScheduledCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scheduled",
		Short: "Manage scheduled email sends",
		Long:  "List and cancel pending scheduled email deliveries.",
	}

	cmd.AddCommand(newMailScheduledListCommand())
	cmd.AddCommand(newMailScheduledCancelCommand())

	return cmd
}

// newMailScheduledListCommand creates the mail scheduled list command.
func newMailScheduledListCommand() *cobra.Command {
	var limit uint64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pending scheduled sends",
		Long:  "List emails that are scheduled for future delivery.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			scheduled, err := client.Mail().ListScheduled(ctx, limit)
			if err != nil {
				return fmt.Errorf("listing scheduled sends: %w", err)
			}

			return outputScheduledList(cmd, scheduled)
		},
	}

	cmd.Flags().Uint64VarP(&limit, "limit", "n", 10, "maximum results to return")

	return cmd
}

// newMailScheduledCancelCommand creates the mail scheduled cancel command.
func newMailScheduledCancelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel SUBMISSION_ID",
		Short: "Cancel a scheduled send",
		Long:  "Cancel a pending scheduled email delivery by its submission ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			submissionID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			if err := client.Mail().CancelScheduled(ctx, submissionID); err != nil {
				return fmt.Errorf("canceling scheduled send: %w", err)
			}

			return outputCancelResult(cmd, submissionID)
		},
	}

	return cmd
}

// outputScheduledList writes the scheduled sends list to output.
func outputScheduledList(cmd *cobra.Command, scheduled []fastmail.ScheduledSend) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		type jsonScheduled struct {
			SubmissionID string   `json:"submission_id"`
			EmailID      string   `json:"email_id"`
			Subject      string   `json:"subject"`
			To           []string `json:"to"`
			SendAt       string   `json:"send_at"`
			Status       string   `json:"status"`
		}
		out := make([]jsonScheduled, len(scheduled))
		for i, s := range scheduled {
			toAddrs := make([]string, len(s.To))
			for j, addr := range s.To {
				toAddrs[j] = addr.String()
			}
			out[i] = jsonScheduled{
				SubmissionID: s.SubmissionID,
				EmailID:      s.EmailID,
				Subject:      s.Subject,
				To:           toAddrs,
				SendAt:       s.SendAt.Format(time.RFC3339),
				Status:       s.UndoStatus,
			}
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	w := cmd.OutOrStdout()
	if len(scheduled) == 0 {
		_, _ = fmt.Fprintln(w, "No scheduled sends")
		return nil
	}

	for _, s := range scheduled {
		toAddrs := make([]string, len(s.To))
		for i, addr := range s.To {
			toAddrs[i] = addr.String()
		}
		_, _ = fmt.Fprintf(w, "%s  %s  to:%s  at:%s  [%s]\n",
			s.SubmissionID, s.Subject, strings.Join(toAddrs, ","), s.SendAt.Format(time.RFC3339), s.UndoStatus)
	}
	return nil
}

// outputCancelResult writes the cancel result to output.
func outputCancelResult(cmd *cobra.Command, submissionID string) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		result := map[string]string{"submission_id": submissionID, "status": "canceled"}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Scheduled send %s canceled\n", submissionID)
	return nil
}

// formatSize formats a byte count into a human-readable string.
func formatSize(bytes uint64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

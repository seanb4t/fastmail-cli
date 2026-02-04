package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
	"github.com/spf13/cobra"
)

// Export format constants.
const (
	formatJSONL   = "jsonl"
	formatMaildir = "maildir"
	formatMbox    = "mbox"
)

// newExportCommand creates the export command.
func newExportCommand() *cobra.Command {
	var folder string
	var format string
	var output string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export emails from a folder",
		Long: `Export emails from a mailbox folder to various formats.

Formats:
  jsonl   - JSON Lines format, one email per line (default)
  maildir - Maildir directory structure
  mbox    - Standard Unix mbox format

Examples:
  fastmail export --folder Inbox --format jsonl > emails.jsonl
  fastmail export --folder Inbox --format maildir --output ~/backup/inbox
  fastmail export --folder Archive --format mbox --output archive.mbox`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runExport(cmd, folder, format, output)
		},
	}

	cmd.Flags().StringVarP(&folder, "folder", "f", "Inbox", "mailbox folder name")
	cmd.Flags().StringVar(&format, "format", formatJSONL, "export format: jsonl, maildir, mbox")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file/directory (default: stdout for jsonl/mbox)")

	return cmd
}

func runExport(cmd *cobra.Command, folder, format, output string) error {
	// Validate format
	switch format {
	case formatJSONL, formatMaildir, formatMbox:
		// valid
	default:
		return fmt.Errorf("unsupported format %q: must be jsonl, maildir, or mbox", format)
	}

	// Maildir requires output directory
	if format == formatMaildir && output == "" {
		return fmt.Errorf("--output is required for maildir format")
	}

	client, err := createClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connecting: %w", err)
	}

	// Fetch all emails (no limit for export)
	if !IsQuiet() {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Fetching emails from %s...\n", folder)
	}

	emails, err := client.Mail().List(ctx, folder, 0)
	if err != nil {
		return fmt.Errorf("listing emails: %w", err)
	}

	if !IsQuiet() {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Exporting %d emails...\n", len(emails))
	}

	switch format {
	case formatJSONL:
		return exportJSONL(cmd, emails, output)
	case formatMaildir:
		return exportMaildir(cmd, emails, output)
	case formatMbox:
		return exportMbox(cmd, emails, output)
	}

	return nil
}

func exportJSONL(cmd *cobra.Command, emails []fastmail.Email, output string) (err error) {
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

	if err := fastmail.ExportJSONL(w, emails); err != nil {
		return fmt.Errorf("exporting jsonl: %w", err)
	}

	if !IsQuiet() && output != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Exported to %s\n", output)
	}
	return nil
}

func exportMaildir(cmd *cobra.Command, emails []fastmail.Email, output string) error {
	if err := fastmail.ExportMaildir(output, emails); err != nil {
		return fmt.Errorf("exporting maildir: %w", err)
	}

	if !IsQuiet() {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Exported to %s\n", output)
	}
	return nil
}

func exportMbox(cmd *cobra.Command, emails []fastmail.Email, output string) (err error) {
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

	if err := fastmail.ExportMbox(w, emails); err != nil {
		return fmt.Errorf("exporting mbox: %w", err)
	}

	if !IsQuiet() && output != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Exported to %s\n", output)
	}
	return nil
}

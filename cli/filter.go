package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/pkg/fastmail"
)

// newFilterCommand creates the filter command with subcommands.
func newFilterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Sieve filter script operations",
		Long:  "Commands for managing server-side Sieve email filter scripts.",
	}

	cmd.AddCommand(newFilterListCommand())
	cmd.AddCommand(newFilterShowCommand())
	cmd.AddCommand(newFilterCreateCommand())
	cmd.AddCommand(newFilterActivateCommand())
	cmd.AddCommand(newFilterDeactivateCommand())
	cmd.AddCommand(newFilterValidateCommand())
	cmd.AddCommand(newFilterDeleteCommand())

	return cmd
}

// newFilterListCommand creates the filter list command.
func newFilterListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List filter scripts",
		Long:  "List all Sieve filter scripts.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			scripts, err := client.Sieve().List(ctx)
			if err != nil {
				return fmt.Errorf("listing filter scripts: %w", err)
			}

			return outputFilterScripts(cmd, scripts)
		},
	}
}

// newFilterShowCommand creates the filter show command.
func newFilterShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show ID",
		Short: "Show filter script details",
		Long:  "Show a Sieve filter script with its content.",
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

			script, err := client.Sieve().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("getting filter script: %w", err)
			}

			return outputFilterScript(cmd, script)
		},
	}
}

// newFilterCreateCommand creates the filter create command.
func newFilterCreateCommand() *cobra.Command {
	var (
		name     string
		file     string
		activate bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a filter script",
		Long:  "Create a new Sieve filter script from a file or stdin.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			scriptContent, err := readScriptInput(file)
			if err != nil {
				return err
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			opts := fastmail.CreateSieveScriptOptions{
				Name:     name,
				Script:   scriptContent,
				Activate: activate,
			}

			script, err := client.Sieve().Create(ctx, opts)
			if err != nil {
				return fmt.Errorf("creating filter script: %w", err)
			}

			return outputFilterCreated(cmd, script)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "filter script name (required)")
	cmd.Flags().StringVar(&file, "file", "", "path to sieve script file (reads stdin if omitted)")
	cmd.Flags().BoolVar(&activate, "activate", false, "activate the script on creation")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// newFilterActivateCommand creates the filter activate command.
func newFilterActivateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "activate ID",
		Short: "Activate a filter script",
		Long:  "Activate a Sieve filter script. Only one script can be active at a time.",
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

			if err := client.Sieve().Activate(ctx, id); err != nil {
				return fmt.Errorf("activating filter script: %w", err)
			}

			return outputFilterStatus(cmd, id, "Activated")
		},
	}
}

// newFilterDeactivateCommand creates the filter deactivate command.
func newFilterDeactivateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "deactivate ID",
		Short: "Deactivate a filter script",
		Long:  "Deactivate an active Sieve filter script.",
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

			if err := client.Sieve().Deactivate(ctx, id); err != nil {
				return fmt.Errorf("deactivating filter script: %w", err)
			}

			return outputFilterStatus(cmd, id, "Deactivated")
		},
	}
}

// newFilterValidateCommand creates the filter validate command.
func newFilterValidateCommand() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a filter script",
		Long:  "Validate a Sieve filter script syntax without storing it.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			scriptContent, err := readScriptInput(file)
			if err != nil {
				return err
			}

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()
			if err := client.Connect(ctx); err != nil {
				return fmt.Errorf("connecting: %w", err)
			}

			result, err := client.Sieve().Validate(ctx, scriptContent)
			if err != nil {
				return fmt.Errorf("validating filter script: %w", err)
			}

			return outputFilterValidation(cmd, result)
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "path to sieve script file (reads stdin if omitted)")

	return cmd
}

// newFilterDeleteCommand creates the filter delete command.
func newFilterDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ID",
		Short: "Delete a filter script",
		Long:  "Permanently delete a Sieve filter script.",
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

			if err := client.Sieve().Delete(ctx, id); err != nil {
				return fmt.Errorf("deleting filter script: %w", err)
			}

			return outputFilterStatus(cmd, id, "Deleted")
		},
	}
}

// readScriptInput reads sieve script content from a file or stdin.
func readScriptInput(filePath string) (string, error) {
	if filePath != "" {
		data, err := os.ReadFile(filePath) //nolint:gosec // user-specified file path from CLI flag
		if err != nil {
			return "", fmt.Errorf("reading file %s: %w", filePath, err)
		}
		return string(data), nil
	}

	// Check if stdin has data (is a pipe)
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("checking stdin: %w", err)
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		return "", fmt.Errorf("no input: use --file or pipe script content via stdin")
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}
	return string(data), nil
}

// outputFilterScripts writes the filter script list to output.
func outputFilterScripts(cmd *cobra.Command, scripts []fastmail.SieveScript) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(scripts)
	}

	for _, s := range scripts {
		active := " "
		if s.IsActive {
			active = "*"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  [%s]  %s\n",
			s.ID, s.Name, active, activeLabel(s.IsActive))
	}

	return nil
}

// outputFilterScript writes a single filter script with content to output.
func outputFilterScript(cmd *cobra.Command, script *fastmail.SieveScript) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(script)
	}

	w := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(w, "ID:       %s\n", script.ID)
	_, _ = fmt.Fprintf(w, "Name:     %s\n", script.Name)
	_, _ = fmt.Fprintf(w, "Active:   %s\n", activeLabel(script.IsActive))
	if script.BlobID != "" {
		_, _ = fmt.Fprintf(w, "Blob ID:  %s\n", script.BlobID)
	}
	if script.Script != "" {
		_, _ = fmt.Fprintf(w, "\n--- Script ---\n%s\n", script.Script)
	}

	return nil
}

// outputFilterCreated writes the created filter script to output.
func outputFilterCreated(cmd *cobra.Command, script *fastmail.SieveScript) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(script)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created: %s (%s)\n", script.Name, script.ID)
	return nil
}

// outputFilterStatus writes a status update to output.
func outputFilterStatus(cmd *cobra.Command, id, status string) error {
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

// outputFilterValidation writes validation results to output.
func outputFilterValidation(cmd *cobra.Command, result *fastmail.SieveValidationResult) error {
	if IsQuiet() {
		return nil
	}

	if IsJSONOutput() {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	w := cmd.OutOrStdout()
	if result.IsValid {
		_, _ = fmt.Fprintln(w, "Script is valid")
	} else {
		_, _ = fmt.Fprintf(w, "Script is invalid: %s\n", result.Description)
	}

	return nil
}

// activeLabel returns a human-readable label for the active state.
func activeLabel(active bool) string {
	if active {
		return "active"
	}
	return "inactive"
}

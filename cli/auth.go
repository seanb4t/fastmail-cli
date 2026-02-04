package cli

import (
	"encoding/json"
	"fmt"
	"syscall"

	"github.com/seanb4t/fastmail-cli/internal/auth"
	"github.com/seanb4t/fastmail-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// newAuthCommand creates the auth subcommand with login/logout/status.
func newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Commands for managing FastMail authentication credentials.",
	}

	cmd.AddCommand(newAuthLoginCommand())
	cmd.AddCommand(newAuthLogoutCommand())
	cmd.AddCommand(newAuthStatusCommand())

	return cmd
}

// newAuthLoginCommand creates the auth login command.
func newAuthLoginCommand() *cobra.Command {
	var tokenFlag string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store API token",
		Long:  "Store your FastMail API token for authentication.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			configPath := GetConfigPath()
			if configPath == "" {
				configPath = config.DefaultConfigPath()
			}

			store := auth.NewStore(configPath)
			store.DisableKeychain() // Use file storage for now

			var token string
			if tokenFlag != "" {
				token = tokenFlag
			} else {
				// Interactive mode - prompt for token
				if !term.IsTerminal(syscall.Stdin) {
					return fmt.Errorf("no token provided: use --token flag or run interactively")
				}

				_, _ = fmt.Fprint(cmd.OutOrStdout(), "Enter API token: ")
				tokenBytes, err := term.ReadPassword(syscall.Stdin)
				if err != nil {
					return fmt.Errorf("reading token: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout()) // newline after hidden input
				token = string(tokenBytes)
			}

			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}

			if err := store.SetToken(token); err != nil {
				return fmt.Errorf("storing token: %w", err)
			}

			if !IsQuiet() {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Logged in successfully")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&tokenFlag, "token", "", "API token (for non-interactive use)")

	return cmd
}

// newAuthLogoutCommand creates the auth logout command.
func newAuthLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		Long:  "Remove stored FastMail API token.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			configPath := GetConfigPath()
			if configPath == "" {
				configPath = config.DefaultConfigPath()
			}

			store := auth.NewStore(configPath)
			store.DisableKeychain() // Use file storage for now

			if err := store.DeleteToken(); err != nil {
				return fmt.Errorf("removing token: %w", err)
			}

			if !IsQuiet() {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Logged out successfully")
			}
			return nil
		},
	}
}

// newAuthStatusCommand creates the auth status command.
func newAuthStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  "Show whether you are currently logged in.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			configPath := GetConfigPath()
			if configPath == "" {
				configPath = config.DefaultConfigPath()
			}

			store := auth.NewStore(configPath)
			store.DisableKeychain() // Use file storage for now

			loggedIn := store.HasToken()

			if IsJSONOutput() {
				result := map[string]any{
					"logged_in": loggedIn,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			if loggedIn {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Logged in")
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Not logged in")
			}
			return nil
		},
	}
}

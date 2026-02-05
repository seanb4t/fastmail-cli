package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/seanb4t/fastmail-cli/internal/auth"
	"github.com/seanb4t/fastmail-cli/internal/config"
	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// authStatusHTTPClient allows injecting a custom HTTP client for testing.
var authStatusHTTPClient *http.Client

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
		Long:  "Validate your FastMail API token and show authentication status.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			configPath := GetConfigPath()
			if configPath == "" {
				configPath = config.DefaultConfigPath()
			}

			store := auth.NewStore(configPath)
			store.DisableKeychain() // Use file storage for now

			// Check if token exists
			token, err := store.GetToken()
			if err != nil || token == "" {
				if IsJSONOutput() {
					_ = outputAuthJSON(cmd, map[string]any{
						"authenticated": false,
						"reason":        "no_token",
					})
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Not logged in")
				}
				return &AuthStatusError{Code: ExitNoToken, Message: "no token stored"}
			}

			// Validate token against API
			var opts []jmap.ClientOption
			if authStatusHTTPClient != nil {
				opts = append(opts, jmap.WithHTTPClient(authStatusHTTPClient))
			}
			client := jmap.NewClient(jmap.DefaultSessionURL, token, opts...)
			session, err := client.Authenticate(ctx)
			if err != nil {
				return handleAuthStatusError(cmd, err)
			}

			// Success
			username := session.Username
			if IsJSONOutput() {
				return outputAuthJSON(cmd, map[string]any{
					"authenticated": true,
					"username":      username,
				})
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Authenticated as %s\n", username)
			return nil
		},
	}
}

func handleAuthStatusError(cmd *cobra.Command, err error) error {
	// Check for auth errors (401, 403)
	if isAuthError(err) {
		if IsJSONOutput() {
			_ = outputAuthJSON(cmd, map[string]any{
				"authenticated": false,
				"reason":        "invalid_token",
			})
		} else {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Token expired or revoked")
		}
		return &AuthStatusError{Code: ExitInvalidToken, Message: "token invalid"}
	}

	// Network error
	if IsJSONOutput() {
		_ = outputAuthJSON(cmd, map[string]any{
			"authenticated": false,
			"reason":        "network_error",
			"error":         err.Error(),
		})
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cannot reach FastMail API: %v\n", err)
	}
	return &AuthStatusError{Code: ExitNetworkError, Message: err.Error()}
}

func isAuthError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "401") || strings.Contains(errStr, "403")
}

func outputAuthJSON(cmd *cobra.Command, data map[string]any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

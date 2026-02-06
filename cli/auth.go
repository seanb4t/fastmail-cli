package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/seanb4t/fastmail-cli/internal/auth"
	"github.com/seanb4t/fastmail-cli/internal/config"
	"github.com/seanb4t/fastmail-cli/internal/jmap"
)

// authStatusHTTPClient allows injecting a custom HTTP client for testing.
var authStatusHTTPClient *http.Client

// authLoginHTTPClient allows injecting a custom HTTP client for testing.
var authLoginHTTPClient *http.Client

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
			ctx := commandContext(cmd)
			configPath := defaultConfigPath()
			store := auth.NewStore(configPath)
			setAuthStoreWarningWriter(cmd, store)

			token, err := resolveLoginToken(cmd, tokenFlag)
			if err != nil {
				return err
			}

			endpoint := loadAuthEndpoint(cmd, configPath)
			if err := validateToken(ctx, endpoint, token); err != nil {
				return err
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
			setAuthStoreWarningWriter(cmd, store)

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
			cfg, err := config.Load(configPath)
			if err != nil {
				if !IsQuiet() {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to load config (%v); using default endpoint\n", err)
				}
				cfg = &config.Config{Endpoint: jmap.DefaultSessionURL}
			}

			store := auth.NewStore(configPath)
			setAuthStoreWarningWriter(cmd, store)

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
			endpoint := cfg.Endpoint
			if !IsQuiet() {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Resolved JMAP endpoint: %s\n", endpoint)
			}
			client := jmap.NewClient(endpoint, token, opts...)
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

func setAuthStoreWarningWriter(cmd *cobra.Command, store *auth.Store) {
	setStoreWarningWriter(store, cmd.ErrOrStderr())
}

func commandContext(cmd *cobra.Command) context.Context {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx
}

func defaultConfigPath() string {
	configPath := GetConfigPath()
	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}
	return configPath
}

func resolveLoginToken(cmd *cobra.Command, tokenFlag string) (string, error) {
	if tokenFlag != "" {
		return tokenFlag, nil
	}

	if !term.IsTerminal(syscall.Stdin) {
		return "", fmt.Errorf("no token provided: use --token flag or run interactively")
	}

	_, _ = fmt.Fprint(cmd.OutOrStdout(), "Enter API token: ")
	tokenBytes, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading token: %w", err)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	if token := string(tokenBytes); token != "" {
		return token, nil
	}

	return "", fmt.Errorf("token cannot be empty")
}

func loadAuthEndpoint(cmd *cobra.Command, configPath string) string {
	cfg, err := config.Load(configPath)
	if err != nil {
		if !IsQuiet() {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to load config (%v); using default endpoint\n", err)
		}
		cfg = &config.Config{Endpoint: jmap.DefaultSessionURL}
	}

	if cfg.Endpoint == "" {
		return jmap.DefaultSessionURL
	}

	return cfg.Endpoint
}

func validateToken(ctx context.Context, endpoint, token string) error {
	var opts []jmap.ClientOption
	if authLoginHTTPClient != nil {
		opts = append(opts, jmap.WithHTTPClient(authLoginHTTPClient))
	}
	client := jmap.NewClient(endpoint, token, opts...)
	if _, err := client.Authenticate(ctx); err != nil {
		if isAuthError(err) {
			return fmt.Errorf("invalid token")
		}
		return fmt.Errorf("validating token: %w", err)
	}
	return nil
}

func setStoreWarningWriter(store *auth.Store, errWriter io.Writer) {
	if IsQuiet() {
		store.SetWarningWriter(io.Discard)
		return
	}
	store.SetWarningWriter(errWriter)
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

	if isNetworkError(err) {
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

	if IsJSONOutput() {
		_ = outputAuthJSON(cmd, map[string]any{
			"authenticated": false,
			"reason":        "auth_error",
		})
	} else {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Authentication failed")
	}
	return &AuthStatusError{Code: ExitAuthError, Message: err.Error()}
}

func isAuthError(err error) bool {
	if httpErr, ok := asHTTPError(err); ok {
		return httpErr.StatusCode == http.StatusUnauthorized || httpErr.StatusCode == http.StatusForbidden
	}
	return false
}

func asHTTPError(err error) (*jmap.HTTPError, bool) {
	var httpErr *jmap.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr, true
	}
	return nil, false
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	return false
}

func outputAuthJSON(cmd *cobra.Command, data map[string]any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

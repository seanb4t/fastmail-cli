// Package cli implements the command-line interface.
package cli

import (
	"github.com/spf13/cobra"
)

// Version information - set at build time via ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Global flags accessible to subcommands.
var (
	cfgFile    string
	jsonOutput bool
	quiet      bool
)

// NewRootCommand creates and returns the root cobra command.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fastmail-cli",
		Short: "FastMail CLI - manage your FastMail account from the command line",
		Long: `fastmail-cli is a command-line interface for interacting with FastMail.

It provides commands for authentication, managing contacts, calendars,
and other FastMail features using the JMAP protocol.`,
		Version: version,
		RunE: func(_ *cobra.Command, _ []string) error {
			// Root command with no args shows help - handled by cobra
			return nil
		},
		SilenceUsage: true,
	}

	// Set custom version template to include commit and date.
	cmd.SetVersionTemplate("{{.Name}} version {{.Version}} (commit: " + commit + ", built: " + date + ")\n")

	// Global flags
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fastmail-cli.yaml)")
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	cmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "suppress non-essential output")

	// Add subcommands
	cmd.AddCommand(newAuthCommand())

	return cmd
}

// newAuthCommand creates a placeholder auth command.
func newAuthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  `Commands for managing FastMail authentication and credentials.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
}

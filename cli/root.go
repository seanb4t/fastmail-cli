package cli

import (
	"github.com/spf13/cobra"
)

// Version information set at build time via ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// NewRootCommand creates the root cobra command with global flags.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fastmail-cli",
		Short: "CLI for interacting with Fastmail",
		Long:  `A command-line interface for managing Fastmail email, calendars, and contacts.`,
		// Silence usage on error - we handle it ourselves
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Root command with no subcommand shows help
			return cmd.Help()
		},
	}

	// Version template
	cmd.Version = Version
	cmd.SetVersionTemplate("fastmail-cli version {{.Version}}\n")

	// Global flags
	cmd.PersistentFlags().String("config", "", "custom config file path")
	cmd.PersistentFlags().Bool("json", false, "force JSON output")
	cmd.PersistentFlags().Bool("quiet", false, "suppress output")

	// Add subcommands
	cmd.AddCommand(newAuthCommand())

	return cmd
}

// newAuthCommand creates a placeholder auth subcommand.
func newAuthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  `Commands for managing Fastmail authentication credentials.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
}

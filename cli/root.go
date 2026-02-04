package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information set by build flags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Global flag values accessible to subcommands.
var (
	cfgFile   string
	jsonFlag  bool
	quietFlag bool
)

// NewRootCommand creates the root cobra command for fastmail-cli.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fastmail-cli",
		Short: "Command-line interface for FastMail",
		Long: `fastmail-cli is a command-line tool for interacting with FastMail
using the JMAP protocol. It supports mailbox management, message operations,
and other FastMail features.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	// Global persistent flags
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.config/fastmail-cli/config.yaml)")
	cmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "force JSON output")
	cmd.PersistentFlags().BoolVar(&quietFlag, "quiet", false, "suppress output")

	// Custom version template
	cmd.SetVersionTemplate(`{{.Name}} version {{.Version}}
`)

	// Add subcommands
	cmd.AddCommand(newAuthCommand())

	return cmd
}

// newAuthCommand is defined in auth.go

// GetConfigPath returns the config file path from the flag.
func GetConfigPath() string {
	return cfgFile
}

// IsJSONOutput returns true if JSON output is requested.
func IsJSONOutput() bool {
	return jsonFlag
}

// IsQuiet returns true if quiet mode is enabled.
func IsQuiet() bool {
	return quietFlag
}

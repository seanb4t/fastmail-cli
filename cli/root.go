// Package cli implements the command-line interface using cobra.
package cli

import (
	"io"

	"github.com/spf13/cobra"
)

// Version information set at build time via ldflags.
var version = "dev"

// rootCmd wraps the cobra root command for testing.
type rootCmd struct {
	cmd *cobra.Command
}

// newRootCmd creates the root command with all global flags.
func newRootCmd() *rootCmd {
	root := &rootCmd{}

	root.cmd = &cobra.Command{
		Use:   "fastmail-cli",
		Short: "CLI for Fastmail JMAP API",
		Long: `fastmail-cli is a command-line tool for interacting with
the Fastmail JMAP API. It supports authentication, email management,
and other Fastmail features.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	// Custom version template
	root.cmd.SetVersionTemplate("fastmail-cli version {{.Version}}\n")

	// Global flags
	root.cmd.PersistentFlags().StringP("config", "c", "", "config file path")
	root.cmd.PersistentFlags().Bool("json", false, "force JSON output")
	root.cmd.PersistentFlags().BoolP("quiet", "q", false, "suppress output")

	// Add subcommands
	root.cmd.AddCommand(newAuthCmd())

	return root
}

// Execute runs the root command.
func (r *rootCmd) Execute() error {
	return r.cmd.Execute()
}

// SetOut sets the output writer.
func (r *rootCmd) SetOut(w io.Writer) {
	r.cmd.SetOut(w)
}

// SetErr sets the error writer.
func (r *rootCmd) SetErr(w io.Writer) {
	r.cmd.SetErr(w)
}

// SetArgs sets command arguments for testing.
func (r *rootCmd) SetArgs(args []string) {
	r.cmd.SetArgs(args)
}

// newAuthCmd creates a placeholder auth subcommand.
func newAuthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Commands for managing Fastmail authentication and credentials.",
	}
}

// Execute runs the CLI.
func Execute() error {
	return newRootCmd().Execute()
}

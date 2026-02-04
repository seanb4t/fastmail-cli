package cli

import (
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

// RootCommand holds the root cobra command and global flag values.
type RootCommand struct {
	cmd        *cobra.Command
	ConfigFile string
	JSONOutput bool
	Quiet      bool
}

// global instance for flag access from subcommands.
var rootCmd *RootCommand

// NewRootCommand creates and configures the root command.
func NewRootCommand() *RootCommand {
	root := &RootCommand{}
	rootCmd = root

	root.cmd = &cobra.Command{
		Use:   "fastmail-cli",
		Short: "CLI for FastMail JMAP API",
		Long:  "A command-line interface for interacting with FastMail via JMAP.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		SilenceUsage: true,
	}

	root.cmd.Version = Version
	root.cmd.SetVersionTemplate("fastmail-cli version {{.Version}}\n")

	// Global flags
	root.cmd.PersistentFlags().StringVar(&root.ConfigFile, "config", "", "config file path")
	root.cmd.PersistentFlags().BoolVar(&root.JSONOutput, "json", false, "output as JSON")
	root.cmd.PersistentFlags().BoolVar(&root.Quiet, "quiet", false, "suppress output")

	// Add subcommands
	root.cmd.AddCommand(newAuthCommand())

	return root
}

// GetConfigPath returns the current config file path.
func GetConfigPath() string {
	if rootCmd != nil {
		return rootCmd.ConfigFile
	}
	return ""
}

// IsJSONOutput returns whether JSON output is enabled.
func IsJSONOutput() bool {
	if rootCmd != nil {
		return rootCmd.JSONOutput
	}
	return false
}

// IsQuiet returns whether quiet mode is enabled.
func IsQuiet() bool {
	if rootCmd != nil {
		return rootCmd.Quiet
	}
	return false
}

// Execute runs the root command.
func (r *RootCommand) Execute() error {
	return r.cmd.Execute()
}

// SetOut sets the output writer for the command.
func (r *RootCommand) SetOut(w interface{ Write([]byte) (int, error) }) {
	r.cmd.SetOut(w)
}

// SetErr sets the error writer for the command.
func (r *RootCommand) SetErr(w interface{ Write([]byte) (int, error) }) {
	r.cmd.SetErr(w)
}

// SetArgs sets the arguments for the command.
func (r *RootCommand) SetArgs(args []string) {
	r.cmd.SetArgs(args)
}

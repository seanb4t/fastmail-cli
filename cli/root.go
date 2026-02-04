package cli

import (
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags
var Version = "dev"

// RootCommand holds the root cobra command and global flag values.
type RootCommand struct {
	cmd        *cobra.Command
	ConfigFile string
	JSONOutput bool
	Quiet      bool
}

// NewRootCommand creates and configures the root command.
func NewRootCommand() *RootCommand {
	root := &RootCommand{}

	root.cmd = &cobra.Command{
		Use:   "fastmail-cli",
		Short: "CLI for FastMail JMAP API",
		Long:  "A command-line interface for interacting with FastMail via JMAP.",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	// Add auth subcommand placeholder
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}
	root.cmd.AddCommand(authCmd)

	return root
}

// Execute runs the root command.
func (r *RootCommand) Execute() error {
	return r.cmd.Execute()
}

// SetOut sets the output writer for the command.
func (r *RootCommand) SetOut(w interface{ Write([]byte) (int, error) }) {
	r.cmd.SetOut(w)
}

// SetArgs sets the arguments for the command.
func (r *RootCommand) SetArgs(args []string) {
	r.cmd.SetArgs(args)
}

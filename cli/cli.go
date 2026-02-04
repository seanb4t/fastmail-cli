// Package cli implements the command-line interface.
package cli

// Execute runs the root command.
// Returns an error if command execution fails.
func Execute() error {
	return NewRootCommand().Execute()
}

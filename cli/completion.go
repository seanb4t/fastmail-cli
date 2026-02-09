package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newCompletionCommand creates the completion command for generating shell completions.
func newCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate shell completion script",
		Long:      "Generate a shell completion script for the specified shell. Load the output into your shell to enable tab completions.",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompletion(cmd, args[0])
		},
	}

	return cmd
}

func runCompletion(cmd *cobra.Command, shell string) error {
	root := cmd.Root()
	out := cmd.OutOrStdout()

	switch shell {
	case "bash":
		return root.GenBashCompletionV2(out, true)
	case "zsh":
		return root.GenZshCompletion(out)
	case "fish":
		return root.GenFishCompletion(out, true)
	case "powershell":
		return root.GenPowerShellCompletionWithDesc(out)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

package cli

import (
	"github.com/spf13/cobra"

	"github.com/seanb4t/fastmail-cli/internal/tui"
)

func newTUICommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Interactive terminal UI",
		Long:  "Launch an interactive terminal interface for browsing mailboxes and reading email.",
		RunE: func(_ *cobra.Command, _ []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}

			return tui.Run(client)
		},
	}
}

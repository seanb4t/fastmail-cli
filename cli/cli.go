// Package cli implements the command-line interface.
package cli

import (
	"errors"
	"os"
)

// Execute runs the root command.
// Handles exit codes for AuthStatusError, calling os.Exit directly.
func Execute() error {
	err := NewRootCommand().Execute()
	if err != nil {
		var authErr *AuthStatusError
		if errors.As(err, &authErr) {
			os.Exit(authErr.Code)
		}
		return err
	}
	return nil
}

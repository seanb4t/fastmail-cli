// Package main is the entrypoint for the fastmail-cli binary.
package main

import (
	"os"

	"github.com/seanb4t/fastmail-cli/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

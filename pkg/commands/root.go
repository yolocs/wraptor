// Package commands contains all the commands for the application.
// Even though this makes it look like a CLI, it doesn't have to be.
package commands

import (
	"context"

	"github.com/abcxyz/pkg/cli"
)

var rootCmd = func() cli.Command {
	return &cli.RootCommand{
		Name:    "wraptor",
		Version: "dev",
		Commands: map[string]cli.CommandFactory{
			"wrap": func() cli.Command { return &WrapCommand{} },
		},
	}
}

// Run executes the CLI.
func Run(ctx context.Context, args []string) error {
	return rootCmd().Run(ctx, args) //nolint:wrapcheck // Want passthrough
}

package main

import (
	"fmt"
	"os"

	"github.com/cstone-io/twine/cmd/twine/commands"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "twine",
		Short: "Twine - A full-stack Go web framework",
		Long:  "Twine is a full-stack Go web framework for building server-side rendered applications with HTMX.",
	}

	// Add subcommands
	rootCmd.AddCommand(commands.NewDevCommand())
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewRoutesCommand())
	rootCmd.AddCommand(commands.NewUpdateCommand())
	rootCmd.AddCommand(commands.NewVersionCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

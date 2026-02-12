package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	// Build-time variables (set via -ldflags)
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version, commit, build date, and other build information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Twine CLI\n")
			fmt.Printf("  Version:    %s\n", version)
			fmt.Printf("  Commit:     %s\n", commit)
			fmt.Printf("  Built:      %s\n", date)
			fmt.Printf("  Built by:   %s\n", builtBy)
		},
	}
}

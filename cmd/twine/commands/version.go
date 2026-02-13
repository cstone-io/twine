package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	// Build-time variables (set via -ldflags)
	// Exported for use by update command
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "unknown"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version, commit, build date, and other build information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Twine CLI\n")
			fmt.Printf("  Version:    %s\n", Version)
			fmt.Printf("  Commit:     %s\n", Commit)
			fmt.Printf("  Built:      %s\n", Date)
			fmt.Printf("  Built by:   %s\n", BuiltBy)
		},
	}
}

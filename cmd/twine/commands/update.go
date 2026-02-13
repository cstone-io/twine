package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/cstone-io/twine/internal/updater"
	"github.com/spf13/cobra"
)

var (
	updateVersion string
	checkOnly     bool
	listReleases  bool
	skipConfirm   bool
)

func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update twine to the latest version",
		Long: `Update twine CLI to the latest version or a specific version.

Examples:
  twine update                      # Update to latest with confirmation
  twine update --check              # Check if update available
  twine update --list               # List all available releases
  twine update --version v0.2.0     # Update to specific version
  twine update --yes                # Update without confirmation`,
		RunE: runUpdate,
	}

	cmd.Flags().StringVar(&updateVersion, "version", "", "Update to specific version (e.g., v0.2.0)")
	cmd.Flags().BoolVar(&checkOnly, "check", false, "Check if update is available without installing")
	cmd.Flags().BoolVar(&listReleases, "list", false, "List all available releases")
	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	u := updater.NewUpdater()

	// Handle --list flag
	if listReleases {
		return handleListReleases(u)
	}

	// Handle --check flag
	if checkOnly {
		return handleCheckForUpdate(u)
	}

	// Perform the update
	return handleUpdate(u)
}

func handleListReleases(u *updater.Updater) error {
	releases, err := u.GetGitHubClient().ListReleases()
	if err != nil {
		return fmt.Errorf("failed to fetch releases: %w", err)
	}

	if len(releases) == 0 {
		fmt.Println("No releases found")
		return nil
	}

	fmt.Println("Available releases:")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "VERSION\tRELEASE DATE\tPRERELEASE")
	fmt.Fprintln(w, "-------\t------------\t----------")

	for _, release := range releases {
		prerelease := ""
		if release.Prerelease {
			prerelease = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			release.TagName,
			release.PublishedAt.Format("2006-01-02"),
			prerelease,
		)
	}

	w.Flush()
	return nil
}

func handleCheckForUpdate(u *updater.Updater) error {
	fmt.Printf("Current version: %s\n", Version)
	fmt.Println("Checking for updates...")

	result, err := u.CheckForUpdate(Version)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	fmt.Println(result.Message)

	if updater.IsNewer(Version, result.ToVersion) {
		fmt.Printf("\nRun 'twine update' to upgrade to %s\n", result.ToVersion)
	}

	return nil
}

func handleUpdate(u *updater.Updater) error {
	fmt.Printf("Current version: %s\n", Version)

	// If no specific version requested, check for latest
	if updateVersion == "" {
		fmt.Println("Checking for updates...")
		result, err := u.CheckForUpdate(Version)
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		if !updater.IsNewer(Version, result.ToVersion) {
			fmt.Println(result.Message)
			return nil
		}

		fmt.Printf("Update available: %s → %s\n", Version, result.ToVersion)
	} else {
		fmt.Printf("Target version: %s\n", updateVersion)
	}

	// Prompt for confirmation unless --yes is set or current version is dev
	if !skipConfirm {
		if Version == "dev" {
			fmt.Println("\nCurrent version is 'dev' (development build).")
			fmt.Print("Update anyway? [y/N]: ")
		} else {
			fmt.Print("\nProceed with update? [y/N]: ")
		}

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Update cancelled")
			return nil
		}
	}

	// Perform the update
	fmt.Println("Downloading update...")

	opts := updater.UpdateOptions{
		CurrentVersion: Version,
		TargetVersion:  updateVersion,
	}

	result, err := u.Update(opts)
	if err != nil {
		// Provide helpful error messages
		errMsg := err.Error()
		if strings.Contains(errMsg, "no binary available") {
			return fmt.Errorf("%w\n\nVisit https://github.com/cstone-io/twine/releases to download manually", err)
		}
		if strings.Contains(errMsg, "not found") && updateVersion != "" {
			return fmt.Errorf("%w\n\nRun 'twine update --list' to see available versions", err)
		}
		if strings.Contains(errMsg, "permission denied") {
			return err // Already has helpful message from updater
		}
		return fmt.Errorf("update failed: %w", err)
	}

	if result.Updated {
		fmt.Println("✓", result.Message)
		fmt.Println("\nRestart your terminal or run 'twine version' to verify the update")
	} else {
		fmt.Println(result.Message)
	}

	return nil
}

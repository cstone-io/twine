package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// UpdateOptions configures the update behavior.
type UpdateOptions struct {
	// TargetVersion specifies a specific version to update to (e.g., "v1.0.0").
	// If empty, updates to the latest version.
	TargetVersion string

	// CurrentVersion is the version of the currently running binary.
	CurrentVersion string
}

// UpdateResult contains information about the update operation.
type UpdateResult struct {
	// Updated indicates whether an update was performed.
	Updated bool

	// FromVersion is the version before the update.
	FromVersion string

	// ToVersion is the version after the update.
	ToVersion string

	// Message provides additional context about the update.
	Message string
}

// Updater handles the self-update process.
type Updater struct {
	github *GitHubClient
}

// NewUpdater creates a new updater instance.
func NewUpdater() *Updater {
	return &Updater{
		github: NewGitHubClient(),
	}
}

// GetGitHubClient returns the GitHub client for accessing release information.
func (u *Updater) GetGitHubClient() *GitHubClient {
	return u.github
}

// CheckForUpdate checks if an update is available without downloading.
func (u *Updater) CheckForUpdate(currentVersion string) (*UpdateResult, error) {
	// Get the latest release
	release, err := u.github.GetLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	// Normalize and compare versions
	current := NormalizeVersion(currentVersion)
	latest := NormalizeVersion(release.TagName)

	if !IsNewer(current, latest) {
		return &UpdateResult{
			Updated:     false,
			FromVersion: currentVersion,
			ToVersion:   release.TagName,
			Message:     fmt.Sprintf("Already up-to-date (%s)", currentVersion),
		}, nil
	}

	return &UpdateResult{
		Updated:     false,
		FromVersion: currentVersion,
		ToVersion:   release.TagName,
		Message:     fmt.Sprintf("Update available: %s → %s", currentVersion, release.TagName),
	}, nil
}

// Update performs the self-update process.
func (u *Updater) Update(opts UpdateOptions) (*UpdateResult, error) {
	var release *GitHubRelease
	var err error

	// Fetch the target release
	if opts.TargetVersion != "" {
		targetVersion := NormalizeVersion(opts.TargetVersion)
		release, err = u.github.GetRelease(targetVersion)
		if err != nil {
			return nil, fmt.Errorf("version %s not found: %w", opts.TargetVersion, err)
		}
	} else {
		release, err = u.github.GetLatestRelease()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch latest release: %w", err)
		}
	}

	// Check if we're already at this version
	if !IsNewer(opts.CurrentVersion, release.TagName) && opts.TargetVersion == "" {
		return &UpdateResult{
			Updated:     false,
			FromVersion: opts.CurrentVersion,
			ToVersion:   release.TagName,
			Message:     fmt.Sprintf("Already up-to-date (%s)", opts.CurrentVersion),
		}, nil
	}

	// Find the binary for the current platform
	assetName := getBinaryName(runtime.GOOS, runtime.GOARCH)
	var assetURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			assetURL = asset.BrowserDownloadURL
			break
		}
	}

	if assetURL == "" {
		return nil, fmt.Errorf("no binary available for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, release.TagName)
	}

	// Download the new binary
	data, err := u.github.DownloadAsset(assetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download binary: %w", err)
	}

	// Install the new binary
	if err := installBinary(data); err != nil {
		return nil, fmt.Errorf("failed to install binary: %w", err)
	}

	return &UpdateResult{
		Updated:     true,
		FromVersion: opts.CurrentVersion,
		ToVersion:   release.TagName,
		Message:     fmt.Sprintf("Successfully updated from %s to %s", opts.CurrentVersion, release.TagName),
	}, nil
}

// getBinaryName returns the expected binary name for the given OS and architecture.
func getBinaryName(goos, goarch string) string {
	name := fmt.Sprintf("twine-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name
}

// installBinary performs an atomic replacement of the current binary.
func installBinary(data []byte) error {
	// Get the path of the current executable
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Resolve symlinks to get the real path
	currentPath, err = filepath.EvalSymlinks(currentPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Check if we have write permissions
	if err := checkWritePermissions(currentPath); err != nil {
		return err
	}

	// Create a temporary file in the same directory as the current binary
	// This ensures the temp file is on the same filesystem for atomic rename
	dir := filepath.Dir(currentPath)
	tmpFile, err := os.CreateTemp(dir, "twine-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file (try running with sudo): %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	// Write the new binary data
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write binary data: %w", err)
	}

	// Close the temp file before changing permissions
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Make the new binary executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpPath, 0755); err != nil {
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	// Atomically replace the current binary with the new one
	if err := os.Rename(tmpPath, currentPath); err != nil {
		return fmt.Errorf("failed to replace binary (try running with sudo): %w", err)
	}

	// Success - prevent cleanup from removing the file
	tmpFile = nil

	return nil
}

// checkWritePermissions verifies we have permission to write to the binary location.
func checkWritePermissions(path string) error {
	// Try to open the file for writing
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied. Try running with sudo: sudo twine update")
		}
		return err
	}
	file.Close()
	return nil
}

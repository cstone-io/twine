package updater

import (
	"strings"

	"golang.org/x/mod/semver"
)

// NormalizeVersion ensures the version string has a "v" prefix for semver comparison.
// If the version is "dev" or empty, it returns as-is.
func NormalizeVersion(v string) string {
	if v == "" || v == "dev" {
		return v
	}
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

// CompareVersions compares two semantic versions.
// Returns:
//   -1 if v1 < v2
//    0 if v1 == v2
//    1 if v1 > v2
//
// Special cases:
//   - "dev" is treated as always older than any versioned release
//   - Empty versions are treated as invalid and return 0
func CompareVersions(v1, v2 string) int {
	v1 = NormalizeVersion(v1)
	v2 = NormalizeVersion(v2)

	// Handle dev versions
	if v1 == "dev" && v2 == "dev" {
		return 0
	}
	if v1 == "dev" {
		return -1 // dev is always older
	}
	if v2 == "dev" {
		return 1 // anything is newer than dev
	}

	// Handle empty versions
	if v1 == "" || v2 == "" {
		return 0
	}

	// Use semver comparison
	return semver.Compare(v1, v2)
}

// IsNewer returns true if the new version is semantically newer than the current version.
// Returns false if versions are equal or if current is newer.
func IsNewer(current, new string) bool {
	return CompareVersions(current, new) < 0
}

// IsValid returns true if the version string is a valid semantic version.
// "dev" and empty strings are considered invalid for release purposes.
func IsValid(v string) bool {
	if v == "" || v == "dev" {
		return false
	}
	v = NormalizeVersion(v)
	return semver.IsValid(v)
}

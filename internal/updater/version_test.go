package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with v prefix", "v1.0.0", "v1.0.0"},
		{"without v prefix", "1.0.0", "v1.0.0"},
		{"dev version", "dev", "dev"},
		{"empty string", "", ""},
		{"with v and prerelease", "v1.0.0-alpha", "v1.0.0-alpha"},
		{"without v and prerelease", "1.0.0-alpha", "v1.0.0-alpha"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVersion(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		// Basic comparisons
		{"equal versions", "v1.0.0", "v1.0.0", 0},
		{"v1 less than v2", "v1.0.0", "v2.0.0", -1},
		{"v1 greater than v2", "v2.0.0", "v1.0.0", 1},

		// Minor version comparisons
		{"minor v1 less than v2", "v1.0.0", "v1.1.0", -1},
		{"minor v1 greater than v2", "v1.1.0", "v1.0.0", 1},

		// Patch version comparisons
		{"patch v1 less than v2", "v1.0.0", "v1.0.1", -1},
		{"patch v1 greater than v2", "v1.0.1", "v1.0.0", 1},

		// Pre-release versions
		{"prerelease less than release", "v1.0.0-alpha", "v1.0.0", -1},
		{"release greater than prerelease", "v1.0.0", "v1.0.0-alpha", 1},
		{"alpha less than beta", "v1.0.0-alpha", "v1.0.0-beta", -1},
		{"beta greater than alpha", "v1.0.0-beta", "v1.0.0-alpha", 1},

		// Without v prefix
		{"without prefix equal", "1.0.0", "1.0.0", 0},
		{"without prefix v1 less", "1.0.0", "2.0.0", -1},
		{"mixed prefix", "v1.0.0", "2.0.0", -1},

		// Dev versions
		{"dev equals dev", "dev", "dev", 0},
		{"dev less than release", "dev", "v1.0.0", -1},
		{"release greater than dev", "v1.0.0", "dev", 1},
		{"dev less than any version", "dev", "v0.0.1", -1},

		// Empty versions
		{"empty equals empty", "", "", 0},

		// Real-world version progressions
		{"0.1.0 to 0.2.0", "v0.1.0", "v0.2.0", -1},
		{"0.2.0 to 0.3.0", "v0.2.0", "v0.3.0", -1},
		{"0.3.0 to 1.0.0", "v0.3.0", "v1.0.0", -1},
		{"1.0.0 to 1.0.1", "v1.0.0", "v1.0.1", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result, "CompareVersions(%s, %s)", tt.v1, tt.v2)
		})
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		new      string
		expected bool
	}{
		{"newer version", "v1.0.0", "v2.0.0", true},
		{"same version", "v1.0.0", "v1.0.0", false},
		{"older version", "v2.0.0", "v1.0.0", false},
		{"dev to release", "dev", "v1.0.0", true},
		{"release to dev", "v1.0.0", "dev", false},
		{"minor bump", "v1.0.0", "v1.1.0", true},
		{"patch bump", "v1.0.0", "v1.0.1", true},
		{"prerelease to release", "v1.0.0-alpha", "v1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNewer(tt.current, tt.new)
			assert.Equal(t, tt.expected, result, "IsNewer(%s, %s)", tt.current, tt.new)
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"valid version with v", "v1.0.0", true},
		{"valid version without v", "1.0.0", true},
		{"valid prerelease", "v1.0.0-alpha", true},
		{"dev version", "dev", false},
		{"empty string", "", false},
		{"invalid format", "abc", false},
		{"invalid format 2", "not.a.version", false},
		{"valid complex prerelease", "v1.0.0-alpha.1", true},
		{"valid with build metadata", "v1.0.0+build123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValid(tt.version)
			assert.Equal(t, tt.expected, result, "IsValid(%s)", tt.version)
		})
	}
}

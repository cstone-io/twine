package template

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestFormatDate tests date formatting
func TestFormatDate(t *testing.T) {
	t.Run("formats date correctly", func(t *testing.T) {
		date := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := formatDate(date)
		assert.Equal(t, "2024-01-15", result)
	})

	t.Run("formats different dates", func(t *testing.T) {
		testCases := []struct {
			date     time.Time
			expected string
		}{
			{time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), "2024-12-31"},
			{time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), "2023-01-01"},
			{time.Date(2025, 6, 15, 12, 45, 30, 0, time.UTC), "2025-06-15"},
		}

		for _, tc := range testCases {
			result := formatDate(tc.date)
			assert.Equal(t, tc.expected, result)
		}
	})
}

// TestFormatDateTime tests datetime formatting
func TestFormatDateTime(t *testing.T) {
	t.Run("formats datetime correctly", func(t *testing.T) {
		date := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
		result := formatDateTime(date)
		assert.Equal(t, "2024-01-15 10:30:45", result)
	})

	t.Run("formats different datetimes", func(t *testing.T) {
		testCases := []struct {
			date     time.Time
			expected string
		}{
			{time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC), "2024-12-31 23:59:59"},
			{time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), "2023-01-01 00:00:00"},
			{time.Date(2025, 6, 15, 12, 45, 30, 0, time.UTC), "2025-06-15 12:45:30"},
		}

		for _, tc := range testCases {
			result := formatDateTime(tc.date)
			assert.Equal(t, tc.expected, result)
		}
	})
}

// TestMathFunctions tests math helper functions
func TestMathFunctions(t *testing.T) {
	t.Run("add function", func(t *testing.T) {
		assert.Equal(t, 8, add(5, 3))
		assert.Equal(t, 0, add(5, -5))
		assert.Equal(t, -8, add(-5, -3))
	})

	t.Run("sub function", func(t *testing.T) {
		assert.Equal(t, 2, sub(5, 3))
		assert.Equal(t, 10, sub(5, -5))
		assert.Equal(t, -2, sub(-5, -3))
	})

	t.Run("mul function", func(t *testing.T) {
		assert.Equal(t, 15, mul(5, 3))
		assert.Equal(t, -15, mul(5, -3))
		assert.Equal(t, 15, mul(-5, -3))
		assert.Equal(t, 0, mul(5, 0))
	})

	t.Run("div function", func(t *testing.T) {
		assert.Equal(t, 5, div(15, 3))
		assert.Equal(t, -5, div(15, -3))
		assert.Equal(t, 2, div(7, 3)) // Integer division
	})

	t.Run("mod function", func(t *testing.T) {
		assert.Equal(t, 2, mod(17, 5))
		assert.Equal(t, 0, mod(15, 5))
		assert.Equal(t, 1, mod(7, 3))
	})
}

// TestComparisonFunctions tests comparison helper functions
func TestComparisonFunctions(t *testing.T) {
	t.Run("eq function", func(t *testing.T) {
		assert.True(t, eq(5, 5))
		assert.True(t, eq("hello", "hello"))
		assert.False(t, eq(5, 3))
		assert.False(t, eq("hello", "world"))
	})

	t.Run("ne function", func(t *testing.T) {
		assert.True(t, ne(5, 3))
		assert.True(t, ne("hello", "world"))
		assert.False(t, ne(5, 5))
		assert.False(t, ne("hello", "hello"))
	})

	t.Run("lt function", func(t *testing.T) {
		assert.True(t, lt(3, 5))
		assert.False(t, lt(5, 3))
		assert.False(t, lt(5, 5))
	})

	t.Run("le function", func(t *testing.T) {
		assert.True(t, le(3, 5))
		assert.True(t, le(5, 5))
		assert.False(t, le(5, 3))
	})

	t.Run("gt function", func(t *testing.T) {
		assert.True(t, gt(5, 3))
		assert.False(t, gt(3, 5))
		assert.False(t, gt(5, 5))
	})

	t.Run("ge function", func(t *testing.T) {
		assert.True(t, ge(5, 3))
		assert.True(t, ge(5, 5))
		assert.False(t, ge(3, 5))
	})
}

// TestAsset tests asset path generation
func TestAsset(t *testing.T) {
	t.Run("generates correct asset path", func(t *testing.T) {
		assert.Equal(t, "/public/assets/style.css", asset("style.css"))
		assert.Equal(t, "/public/assets/app.js", asset("app.js"))
		assert.Equal(t, "/public/assets/images/logo.png", asset("images/logo.png"))
	})

	t.Run("handles empty string", func(t *testing.T) {
		assert.Equal(t, "/public/assets/", asset(""))
	})

	t.Run("handles paths with slashes", func(t *testing.T) {
		assert.Equal(t, "/public/assets/css/main.css", asset("css/main.css"))
		assert.Equal(t, "/public/assets/js/vendor/jquery.js", asset("js/vendor/jquery.js"))
	})
}

// TestFuncMap tests FuncMap registration
func TestFuncMap(t *testing.T) {
	t.Run("contains all helper functions", func(t *testing.T) {
		funcMap := FuncMap()

		expectedFuncs := []string{
			"formatDate",
			"formatDateTime",
			"add",
			"sub",
			"mul",
			"div",
			"mod",
			"eq",
			"ne",
			"lt",
			"le",
			"gt",
			"ge",
			"asset",
		}

		for _, name := range expectedFuncs {
			_, exists := funcMap[name]
			assert.True(t, exists, "FuncMap should contain %s", name)
		}
	})

	t.Run("functions are callable", func(t *testing.T) {
		funcMap := FuncMap()

		// Test formatDate
		formatDateFunc := funcMap["formatDate"].(func(time.Time) string)
		date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, "2024-01-01", formatDateFunc(date))

		// Test add
		addFunc := funcMap["add"].(func(int, int) int)
		assert.Equal(t, 8, addFunc(5, 3))

		// Test asset
		assetFunc := funcMap["asset"].(func(string) string)
		assert.Equal(t, "/public/assets/app.css", assetFunc("app.css"))
	})
}

// TestHelpers_Integration tests helpers in realistic scenarios
func TestHelpers_Integration(t *testing.T) {
	t.Run("combined math operations", func(t *testing.T) {
		// ((10 + 5) * 2) / 3 - 1
		result := sub(div(mul(add(10, 5), 2), 3), 1)
		assert.Equal(t, 9, result)
	})

	t.Run("comparison chain", func(t *testing.T) {
		assert.True(t, lt(3, 5) && le(5, 5) && gt(5, 3))
		assert.True(t, ne(5, 3) && eq(5, 5))
	})

	t.Run("date formatting for display", func(t *testing.T) {
		now := time.Now()

		dateStr := formatDate(now)
		assert.Contains(t, dateStr, "-")
		assert.Len(t, dateStr, 10) // YYYY-MM-DD

		dateTimeStr := formatDateTime(now)
		assert.Contains(t, dateTimeStr, " ")
		assert.Len(t, dateTimeStr, 19) // YYYY-MM-DD HH:MM:SS
	})

	t.Run("asset paths for different file types", func(t *testing.T) {
		cssPath := asset("css/style.css")
		jsPath := asset("js/app.js")
		imgPath := asset("images/logo.png")

		assert.Contains(t, cssPath, "/public/assets/")
		assert.Contains(t, jsPath, "/public/assets/")
		assert.Contains(t, imgPath, "/public/assets/")
	})
}

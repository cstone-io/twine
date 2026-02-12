package template

import (
	"bytes"
	"html/template"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetTemplates resets the global templates for testing
func resetTemplates() {
	templateMutex.Lock()
	defer templateMutex.Unlock()
	templates = nil
}

// TestLoadTemplates tests template loading
func TestLoadTemplates(t *testing.T) {
	t.Run("loads single pattern successfully", func(t *testing.T) {
		resetTemplates()

		pattern := filepath.Join("testdata", "test.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		tmpl := GetTemplates()
		assert.NotNil(t, tmpl)
	})

	t.Run("loads multiple patterns", func(t *testing.T) {
		resetTemplates()

		pattern1 := filepath.Join("testdata", "test.html")
		pattern2 := filepath.Join("testdata", "partial.html")

		err := LoadTemplates(pattern1, pattern2)
		require.NoError(t, err)

		tmpl := GetTemplates()
		assert.NotNil(t, tmpl)
	})

	t.Run("returns error for invalid pattern", func(t *testing.T) {
		resetTemplates()

		err := LoadTemplates("/nonexistent/*.html")
		assert.Error(t, err)
	})

	t.Run("loads templates with FuncMap", func(t *testing.T) {
		resetTemplates()

		pattern := filepath.Join("testdata", "helpers.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		tmpl := GetTemplates()
		assert.NotNil(t, tmpl)
	})

	t.Run("overwrites previously loaded templates", func(t *testing.T) {
		resetTemplates()

		// Load first set
		pattern1 := filepath.Join("testdata", "test.html")
		err := LoadTemplates(pattern1)
		require.NoError(t, err)

		// Load second set (should overwrite)
		pattern2 := filepath.Join("testdata", "partial.html")
		err = LoadTemplates(pattern2)
		require.NoError(t, err)

		tmpl := GetTemplates()
		assert.NotNil(t, tmpl)
	})
}

// TestSetTemplates tests custom template setting
func TestSetTemplates(t *testing.T) {
	t.Run("sets custom template", func(t *testing.T) {
		resetTemplates()

		customTmpl := template.New("custom")
		SetTemplates(customTmpl)

		retrieved := GetTemplates()
		assert.Equal(t, customTmpl, retrieved)
	})

	t.Run("can set nil template", func(t *testing.T) {
		resetTemplates()

		SetTemplates(nil)
		assert.Nil(t, GetTemplates())
	})
}

// TestGetTemplates tests template retrieval
func TestGetTemplates(t *testing.T) {
	t.Run("returns nil when not loaded", func(t *testing.T) {
		resetTemplates()

		tmpl := GetTemplates()
		assert.Nil(t, tmpl)
	})

	t.Run("returns loaded templates", func(t *testing.T) {
		resetTemplates()

		pattern := filepath.Join("testdata", "test.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		tmpl := GetTemplates()
		assert.NotNil(t, tmpl)
	})
}

// TestRenderFull tests full page template rendering
func TestRenderFull(t *testing.T) {
	t.Run("renders template successfully", func(t *testing.T) {
		resetTemplates()

		pattern := filepath.Join("testdata", "test.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		var buf bytes.Buffer
		data := map[string]string{"Name": "World"}

		err = RenderFull(&buf, "test", data)
		require.NoError(t, err)
		assert.Equal(t, "Hello World", buf.String())
	})

	t.Run("returns error when templates not loaded", func(t *testing.T) {
		resetTemplates()

		var buf bytes.Buffer
		err := RenderFull(&buf, "test", nil)
		assert.Error(t, err) // Empty template execution fails
	})

	t.Run("returns error for missing template", func(t *testing.T) {
		resetTemplates()

		pattern := filepath.Join("testdata", "test.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		var buf bytes.Buffer
		err = RenderFull(&buf, "nonexistent", nil)
		assert.Error(t, err)
	})

	t.Run("renders with different data types", func(t *testing.T) {
		resetTemplates()

		pattern := filepath.Join("testdata", "test.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		testCases := []struct {
			name string
			data any
		}{
			{"map", map[string]string{"Name": "Alice"}},
			{"struct", struct{ Name string }{"Bob"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var buf bytes.Buffer
				err := RenderFull(&buf, "test", tc.data)
				require.NoError(t, err)
				assert.Contains(t, buf.String(), "Hello")
			})
		}
	})
}

// TestRenderPartial tests partial template rendering
func TestRenderPartial(t *testing.T) {
	t.Run("renders partial successfully", func(t *testing.T) {
		resetTemplates()

		pattern := filepath.Join("testdata", "partial.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		var buf bytes.Buffer
		data := map[string]string{"Text": "Click me"}

		err = RenderPartial(&buf, "button", data)
		require.NoError(t, err)
		assert.Equal(t, "<button>Click me</button>", buf.String())
	})

	t.Run("returns error when templates not loaded", func(t *testing.T) {
		resetTemplates()

		var buf bytes.Buffer
		err := RenderPartial(&buf, "button", nil)
		assert.Error(t, err) // Empty template execution fails
	})

	t.Run("renders component template", func(t *testing.T) {
		resetTemplates()

		pattern := filepath.Join("testdata", "partial.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		var buf bytes.Buffer
		data := map[string]string{"Text": "Submit"}

		err = RenderPartial(&buf, "button", data)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "<button>")
		assert.Contains(t, buf.String(), "Submit")
	})
}

// TestReload tests template reloading
func TestReload(t *testing.T) {
	t.Run("reloads templates", func(t *testing.T) {
		resetTemplates()

		pattern := filepath.Join("testdata", "test.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		// Reload
		err = Reload(pattern)
		require.NoError(t, err)

		tmpl := GetTemplates()
		assert.NotNil(t, tmpl)
	})

	t.Run("returns error for invalid pattern on reload", func(t *testing.T) {
		resetTemplates()

		err := Reload("/nonexistent/*.html")
		assert.Error(t, err)
	})
}

// TestTemplate_ThreadSafety tests concurrent template operations
func TestTemplate_ThreadSafety(t *testing.T) {
	t.Run("concurrent reads are safe", func(t *testing.T) {
		resetTemplates()

		pattern := filepath.Join("testdata", "test.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				tmpl := GetTemplates()
				assert.NotNil(t, tmpl)
			}()
		}

		wg.Wait()
	})

	t.Run("concurrent renders are safe", func(t *testing.T) {
		resetTemplates()

		pattern := filepath.Join("testdata", "test.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var buf bytes.Buffer
				data := map[string]string{"Name": "Test"}
				err := RenderFull(&buf, "test", data)
				assert.NoError(t, err)
			}()
		}

		wg.Wait()
	})
}

// TestTemplate_Integration tests realistic template scenarios
func TestTemplate_Integration(t *testing.T) {
	t.Run("full template workflow", func(t *testing.T) {
		resetTemplates()

		// Load templates
		pattern := filepath.Join("testdata", "*.html")
		err := LoadTemplates(pattern)
		require.NoError(t, err)

		// Render full page
		var fullBuf bytes.Buffer
		err = RenderFull(&fullBuf, "test", map[string]string{"Name": "User"})
		require.NoError(t, err)
		assert.Contains(t, fullBuf.String(), "Hello User")

		// Render partial
		var partialBuf bytes.Buffer
		err = RenderPartial(&partialBuf, "button", map[string]string{"Text": "Action"})
		require.NoError(t, err)
		assert.Contains(t, partialBuf.String(), "<button>Action</button>")
	})
}

package template

import (
	"html/template"
	"io"
	"sync"
)

var (
	templates     *template.Template
	templateMutex sync.RWMutex
)

// LoadTemplates loads all templates from the given patterns
func LoadTemplates(patterns ...string) error {
	templateMutex.Lock()
	defer templateMutex.Unlock()

	tmpl, err := template.New("").Funcs(FuncMap()).ParseGlob(patterns[0])
	if err != nil {
		return err
	}

	// Parse additional patterns if provided
	for i := 1; i < len(patterns); i++ {
		tmpl, err = tmpl.ParseGlob(patterns[i])
		if err != nil {
			return err
		}
	}

	templates = tmpl
	return nil
}

// SetTemplates allows users to set a custom template instance
func SetTemplates(tmpl *template.Template) {
	templateMutex.Lock()
	defer templateMutex.Unlock()
	templates = tmpl
}

// GetTemplates returns the current template instance
func GetTemplates() *template.Template {
	templateMutex.RLock()
	defer templateMutex.RUnlock()
	return templates
}

// RenderFull renders a full page template
func RenderFull(w io.Writer, name string, data any) error {
	templateMutex.RLock()
	defer templateMutex.RUnlock()

	if templates == nil {
		return template.New("").Execute(w, "Templates not loaded. Call template.LoadTemplates() first.")
	}

	return templates.ExecuteTemplate(w, name, data)
}

// RenderPartial renders a template component (for Ajax partial responses)
func RenderPartial(w io.Writer, name string, data any) error {
	templateMutex.RLock()
	defer templateMutex.RUnlock()

	if templates == nil {
		return template.New("").Execute(w, "Templates not loaded. Call template.LoadTemplates() first.")
	}

	return templates.ExecuteTemplate(w, name, data)
}

// Reload reloads templates from the same patterns (useful in development)
func Reload(patterns ...string) error {
	return LoadTemplates(patterns...)
}

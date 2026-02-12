package kit

import (
	"encoding/json"
	"net/http"

	"github.com/cstone-io/twine/pkg/template"
)

// JSON writes a JSON response
func (k *Kit) JSON(status int, v any) error {
	k.Response.Header().Set("Content-Type", "application/json")
	k.Response.WriteHeader(status)
	return json.NewEncoder(k.Response).Encode(v)
}

// Text writes a plain text response
func (k *Kit) Text(status int, msg string) error {
	k.Response.Header().Set("Content-Type", "text/plain")
	k.Response.WriteHeader(status)
	_, err := k.Response.Write([]byte(msg))
	return err
}

// Bytes writes raw bytes as a response
func (k *Kit) Bytes(status int, b []byte) error {
	k.Response.Header().Set("Content-Type", "text/plain")
	k.Response.WriteHeader(status)
	_, err := k.Response.Write(b)
	return err
}

// HTML writes raw HTML content
func (k *Kit) HTML(status int, htmlContent string) error {
	k.Response.Header().Set("Content-Type", "text/html")
	k.Response.WriteHeader(status)
	_, err := k.Response.Write([]byte(htmlContent))
	return err
}

// NoContent writes a 204 No Content response
func (k *Kit) NoContent() error {
	k.Response.WriteHeader(http.StatusNoContent)
	return nil
}

// RenderTemplate renders a full page template
func (k *Kit) RenderTemplate(name string, data any) error {
	k.Response.Header().Set("Content-Type", "text/html")
	return template.RenderFull(k.Response, name, data)
}

// RenderPartial renders a template component (for Ajax partial responses)
func (k *Kit) RenderPartial(name string, data any) error {
	k.Response.Header().Set("Content-Type", "text/html")
	return template.RenderPartial(k.Response, name, data)
}

// Render automatically chooses between full and partial rendering based on X-Alpine-Request header
func (k *Kit) Render(name string, data any) error {
	if k.IsAjax() {
		return k.RenderPartial(name, data)
	}
	return k.RenderTemplate(name, data)
}

// IsAjax returns true if the request is an Ajax request (from Alpine Ajax)
func (k *Kit) IsAjax() bool {
	return len(k.Request.Header.Get("X-Alpine-Request")) > 0
}

// Redirect performs an HTTP redirect using standard Location header
func (k *Kit) Redirect(url string) error {
	http.Redirect(k.Response, k.Request, url, http.StatusSeeOther)
	return nil
}

package public

import (
	"embed"
	"net/http"
	"strings"
)

// AssetsFS should be set by the user application using //go:embed
// Example in user's code:
//
//	//go:embed assets
//	var AssetsFS embed.FS
//
//	func init() {
//	    public.AssetsFS = AssetsFS
//	}
var AssetsFS embed.FS

const (
	AssetsPath = "/public/assets/"
	PublicPath = "/public/"
)

// FileServerHandler returns an HTTP handler for serving embedded static files
func FileServerHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, PublicPath) {
			http.StripPrefix(PublicPath, http.FileServer(http.FS(AssetsFS))).ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}

// Asset returns the path to a static asset
func Asset(name string) string {
	return AssetsPath + name
}

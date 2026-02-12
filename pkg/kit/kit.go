package kit

import (
	"net/http"
)

// Kit wraps http.ResponseWriter and *http.Request for convenient access
type Kit struct {
	Response http.ResponseWriter
	Request  *http.Request
}

// HandlerFunc is the signature for Twine handlers that return errors
type HandlerFunc func(kit *Kit) error

// Handler converts a Kit.HandlerFunc to an http.HandlerFunc
func Handler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kit := &Kit{
			Response: w,
			Request:  r,
		}
		if err := h(kit); err != nil {
			if errorHandler != nil {
				errorHandler(kit, err)
				return
			}
			kit.Text(http.StatusInternalServerError, err.Error())
		}
	}
}

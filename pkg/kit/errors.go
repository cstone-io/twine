package kit

import (
	"net/http"

	"github.com/cstone-io/twine/pkg/errors"
	"github.com/cstone-io/twine/pkg/logger"
)

// ErrorHandlerFunc is the signature for custom error handlers
type ErrorHandlerFunc func(kit *Kit, err error)

// UseErrorHandler sets a custom error handler for all Kit handlers
func UseErrorHandler(h ErrorHandlerFunc) {
	errorHandler = h
}

var (
	errorHandler = func(kit *Kit, err error) {
		if e, ok := err.(*errors.Error); ok {
			logger.Get().CustomError(e)
			// If user has set up templates, they can render an error page
			// For now, return JSON error
			status := e.HTTPStatus
			if status == 0 {
				status = http.StatusInternalServerError
			}
			kit.JSON(status, map[string]any{
				"error":  e.Message,
				"code":   e.Code,
				"status": e.HTTPStatus,
			})
		} else {
			e := errors.ErrDefaultError.Wrap(err)
			logger.Get().CustomError(e)
			kit.JSON(http.StatusInternalServerError, map[string]any{
				"error": e.Message,
				"code":  e.Code,
			})
		}
	}
)

// NotFoundHandler returns a handler for 404 errors
func NotFoundHandler() http.HandlerFunc {
	return Handler(func(kit *Kit) error {
		return errors.ErrNotFound
	})
}

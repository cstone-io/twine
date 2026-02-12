package middleware

import (
	"context"
	"time"

	"github.com/cstone-io/twine/pkg/kit"
	"github.com/cstone-io/twine/pkg/logger"
)

// LoggingMiddleware logs incoming requests
func LoggingMiddleware() Middleware {
	return func(next kit.HandlerFunc) kit.HandlerFunc {
		return func(k *kit.Kit) error {
			logger.Get().Info("Request: %s %s", k.Request.Method, k.Request.URL.Path)
			return next(k)
		}
	}
}

// TimeoutMiddleware adds a timeout to request processing
func TimeoutMiddleware(d time.Duration) Middleware {
	return func(next kit.HandlerFunc) kit.HandlerFunc {
		return func(k *kit.Kit) error {
			ctx, cancel := context.WithTimeout(k.Request.Context(), d)
			defer cancel()

			k.Request = k.Request.WithContext(ctx)
			return next(k)
		}
	}
}

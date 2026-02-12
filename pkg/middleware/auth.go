package middleware

import (
	"github.com/cstone-io/twine/pkg/auth"
	"github.com/cstone-io/twine/pkg/kit"
)

// JWTMiddleware validates JWT tokens and auto-redirects on failure
func JWTMiddleware() Middleware {
	return func(next kit.HandlerFunc) kit.HandlerFunc {
		return func(k *kit.Kit) error {
			token, err := k.Authorization()
			if err != nil {
				return k.Redirect("/auth/login")
			}

			userID, err := auth.ParseToken(token)
			if err != nil {
				return k.Redirect("/auth/login")
			}

			k.SetContext("user", userID)
			return next(k)
		}
	}
}

package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/cstone-io/twine/pkg/config"
	"github.com/cstone-io/twine/pkg/errors"
)

// Token wraps a JWT token string
type Token struct {
	Token string `json:"token"`
}

// NewToken generates a new JWT token for a user
func NewToken(userID uuid.UUID, email string) (*Token, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	cfg := config.Get()
	key := cfg.Auth.SecretKey

	signed, err := token.SignedString([]byte(key))
	if err != nil {
		return nil, errors.ErrGenerateToken.Wrap(err).WithValue(signed)
	}

	return &Token{Token: signed}, nil
}

// ParseToken validates and parses a JWT token, returning the user ID
func ParseToken(tokenString string) (string, error) {
	cfg := config.Get()
	key := cfg.Auth.SecretKey

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http.ErrAbortHandler
		}
		return []byte(key), nil
	})

	if err != nil || !token.Valid {
		return "", errors.ErrAuthInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.ErrAuthInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.ErrAuthInvalidToken
	}

	return userID, nil
}

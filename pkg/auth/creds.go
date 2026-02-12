package auth

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/cstone-io/twine/pkg/errors"
)

// Credentials holds user authentication credentials
type Credentials struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

// Authenticate compares a password with a stored hash
func (creds *Credentials) Authenticate(hashedPassword string) error {
	if err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(creds.Password),
	); err != nil {
		return errors.ErrAuthInvalidCredentials.Wrap(err)
	}
	return nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.ErrHashPassword.Wrap(err)
	}
	return string(hash), nil
}

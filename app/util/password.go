package util

import (
	"strings"

	"github.com/music-gang/music-gang-api/app/apperr"
	"golang.org/x/crypto/bcrypt"
)

const (
	minLength = 8
	maxLength = 64

	upperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerLetters = "abcdefghijklmnopqrstuvwxyz"
	numbers      = "0123456789"
	specialChars = "!@#$%^&*()-_=+[{]};:,<.>/?|"

	PasswordRequirements = "Password must be at least 8 characters but no more than 64 characters. " +
		"It must contain at least one lower case letter, one upper case letter, one number and one special character."
)

// CompareHashAndPassword compares the hash and password.
// It is a wrapper around bcrypt.CompareHashAndPassword.
func CompareHashAndPassword(hashedPassword, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

// HashPassword hashes the password.
func HashPassword(password string) ([]byte, error) {
	p, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperr.Errorf(apperr.EINTERNAL, "failed to hash password: %v", err)
	}
	return p, nil
}

// isPasswordValid checks if the password is valid.
// the password must be at least 8 characters but no more than 64 characters.
// it must contain at least one lower case letter, one upper case letter, one number and one special character.
func IsValidPassword(password string) bool {
	return len(password) >= minLength && len(password) <= maxLength &&
		strings.ContainsAny(password, lowerLetters) &&
		strings.ContainsAny(password, upperLetters) &&
		strings.ContainsAny(password, numbers) &&
		strings.ContainsAny(password, specialChars)
}

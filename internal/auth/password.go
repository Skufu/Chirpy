package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword takes a plain text password and returns a bcrypt hash
func HashPassword(password string) (string, error) {
	// Generate a bcrypt hash using the default cost (10)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Return the hash as a string
	return string(bytes), nil
}

// CheckPasswordHash compares a bcrypt hashed password with a plain text password
// Returns nil if the passwords match, or an error if they don't
func CheckPasswordHash(hash, password string) error {
	// Compare the stored hash with the provided password
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

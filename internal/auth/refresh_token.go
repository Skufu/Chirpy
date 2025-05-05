package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// MakeRefreshToken generates a random 256-bit (32-byte) hex-encoded string
// that can be used as a refresh token.
func MakeRefreshToken() (string, error) {
	// Create a byte slice to store the random data
	randomBytes := make([]byte, 32)

	// Generate random data
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Convert the random data to a hex string
	tokenString := hex.EncodeToString(randomBytes)

	return tokenString, nil
}

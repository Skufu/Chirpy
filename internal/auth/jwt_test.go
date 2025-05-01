package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	// Create a sample user ID
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := time.Minute * 10

	// Test making a JWT
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	// Test that token is not empty
	if tokenString == "" {
		t.Fatal("MakeJWT returned empty token")
	}

	// Test validating the JWT
	extractedID, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}

	// Test that the extracted ID matches the original
	if extractedID != userID {
		t.Fatalf("ValidateJWT returned wrong user ID, got %v, want %v", extractedID, userID)
	}
}

func TestExpiredToken(t *testing.T) {
	// Create a token that expires immediately
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := -time.Minute // Negative duration to make token already expired

	// Create an expired token
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	// Validate the expired token - should fail
	_, err = ValidateJWT(tokenString, tokenSecret)
	if err == nil {
		t.Fatal("ValidateJWT should reject expired token")
	}
}

func TestInvalidSecret(t *testing.T) {
	// Create a token with one secret
	userID := uuid.New()
	tokenSecret := "correct-secret"
	wrongSecret := "wrong-secret"
	expiresIn := time.Minute * 10

	// Create a token
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	// Validate with wrong secret - should fail
	_, err = ValidateJWT(tokenString, wrongSecret)
	if err == nil {
		t.Fatal("ValidateJWT should reject token signed with wrong secret")
	}
}

func TestInvalidToken(t *testing.T) {
	// Test with an invalid token string
	invalidToken := "not.a.valid.jwt"
	tokenSecret := "test-secret"

	// Validate invalid token - should fail
	_, err := ValidateJWT(invalidToken, tokenSecret)
	if err == nil {
		t.Fatal("ValidateJWT should reject invalid token format")
	}
}

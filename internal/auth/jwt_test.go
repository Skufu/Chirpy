package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		headers       http.Header
		expectedToken string
		expectError   bool
	}{
		{
			name: "Valid Bearer Token",
			headers: http.Header{
				"Authorization": []string{"Bearer abc123.def456.ghi789"},
			},
			expectedToken: "abc123.def456.ghi789",
			expectError:   false,
		},
		{
			name: "Missing Authorization Header",
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			expectedToken: "",
			expectError:   true,
		},
		{
			name: "Invalid Authorization Format",
			headers: http.Header{
				"Authorization": []string{"Basic dXNlcjpwYXNz"},
			},
			expectedToken: "",
			expectError:   true,
		},
		{
			name: "Empty Token",
			headers: http.Header{
				"Authorization": []string{"Bearer "},
			},
			expectedToken: "",
			expectError:   true,
		},
		{
			name: "Token with Extra Spaces",
			headers: http.Header{
				"Authorization": []string{"Bearer  token123  "},
			},
			expectedToken: "token123",
			expectError:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := GetBearerToken(tc.headers)

			// Check error expectation
			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Did not expect error but got: %v", err)
			}

			// If we don't expect an error, check the token
			if !tc.expectError {
				if token != tc.expectedToken {
					t.Fatalf("Expected token %q but got %q", tc.expectedToken, token)
				}
			}
		})
	}
}

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

package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "test-password"

	// Test that hash can be generated
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	// Test that hash is not empty
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}

	// Test that hash is not the same as the original password
	if hash == password {
		t.Fatal("Hash should not be the same as the original password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "test-password"
	wrongPassword := "wrong-password"

	// Generate a hash
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	// Test that correct password validates
	err = CheckPasswordHash(hash, password)
	if err != nil {
		t.Fatalf("CheckPasswordHash should validate correct password, got error: %v", err)
	}

	// Test that incorrect password fails
	err = CheckPasswordHash(hash, wrongPassword)
	if err == nil {
		t.Fatal("CheckPasswordHash should reject incorrect password")
	}
}

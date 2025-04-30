package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse the request body
	decoder := json.NewDecoder(r.Body)
	reqBody := requestBody{}
	err := decoder.Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if reqBody.Email == "" || reqBody.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Look up the user by email
	user, err := cfg.db.GetUserByEmail(context.Background(), reqBody.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// User not found, but don't reveal that information
			log.Printf("Login attempt for non-existent email: %s", reqBody.Email)
			respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
			return
		}
		// Other database error
		log.Printf("Database error during login: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error during login")
		return
	}

	// Check if the password matches
	err = auth.CheckPasswordHash(user.HashedPassword, reqBody.Password)
	if err != nil {
		// Password doesn't match
		log.Printf("Invalid password attempt for user: %s", reqBody.Email)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	// Login successful - return user without the password
	respondWithJSON(w, http.StatusOK, UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

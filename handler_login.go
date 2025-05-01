package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	type requestBody struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
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

	// Determine token expiration time
	// Default is 1 hour (3600 seconds)
	expiresIn := time.Hour

	// If the client specified an expiration time
	if reqBody.ExpiresInSeconds > 0 {
		// Use client's expiration time, but cap at 1 hour
		requestedExpiration := time.Duration(reqBody.ExpiresInSeconds) * time.Second
		if requestedExpiration > time.Hour {
			// Cap at 1 hour
			expiresIn = time.Hour
		} else {
			expiresIn = requestedExpiration
		}
	}

	// Create JWT token
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expiresIn)
	if err != nil {
		log.Printf("Error creating JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error during login")
		return
	}

	// Define response type with token
	type loginResponse struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Token     string    `json:"token"`
	}

	// Login successful - return user with token
	respondWithJSON(w, http.StatusOK, loginResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Token:     token,
	})
}

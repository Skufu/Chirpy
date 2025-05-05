package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/auth"
	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/database"
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

	// Access tokens expire after 1 hour
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		log.Printf("Error creating JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error during login")
		return
	}

	// Generate refresh token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error creating refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error during login")
		return
	}

	// Store refresh token in database with 60 day expiration
	_, err = cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60), // 60 days
	})
	if err != nil {
		log.Printf("Error storing refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error during login")
		return
	}

	// Define response type with token and refresh token
	type loginResponse struct {
		ID           string    `json:"id"`
		Email        string    `json:"email"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	// Login successful - return user with tokens
	respondWithJSON(w, http.StatusOK, loginResponse{
		ID:           user.ID.String(),
		Email:        user.Email,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

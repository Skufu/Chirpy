package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract the refresh token from the Authorization header
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing refresh token")
		return
	}

	// Get the user associated with the refresh token
	// This also validates that the token exists, is not expired, and is not revoked
	user, err := cfg.db.GetUserFromRefreshToken(context.Background(), refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Invalid, expired, or revoked refresh token")
			return
		}
		log.Printf("Error looking up refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error processing refresh token")
		return
	}

	// Generate a new access token for the user
	newAccessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		log.Printf("Error creating JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error generating access token")
		return
	}

	// Return the new access token
	type refreshResponse struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, refreshResponse{
		Token: newAccessToken,
	})
}

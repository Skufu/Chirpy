package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
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

	// Revoke the token in the database
	err = cfg.db.RevokeRefreshToken(context.Background(), refreshToken)
	if err != nil {
		log.Printf("Error revoking refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error revoking token")
		return
	}

	// Return 204 No Content status
	w.WriteHeader(http.StatusNoContent)
}

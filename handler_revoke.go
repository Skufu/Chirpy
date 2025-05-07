package main

import (
	"log"
	"net/http"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	// Extract the refresh token from the Authorization header
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing refresh token")
		return
	}

	// Revoke the token in the database
	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("Error revoking refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error revoking token")
		return
	}

	// Return 204 No Content status
	w.WriteHeader(http.StatusNoContent)
}

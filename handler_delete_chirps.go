package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	//get Bearer token
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Invalid JWT: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT")
		return
	}

	//validate jwt and extractuserID
	userID, err := auth.ValidateJWT(bearerToken, cfg.jwtSecret)
	if err != nil {
		log.Printf("Invalid JWT: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT")
		return
	}

	//extract chirp id from path
	chirpIDStr := r.PathValue("chirpID")
	if chirpIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp ID cannot be empty")
		return
	}

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		log.Printf("Invalid chirp ID format: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format")
		return
	}

	// Get the chirp to verify ownership
	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
		} else {
			log.Printf("Error getting chirp: %s", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to get chirp")
		}
		return
	}

	// Check if the authenticated user is the author of the chirp
	if chirp.UserID != userID {
		log.Printf("User %s attempted to delete chirp %s owned by user %s", userID, chirpID, chirp.UserID)
		respondWithError(w, http.StatusForbidden, "You are not authorized to delete this chirp")
		return
	}

	// Delete the chirp
	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error deleting chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

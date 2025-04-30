package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	// Only handle GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract chirp ID from URL path
	// The URL will be like /api/chirps/91c19d70-286e-4924-b399-da1dd0fb5596
	path := r.URL.Path
	pathParts := strings.Split(path, "/")

	// Path should have format /api/chirps/{id}
	if len(pathParts) != 4 {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	chirpIDStr := pathParts[3]

	// Parse the ID into a UUID
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		log.Printf("Invalid chirp ID format: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	// Fetch the chirp from the database
	chirp, err := cfg.db.GetChirp(context.Background(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Chirp not found
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		// Other database error
		log.Printf("Error fetching chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch chirp")
		return
	}

	// Return the chirp in the response format
	respondWithJSON(w, http.StatusOK, chirpResponse{
		ID:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.String(),
	})
}

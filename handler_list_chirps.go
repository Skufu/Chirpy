package main

import (
	"log"
	"net/http"
	"sort"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerListChirps(w http.ResponseWriter, r *http.Request) {
	var chirps []database.Chirp
	var err error

	// Check if author_id query parameter is provided
	sortParam := r.URL.Query().Get("sort")
	authorIDStr := r.URL.Query().Get("author_id")

	if authorIDStr != "" {
		// Parse the author_id to UUID
		authorID, parseErr := uuid.Parse(authorIDStr)
		if parseErr != nil {
			log.Printf("Invalid author_id format: %s", parseErr)
			respondWithError(w, http.StatusBadRequest, "Invalid author_id format")
			return
		}

		// Get chirps filtered by author_id
		chirps, err = cfg.db.GetChirpByUserID(r.Context(), authorID)
	} else {
		// Get all chirps (original behavior)
		chirps, err = cfg.db.ListChirps(r.Context())
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch chirps")
		return
	}

	if sortParam == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	}
	// Convert database chirps to response format
	chirpsResponse := []chirpResponse{}
	for _, chirp := range chirps {
		chirpsResponse = append(chirpsResponse, chirpResponse{
			ID:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID.String(),
		})
	}

	// Respond with the chirps array
	respondWithJSON(w, http.StatusOK, chirpsResponse)
}

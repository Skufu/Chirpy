package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerListChirps(w http.ResponseWriter, r *http.Request) {
	// Get all chirps from the database
	chirps, err := cfg.db.ListChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch chirps")
		return
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

package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	// Check if in dev environment
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}

	// Delete chirps first due to the foreign key constraint
	err := cfg.db.DeleteAllChirps(r.Context())
	if err != nil {
		log.Printf("Failed to reset chirps: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to reset chirps")
		return
	}

	// Delete users after chirps
	err = cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		log.Printf("Failed to reset users: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to reset users")
		return
	}

	// Reset hits counter
	cfg.fileserverHits.Store(0)

	respondWithJSON(w, http.StatusOK, map[string]string{
		"status": "Reset successful",
	})
}

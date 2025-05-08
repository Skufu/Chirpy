package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	// Define the structure of the incoming webhook payload
	type webhookData struct {
		UserID string `json:"user_id"`
	}

	type webhookPayload struct {
		Event string      `json:"event"`
		Data  webhookData `json:"data"`
	}

	// Parse the request body
	var payload webhookPayload
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// If the event is not "user.upgraded", respond with 204 and return
	if payload.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Parse the user ID
	userID, err := uuid.Parse(payload.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// Attempt to update the user to Chirpy Red
	_, err = cfg.db.UpdateUserChirpyBasedOnID(r.Context(), userID)
	if err != nil {
		// If the user doesn't exist, return 404
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		// For other errors, return 500
		log.Printf("Error upgrading user to Chirpy Red: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to upgrade user")
		return
	}

	// Success: return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

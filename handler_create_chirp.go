package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/auth"
	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/database"
)

type createChirpRequest struct {
	Body string `json:"body"`
}

type chirpResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
}

// List of profane words to filter
var profaneWords = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

// cleanChirp removes profane words from a chirp
func cleanChirp(body string) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		// Convert to lowercase for case-insensitive comparison
		// but only compare the word itself without punctuation
		wordLower := strings.ToLower(word)

		// Check if the word is in the profane words list
		for _, profane := range profaneWords {
			if wordLower == profane {
				words[i] = "****"
				break
			}
		}
	}
	return strings.Join(words, " ")
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	// Authenticate the request using JWT
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Authentication error: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Validate the JWT and extract user ID
	userID, err := auth.ValidateJWT(bearerToken, cfg.jwtSecret)
	if err != nil {
		log.Printf("Invalid JWT: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid authentication token")
		return
	}

	// Parse request body
	decoder := json.NewDecoder(r.Body)
	params := createChirpRequest{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate chirp length
	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	// Clean the chirp body by replacing profane words
	cleanedBody := cleanChirp(params.Body)

	// Save the chirp to the database using the user ID from the JWT
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp")
		return
	}

	// Respond with the created chirp
	respondWithJSON(w, http.StatusCreated, chirpResponse{
		ID:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.String(),
	})
}

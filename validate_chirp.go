package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type parameters struct {
	Body string `json:"body"`
}

type validResponse struct {
	Valid bool `json:"valid"`
}

type cleanedResponse struct {
	CleanedBody string `json:"cleaned_body"`
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

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	// Clean the chirp body by replacing profane words
	cleanedBody := cleanChirp(params.Body)

	// Return the cleaned body
	respondWithJSON(w, http.StatusOK, cleanedResponse{
		CleanedBody: cleanedBody,
	})
}

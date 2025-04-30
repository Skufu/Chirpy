package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/auth"
	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/database"
	"github.com/google/uuid"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	reqBody := requestBody{}
	err := decoder.Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if reqBody.Email == "" || reqBody.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating user")
		return
	}

	// Create user with email and hashed password
	user, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Email:          reqBody.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("Error creating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Map database user to response user (excluding password)
	responseUser := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	respondWithJSON(w, http.StatusCreated, responseUser)
}

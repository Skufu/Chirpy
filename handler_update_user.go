package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/auth"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/database"
)

type updateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateUserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	//get Bearer token
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Invalid JWT: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT")
		return
	}

	// validate jwt and extract userID
	userID, err := auth.ValidateJWT(bearerToken, cfg.jwtSecret)
	if err != nil {
		log.Printf("Invalid JWT: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT")
		return
	}

	//Parse request body
	var req updateUserRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// validate fields
	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// get user from database
	user, err := cfg.db.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	//hashpassword
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// update user in database
	updatedUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             user.ID,
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("Error updating user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// respond with updated user
	respondWithJSON(w, http.StatusOK, updateUserResponse{
		ID:        updatedUser.ID.String(),
		Email:     updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	})

}

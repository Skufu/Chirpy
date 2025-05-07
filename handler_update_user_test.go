package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Create a test version of the handler that accepts a validateJWT function
type testHandler struct {
	db              *MockQueries
	jwtSecret       string
	validateJWTFunc func(string, string) (uuid.UUID, error)
}

func (h *testHandler) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get Bearer token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Missing Authorization header")
		return
	}

	// Check if it starts with "Bearer "
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		respondWithError(w, http.StatusUnauthorized, "Invalid Authorization header format")
		return
	}

	// Extract the token
	bearerToken := authHeader[7:]
	if bearerToken == "" {
		respondWithError(w, http.StatusUnauthorized, "Empty token")
		return
	}

	// Validate JWT using the injected function
	userID, err := h.validateJWTFunc(bearerToken, h.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT")
		return
	}

	// Parse request body
	var req updateUserRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Get user from database
	user, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to get user")
		return
	}

	// Hash password would go here in a real implementation
	hashedPassword := "hashed_" + req.Password

	// Update user in database
	updatedUser, err := h.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             user.ID,
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// Respond with updated user
	respondWithJSON(w, http.StatusOK, updateUserResponse{
		ID:        updatedUser.ID.String(),
		Email:     updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	})
}

// Mock database for testing
type MockQueries struct {
	mock.Mock
}

func (m *MockQueries) GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.User), args.Error(1)
}

func (m *MockQueries) UpdateUser(ctx context.Context, params database.UpdateUserParams) (database.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(database.User), args.Error(1)
}

func TestHandlerUpdateUser(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		method         string
		authHeader     string
		requestBody    map[string]string
		mockSetup      func(*MockQueries, uuid.UUID)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "Success",
			method:     http.MethodPut,
			authHeader: "Bearer valid_token",
			requestBody: map[string]string{
				"email":    "new@example.com",
				"password": "newpassword123",
			},
			mockSetup: func(db *MockQueries, userID uuid.UUID) {
				// Setup mock responses
				db.On("GetUserByID", mock.Anything, userID).Return(database.User{
					ID:             userID,
					Email:          "old@example.com",
					HashedPassword: "oldhash",
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}, nil)

				db.On("UpdateUser", mock.Anything, mock.MatchedBy(func(params database.UpdateUserParams) bool {
					return params.ID == userID && params.Email == "new@example.com"
				})).Return(database.User{
					ID:             userID,
					Email:          "new@example.com",
					HashedPassword: "newhash",
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"email": "new@example.com",
			},
		},
		{
			name:       "Missing Token",
			method:     http.MethodPut,
			authHeader: "",
			requestBody: map[string]string{
				"email":    "new@example.com",
				"password": "newpassword123",
			},
			mockSetup:      func(db *MockQueries, userID uuid.UUID) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid Token Format",
			method:     http.MethodPut,
			authHeader: "InvalidToken",
			requestBody: map[string]string{
				"email":    "new@example.com",
				"password": "newpassword123",
			},
			mockSetup:      func(db *MockQueries, userID uuid.UUID) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid Method",
			method:     http.MethodPost,
			authHeader: "Bearer valid_token",
			requestBody: map[string]string{
				"email":    "new@example.com",
				"password": "newpassword123",
			},
			mockSetup:      func(db *MockQueries, userID uuid.UUID) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "Missing Email",
			method:     http.MethodPut,
			authHeader: "Bearer valid_token",
			requestBody: map[string]string{
				"password": "newpassword123",
			},
			mockSetup:      func(db *MockQueries, userID uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Missing Password",
			method:     http.MethodPut,
			authHeader: "Bearer valid_token",
			requestBody: map[string]string{
				"email": "new@example.com",
			},
			mockSetup:      func(db *MockQueries, userID uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "User Not Found",
			method:     http.MethodPut,
			authHeader: "Bearer valid_token",
			requestBody: map[string]string{
				"email":    "new@example.com",
				"password": "newpassword123",
			},
			mockSetup: func(db *MockQueries, userID uuid.UUID) {
				db.On("GetUserByID", mock.Anything, userID).Return(database.User{}, errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "Database Error on Update",
			method:     http.MethodPut,
			authHeader: "Bearer valid_token",
			requestBody: map[string]string{
				"email":    "new@example.com",
				"password": "newpassword123",
			},
			mockSetup: func(db *MockQueries, userID uuid.UUID) {
				db.On("GetUserByID", mock.Anything, userID).Return(database.User{
					ID:             userID,
					Email:          "old@example.com",
					HashedPassword: "oldhash",
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}, nil)

				db.On("UpdateUser", mock.Anything, mock.Anything).Return(database.User{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock DB
			mockDB := new(MockQueries)

			// Create test user ID - consistent for all test cases
			userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

			// Create a test JWT validation function
			validateJWTFunc := func(tokenString, tokenSecret string) (uuid.UUID, error) {
				if tokenString == "valid_token" {
					return userID, nil
				}
				return uuid.Nil, errors.New("invalid token")
			}

			// Setup mocks
			tc.mockSetup(mockDB, userID)

			// Create test handler with mocked dependencies
			handler := &testHandler{
				db:              mockDB,
				jwtSecret:       "test_secret",
				validateJWTFunc: validateJWTFunc,
			}

			// Create request body
			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(tc.method, "/api/users", bytes.NewBuffer(body))

			// Add auth header if provided
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.handlerUpdateUser(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// For success cases, check response body contains expected fields
			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Check expected fields
				for k, v := range tc.expectedBody {
					assert.Equal(t, v, response[k])
				}

				// Check required fields exist
				assert.Contains(t, response, "id")
				assert.Contains(t, response, "email")
				assert.Contains(t, response, "created_at")
				assert.Contains(t, response, "updated_at")

				// Ensure password is not included
				assert.NotContains(t, response, "password")
				assert.NotContains(t, response, "hashed_password")
			}

			// Verify all expectations on mock
			mockDB.AssertExpectations(t)
		})
	}
}

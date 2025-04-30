# Chirpy API Documentation

**Date: April 22, 2024**

## Validate Chirp Endpoint

### Overview
We implemented a new endpoint that validates whether a chirp meets the Chirpy platform requirements (specifically that it's 140 characters or less).

### Implementation Details
- **Endpoint:** `POST /api/validate_chirp`
- **Request Format:**
  ```json
  {
    "body": "Text content of the chirp"
  }
  ```
- **Success Response (200 OK):**
  ```json
  {
    "valid": true
  }
  ```
- **Error Response (400 Bad Request):**
  ```json
  {
    "error": "Chirp is too long"
  }
  ```

### Technical Implementation
1. The endpoint is registered in `main.go` using Go 1.22+ pattern style:
   ```go
   mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidateChirp)
   ```

2. The handler function in `validate_chirp.go`:
   - Decodes the JSON request body
   - Validates the chirp length (must be ≤ 140 characters)
   - Returns appropriate response based on validation result

### What I Learned
1. **Go 1.22+ HTTP router patterns**: The new pattern style in Go 1.22+ allows specifying HTTP methods directly in the pattern string.
   ```go
   // Old style (pre-Go 1.22)
   mux.HandleFunc("/api/endpoint", func(w http.ResponseWriter, r *http.Request) {
     if r.Method != http.MethodPost {
       http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
       return
     }
     // Handle POST request...
   })
   
   // New style (Go 1.22+)
   mux.HandleFunc("POST /api/endpoint", handlePostRequest)
   ```

2. **JSON response formatting**: Using `json.Marshal()` automatically handles proper formatting without newlines or extra spaces:
   ```go
   // Outputs: {"valid":true} without pretty formatting
   dat, _ := json.Marshal(validResponse{Valid: true})
   ```

3. **HTTP status codes**: Different status codes should be used for different types of responses:
   - `200 OK`: For successful operations
   - `400 Bad Request`: For client errors (validation failures)
   - `500 Internal Server Error`: For server-side errors

4. **Error handling**: It's important to log errors for debugging purposes before sending error responses.

---

## Update: Profanity Filter
**Date: April 22, 2024, 12:12 PM**

### Overview
Added profanity filtering to the `/api/validate_chirp` endpoint to replace inappropriate words with asterisks.

### Changes Made
1. Modified the endpoint response to return the cleaned chirp body instead of a validity boolean:
   - **New Response Format:**
     ```json
     {
       "cleaned_body": "Filtered text content of the chirp"
     }
     ```

2. Added a list of profane words to filter:
   - kerfuffle
   - sharbert
   - fornax

3. Implemented case-insensitive matching for profane words
   - For example, "KERFUFFLE", "Kerfuffle", and "kerfuffle" are all replaced

4. Excluded words with punctuation from filtering
   - For example, "Sharbert!" is not filtered because it contains punctuation

### Implementation Details
1. Created a helper function `cleanChirp()` to filter profane words:
   ```go
   func cleanChirp(body string) string {
     words := strings.Split(body, " ")
     for i, word := range words {
       wordLower := strings.ToLower(word)
       for _, profane := range profaneWords {
         if wordLower == profane {
           words[i] = "****"
           break
         }
       }
     }
     return strings.Join(words, " ")
   }
   ```

2. Updated the handler response to use the new `cleanedResponse` type:
   ```go
   respondWithJSON(w, http.StatusOK, cleanedResponse{
     CleanedBody: cleanedBody,
   })
   ```

### What I Learned
1. **String manipulation in Go**: Using `strings.Split()` and `strings.Join()` to work with word arrays
2. **Case insensitivity**: Using `strings.ToLower()` for case-insensitive comparison
3. **Modular code design**: Breaking functionality into separate functions for better testing and maintenance 

---

## Database Setup and User Management
**Date: April 22, 2024, 8:18 PM**

### Overview
Implemented database connectivity using PostgreSQL and migrations with Goose. Created a user table and set up SQLC for type-safe database queries.

### Database Schema
1. Created the users table with the following schema:
   ```sql
   CREATE TABLE users (
     id uuid PRIMARY KEY,
     created_at TIMESTAMP NOT NULL,
     updated_at TIMESTAMP NOT NULL,
     email TEXT NOT NULL UNIQUE
   );
   ```

### Migration Setup
1. Set up Goose migrations in `sql/schema` directory
2. Created migration file `001_users.sql` with Up/Down migrations:
   ```sql
   -- +goose Up
   CREATE TABLE users (...)
   
   -- +goose Down
   DROP TABLE users;
   ```
3. Migration commands:
   ```bash
   # Run migrations up
   goose postgres "postgres://user:pass@localhost:5432/chirpy?sslmode=disable" up
   
   # Rollback migrations
   goose postgres "postgres://user:pass@localhost:5432/chirpy?sslmode=disable" down
   ```

### Database Queries with SQLC
1. Created type-safe queries in `sql/queries/users.sql`:
   ```sql
   -- name: CreateUser :one
   INSERT INTO users (id, created_at, updated_at, email)
   VALUES (
       gen_random_uuid(),
       NOW(),
       NOW(),
       $1
   )
   RETURNING *;
   ```
2. Generated Go code with `sqlc generate`

### API Configuration
1. Set up database connection in `main.go`:
   ```go
   dbURL := os.Getenv("DB_URL")
   db, err := sql.Open("postgres", dbURL)
   // ...
   dbQueries := database.New(db)
   ```
2. Added database access to the API configuration:
   ```go
   apiCfg := apiConfig{
     fileserverHits: atomic.Int32{},
     db:             dbQueries,
   }
   ```

### Environment Setup
1. Created `.env` file to store connection string securely:
   ```
   DB_URL="postgres://user:pass@localhost:5432/chirpy?sslmode=disable"
   ```
2. Added `.env` to `.gitignore` for security
3. Used `godotenv` to load environment variables

### Dependencies Added
- github.com/lib/pq - PostgreSQL driver
- github.com/google/uuid - UUID handling
- github.com/joho/godotenv - Environment variable loading from .env file

### What I Learned
1. **Migrations with Goose**: How to create and run database migrations
2. **SQLC**: Generating type-safe Go code from SQL queries
3. **Environment Variables**: Secure storage of connection strings
4. **PostgreSQL Features**: Using UUIDs and timestamps effectively
5. **Go SQL Interface**: Working with database/sql package and drivers 

## Database Connection Fix
**Date: April 23, 2024, 9:33 PM**

### Overview
Fixed issues with the database connection and reset endpoint functionality.

### Issues Identified
1. The database connection was failing with error `role "postgres" does not exist`
2. The reset endpoint was not properly handling database errors

### Changes Made
1. Updated the database connection string in `.env` file to use the local user instead of "postgres":
   ```
   DB_URL="postgres://adriangabriellfrancisco:@localhost:5432/chirpy?sslmode=disable"
   ```

2. Improved error handling in the reset endpoint:
   ```go
   func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
     // Check if in dev environment
     if cfg.platform != "dev" {
       respondWithError(w, http.StatusForbidden, "Forbidden")
       return
     }

     err := cfg.db.DeleteAllUsers(context.Background())
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
   ```

3. Enhanced the reset response format to use consistent JSON:
   - **New Response Format:**
     ```json
     {
       "status": "Reset successful"
     }
     ```

### What I Learned
1. **PostgreSQL User Management**: Properly configuring database connection strings for local development
2. **Error Handling**: Implementing more robust error handling with proper logging
3. **JSON Response Consistency**: Using helper functions like `respondWithJSON` to ensure consistent API responses
4. **Environment Configuration**: The importance of environment-specific configuration for local development 

## Added Chirps Database and API
**Date: May 2, 2024, 3:45 PM**

### Overview
Implemented the chirps table in the database and created API endpoints to save chirps with user association.

### Database Schema
1. Created the chirps table with the following schema:
   ```sql
   CREATE TABLE chirps (
       id uuid PRIMARY KEY,
       created_at TIMESTAMP NOT NULL,
       updated_at TIMESTAMP NOT NULL,
       body TEXT NOT NULL,
       user_id uuid NOT NULL,
       FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
   );
   ```

### Migration Setup
1. Created migration file `002_chirps.sql` with Up/Down migrations:
   ```sql
   -- +goose Up
   CREATE TABLE chirps (
       id uuid PRIMARY KEY,
       created_at TIMESTAMP NOT NULL,
       updated_at TIMESTAMP NOT NULL,
       body TEXT NOT NULL,
       user_id uuid NOT NULL,
       FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
   );
   
   -- +goose Down
   DROP TABLE chirps;
   ```

### Database Queries with SQLC
1. Created type-safe queries in `sql/queries/chirps.sql`:
   ```sql
   -- name: CreateChirp :one
   INSERT INTO chirps (id, created_at, updated_at, body, user_id)
   VALUES (
       gen_random_uuid(),
       NOW(),
       NOW(),
       $1,
       $2
   )
   RETURNING *;
   ```

### Model and Database Implementations
1. Added Chirp struct to `internal/database/models.go`:
   ```go
   type Chirp struct {
       ID        uuid.UUID
       CreatedAt time.Time
       UpdatedAt time.Time
       Body      string
       UserID    uuid.UUID
   }
   ```

2. Added CreateChirpParams struct for parameters:
   ```go
   type CreateChirpParams struct {
       Body   string
       UserID uuid.UUID
   }
   ```

3. Implemented the CreateChirp function in `internal/database/chirps.sql.go`:
   ```go
   func (q *Queries) CreateChirp(ctx context.Context, params CreateChirpParams) (Chirp, error) {
       row := q.db.QueryRowContext(ctx, `
           INSERT INTO chirps (id, created_at, updated_at, body, user_id)
           VALUES (
               gen_random_uuid(),
               NOW(),
               NOW(),
               $1,
               $2
           )
           RETURNING id, created_at, updated_at, body, user_id
       `, params.Body, params.UserID)
       var i Chirp
       err := row.Scan(
           &i.ID,
           &i.CreatedAt,
           &i.UpdatedAt,
           &i.Body,
           &i.UserID,
       )
       return i, err
   }
   ```

4. Added DeleteAllChirps function for reset functionality:
   ```go
   func (q *Queries) DeleteAllChirps(ctx context.Context) error {
       _, err := q.db.ExecContext(ctx, "DELETE FROM chirps")
       return err
   }
   ```

### API Endpoint Implementation
1. Updated the handlerCreateChirp function to save chirps to the database:
   ```go
   func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
       // Parse request parameters
       decoder := json.NewDecoder(r.Body)
       params := parameters{}
       err := decoder.Decode(&params)
       if err != nil {
           respondWithError(w, http.StatusInternalServerError, "Something went wrong")
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

       // Save chirp to database
       userID, err := uuid.Parse(params.UserID)
       if err != nil {
           respondWithError(w, http.StatusBadRequest, "Invalid user ID")
           return
       }

       chirp, err := cfg.db.CreateChirp(context.Background(), database.CreateChirpParams{
           Body:   cleanedBody,
           UserID: userID,
       })
       if err != nil {
           respondWithError(w, http.StatusInternalServerError, "Error creating chirp")
           return
       }

       // Return response with 201 Created status
       respondWithJSON(w, http.StatusCreated, chirpResponse{
           ID:        chirp.ID.String(),
           CreatedAt: chirp.CreatedAt,
           UpdatedAt: chirp.UpdatedAt,
           Body:      chirp.Body,
           UserID:    chirp.UserID.String(),
       })
   }
   ```

2. Updated reset handler to delete chirps before users due to foreign key constraints:
   ```go
   func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
       // Check if in dev environment
       if cfg.platform != "dev" {
           respondWithError(w, http.StatusForbidden, "Forbidden")
           return
       }

       // Delete chirps first due to the foreign key constraint
       err := cfg.db.DeleteAllChirps(context.Background())
       if err != nil {
           log.Printf("Failed to reset chirps: %v", err)
           respondWithError(w, http.StatusInternalServerError, "Failed to reset chirps")
           return
       }

       // Delete users after chirps
       err = cfg.db.DeleteAllUsers(context.Background())
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
   ```

### What I Learned
1. **Foreign Key Constraints**: How to set up relationships between tables with ON DELETE CASCADE
2. **Database Transaction Order**: When resetting tables with foreign keys, you must delete child records before parent records
3. **HTTP Status Codes**: Using 201 Created for successful resource creation
4. **UUID Parsing**: Converting string UUIDs to the UUID type using the uuid.Parse function
5. **Profanity Filtering**: Implementing content moderation before storing user-generated content 

## Single Chirp Retrieval Endpoint
**Date: April 30, 2025, 7:18 PM**

### Overview
Implemented a new endpoint that allows users to retrieve a single chirp by its unique ID. This functionality is essential for viewing individual chirps directly, especially as the number of chirps in the system grows.

### Implementation Details

#### 1. Database Query
First, I added a new SQL query in `sql/queries/chirps.sql` to fetch a single chirp by its ID:

```sql
-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1;
```

This SQL query retrieves all columns from the `chirps` table where the ID matches the provided parameter.

#### 2. Code Generation
I ran the `sqlc generate` command to create the Go implementation of this query. This automatically generated a `GetChirp` function in the `database` package that accepts a UUID and returns a single chirp:

```go
func (q *Queries) GetChirp(ctx context.Context, id uuid.UUID) (Chirp, error) {
    row := q.db.QueryRowContext(ctx, getChirp, id)
    var i Chirp
    err := row.Scan(
        &i.ID,
        &i.CreatedAt,
        &i.UpdatedAt,
        &i.Body,
        &i.UserID,
    )
    return i, err
}
```

#### 3. Handler Implementation
Created a new handler function `handlerGetChirp` in `handler_get_chirp.go`:

```go
func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
    // Get chirp ID from path parameter
    chirpIDStr := r.PathValue("chirpID")
    
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
```

This handler:
- Uses `r.PathValue("chirpID")` to extract the chirp ID from the URL path parameter
- Parses the string ID into a UUID using `uuid.Parse()`
- Queries the database for the chirp with the matching ID
- Handles different error scenarios:
  - Returns 400 Bad Request for invalid UUID format
  - Returns 404 Not Found if no chirp exists with that ID
  - Returns 500 Internal Server Error for database errors
- Returns the chirp with a 200 OK status if found

#### 4. Route Registration
Added the new endpoint to `main.go` using Go 1.22+ pattern matching syntax:

```go
mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)
```

This registers the handler for GET requests to paths matching the pattern `/api/chirps/{chirpID}`, where `{chirpID}` is a path parameter.

### Testing
Tested the endpoint by:
1. Creating a user:
   ```bash
   curl -X POST http://localhost:8080/api/users \
     -H "Content-Type: application/json" \
     -d '{"email": "saul@bettercall.com"}'
   ```

2. Creating a chirp from that user:
   ```bash
   curl -X POST http://localhost:8080/api/chirps \
     -H "Content-Type: application/json" \
     -d '{"body": "I'm gonna be a damn good developer, and people are gonna know about it.", "user_id": "4d0e0ab7-ca28-4626-acbf-73115900e8fd"}'
   ```

3. Retrieving the chirp by its ID:
   ```bash
   curl -X GET "http://localhost:8080/api/chirps/91c19d70-286e-4924-b399-da1dd0fb5596"
   ```

### Response Format
For a successful request, the endpoint returns a JSON object with the following structure:
```json
{
  "id": "91c19d70-286e-4924-b399-da1dd0fb5596",
  "created_at": "2025-04-30T19:11:54.202171Z",
  "updated_at": "2025-04-30T19:11:54.202171Z",
  "body": "I'm gonna be a damn good developer, and people are gonna know about it.",
  "user_id": "4d0e0ab7-ca28-4626-acbf-73115900e8fd"
}
```

### What I Learned
1. **Path Parameters in Go 1.22+**: Using the new `r.PathValue()` method to extract path parameters cleanly
2. **UUID Handling**: Converting string IDs to UUIDs using the `uuid.Parse()` function
3. **HTTP Status Codes**: Using appropriate status codes for different scenarios:
   - 200 OK for successful responses
   - 400 Bad Request for client errors (invalid ID format)
   - 404 Not Found for resources that don't exist
   - 500 Internal Server Error for server-side errors
4. **Error Handling**: Distinguishing between different types of errors (e.g., `sql.ErrNoRows` vs. other errors)
5. **JSON Response Formatting**: Building structured JSON responses for RESTful APIs

### Next Steps
This endpoint provides the foundation for several future features:
- Direct links to individual chirps
- Chirp details pages in the frontend
- Chirp sharing functionality
- Comment systems that reference specific chirps

### Tech Stack
- Go 1.23.2 with net/http package
- PostgreSQL database
- SQLC for type-safe database queries
- UUID for unique identifiers 

## Password Hashing Implementation
**Date: April 30, 2025, 7:39 PM**

### Overview
Added password hashing capabilities to the Chirpy application by implementing a database migration to add a `hashed_password` column to the users table and creating a new `auth` package for securely handling passwords.

### Database Changes
1. Created a new migration file `003_users_add_password.sql` to add the `hashed_password` column to the users table:
   ```sql
   -- +goose Up
   ALTER TABLE users 
   ADD COLUMN hashed_password TEXT NOT NULL DEFAULT 'unset';
   
   -- +goose Down
   ALTER TABLE users
   DROP COLUMN hashed_password;
   ```

2. The migration adds a non-null TEXT column with a default value of 'unset', which ensures that existing users in the database don't cause constraint violations.

3. Applied the migration using direct SQL since there were issues with the migration tool and the tables already existed:
   ```bash
   psql "postgres://user:password@localhost:5432/chirpy?sslmode=disable" -c "ALTER TABLE users ADD COLUMN hashed_password TEXT NOT NULL DEFAULT 'unset';"
   ```

### Auth Package Implementation
1. Created a new `internal/auth` package with password hashing functionality:
   ```go
   package auth
   
   import (
       "golang.org/x/crypto/bcrypt"
   )
   
   // HashPassword takes a plain text password and returns a bcrypt hash
   func HashPassword(password string) (string, error) {
       // Generate a bcrypt hash using the default cost (10)
       bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
       if err != nil {
           return "", err
       }
       
       // Return the hash as a string
       return string(bytes), nil
   }
   
   // CheckPasswordHash compares a bcrypt hashed password with a plain text password
   // Returns nil if the passwords match, or an error if they don't
   func CheckPasswordHash(hash, password string) error {
       // Compare the stored hash with the provided password
       return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
   }
   ```

2. Added comprehensive tests in `internal/auth/password_test.go` to verify both functions:
   - `TestHashPassword`: Ensures hashing produces non-empty results that differ from the input
   - `TestCheckPasswordHash`: Verifies that correct passwords validate and incorrect ones fail

### Dependencies Added
- `golang.org/x/crypto/bcrypt` - Industry-standard password hashing library that includes:
  - Salt generation to prevent rainbow table attacks
  - Work factor adjustments to resist brute force attacks
  - Secure comparison to prevent timing attacks

### Security Considerations
1. **Password Storage**: We're following security best practices by:
   - Never storing passwords in plain text
   - Using bcrypt, a purpose-built password hashing algorithm
   - Using the default cost factor (10) which provides a good balance between security and performance

2. **Default Value**: Setting a default value of 'unset' ensures that:
   - Existing users can be identified as needing to set a password
   - The application can check for this value and prompt users to update their passwords

### What I Learned
1. **Secure Password Handling**: 
   - The importance of using specialized algorithms like bcrypt for password storage
   - How bcrypt automatically handles salt generation and secure comparison

2. **Go Cryptography Libraries**:
   - How to use the `golang.org/x/crypto/bcrypt` package
   - The difference between encryption (reversible) and hashing (one-way)

3. **Database Migration Patterns**:
   - How to add non-null columns to existing tables using DEFAULT values
   - Troubleshooting migration issues when tables already exist

4. **Package Organization**:
   - Creating a separate `auth` package to isolate security-related functionality
   - Writing comprehensive tests for security-critical code

### Next Steps
This implementation provides the foundation for several future security features:
1. User registration with password protection
2. Secure login functionality
3. Password change capabilities
4. Account recovery workflows

These changes will allow Chirpy to implement proper user authentication in future updates. 

## User Authentication Implementation
**Date: April 30, 2025, 8:08 PM**

### Overview
Enhanced the Chirpy API with user authentication capabilities by implementing password handling in the user creation endpoint and adding a new login endpoint for credential verification.

### API Changes

#### Updated User Creation Endpoint
1. Modified the `POST /api/users` endpoint to accept and hash passwords:
   ```json
   // Request
   {
     "email": "user@example.com",
     "password": "secure-password"
   }
   
   // Response - 201 Created
   {
     "id": "f0f87ec2-a8b5-48cc-b66a-a85ce7c7b862",
     "created_at": "2021-07-07T00:00:00Z",
     "updated_at": "2021-07-07T00:00:00Z",
     "email": "user@example.com"
   }
   ```

2. Security improvements:
   - Passwords are now required for user creation
   - Passwords are hashed using bcrypt before storage
   - Hashed passwords are never included in API responses

#### New Login Endpoint
1. Added a new `POST /api/login` endpoint that validates user credentials:
   ```json
   // Request
   {
     "email": "user@example.com",
     "password": "secure-password"
   }
   
   // Response - 200 OK (successful login)
   {
     "id": "f0f87ec2-a8b5-48cc-b66a-a85ce7c7b862",
     "created_at": "2021-07-07T00:00:00Z",
     "updated_at": "2021-07-07T00:00:00Z",
     "email": "user@example.com"
   }
   
   // Response - 401 Unauthorized (failed login)
   {
     "error": "Incorrect email or password"
   }
   ```

2. Authentication flow:
   - Looks up the user by email
   - Compares the provided password with the stored hash
   - Returns the user object on success (without password)
   - Returns a generic error on failure (doesn't specify whether email or password was incorrect)

### Database Changes
1. Added a `GetUserByEmail` query to support login functionality:
   ```sql
   -- name: GetUserByEmail :one
   SELECT * FROM users
   WHERE email = $1;
   ```

2. Updated the `CreateUser` query to store the hashed password:
   ```sql
   -- name: CreateUser :one
   INSERT INTO users (id, created_at, updated_at, email, hashed_password)
   VALUES (
       gen_random_uuid(),
       NOW(),
       NOW(),
       $1,
       $2
   )
   RETURNING *;
   ```

### Security Considerations
1. **Password Management**:
   - Raw passwords are only handled in memory and never stored
   - Bcrypt is used with its default work factor (10) for optimal security/performance balance
   - Password comparison is done using constant-time comparison to prevent timing attacks

2. **Login Endpoint Security**:
   - Generic error messages prevent user enumeration attacks
   - Failed login attempts are logged for security monitoring
   - Proper HTTP status codes (401 for auth failure, 400 for bad requests)
   - No session tokens yet (to be added in a future update)

3. **HTTPS Requirement**:
   - The API assumes HTTPS in production to protect sensitive data in transit
   - Raw passwords can be safely transmitted over HTTPS

### Implementation Details
1. Created a new `handlerLogin` function to handle login requests:
   ```go
   func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
       // Parse login credentials from request body
       // Look up user by email
       // Compare password hash
       // Return user data on success or error on failure
   }
   ```

2. Updated `handlerCreateUser` to handle password hashing:
   ```go
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
   ```

### What I Learned
1. **Security Best Practices**:
   - How to design authentication endpoints with security in mind
   - The importance of not revealing sensitive information in error messages
   - Using appropriate HTTP status codes for different authentication scenarios

2. **Password Handling in Go**:
   - How to securely hash and validate passwords using bcrypt
   - Dealing with the performance implications of secure password hashing
   - Planning for future security enhancements like rate limiting

3. **API Design**:
   - Keeping API responses clean and focused on the necessary data
   - Balancing security with usability in authentication endpoints
   - Consistent error handling patterns across endpoints

### Next Steps
This implementation provides the foundation for several future security features:
1. JWT token implementation for authenticated requests
2. Password reset functionality
3. Account locking after failed login attempts
4. Two-factor authentication options

These changes will allow Chirpy to implement proper user authentication in future updates. 
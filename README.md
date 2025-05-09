# Chirpy Server

A Go HTTP server serving static files, providing a Chirp API, and tracking usage metrics.

## Overview

Chirpy Server is a backend application built with Go. It serves static web content, offers a comprehensive API for managing "chirps" (short messages) and users, handles user authentication via JWTs, and includes administrative features like metrics tracking and database reset for development.

## Features

*   **Static File Serving:** Serves static files (HTML, CSS, JS) from the project's root directory via the `/app/*` endpoint.
*   **Comprehensive Chirp API:**
    *   **User Management:** Create new users (`POST /api/users`), and update existing authenticated users' information (`PUT /api/users`).
    *   **Chirp Management:** Create new chirps (`POST /api/chirps`), list all chirps (`GET /api/chirps`), retrieve a single chirp by its ID (`GET /api/chirps/{chirpID}`), and delete chirps (`DELETE /api/chirps/{chirpID}`). Deletion requires the authenticated user to be the author of the chirp.
*   **User Authentication:** Secure, token-based authentication using JSON Web Tokens (JWT). Includes endpoints for user login (`POST /api/login`), access token refreshing (`POST /api/refresh`), and refresh token revocation (`POST /api/revoke`).
*   **"Chirpy Red" Status:** A field `is_chirpy_red` in the `users` table indicates a premium user status.
*   **Webhook Integration Example:** Includes a mock webhook endpoint (`POST /api/polka/webhooks`) for demonstration purposes, simulating integration with an external service like "Polka." This endpoint is secured using an API key (`POLKA_KEY`).
*   **Health Check & Metrics:** Provides a health check endpoint (`GET /api/healthz`) to confirm the server is running and an admin endpoint (`GET /admin/metrics`) to view request counts for the static file server.
*   **Database Management:** Utilizes `goose` for managing database schema migrations, which are applied automatically on server startup.
*   **Development Utilities:** Offers a database reset functionality (`POST /admin/reset`) for development environments (when `PLATFORM=dev`), which clears the database and resets metrics.

## Technologies Used

*   **Backend:** Go (utilizing the standard `net/http` library).
*   **Database:** PostgreSQL.
*   **Database Tools:**
    *   **`sqlc`:** Used for generating type-safe Go code from SQL queries, facilitating database interactions.
    *   **`goose`:** Employed for managing database schema migrations.
*   **Authentication:** JSON Web Tokens (JWT) for secure API authentication.
*   **Configuration:** Primarily through environment variables, typically managed with a `.env` file.

## API Endpoints

| Endpoint                     | Method | Description                                                                 | Authentication Required        |
| ---------------------------- | ------ | --------------------------------------------------------------------------- | ------------------------------ |
| `GET /api/healthz`           | GET    | Health check endpoint; returns "OK" when the server is running.             | None                           |
| `POST /api/users`            | POST   | Creates a new user account.                                                 | None                           |
| `PUT /api/users`             | PUT    | Updates an authenticated user's email and password.                         | JWT Bearer Token               |
| `POST /api/login`            | POST   | Logs in a user, returning an access token and a refresh token.              | Basic Auth (email & password)  |
| `POST /api/refresh`          | POST   | Issues a new access token using a valid refresh token.                      | Refresh Token (as Bearer)      |
| `POST /api/revoke`           | POST   | Revokes a given refresh token, logging the user out of that session.        | Refresh Token (as Bearer)      |
| `POST /api/chirps`           | POST   | Creates a new chirp associated with the authenticated user.                 | JWT Bearer Token               |
| `GET /api/chirps`            | GET    | Lists all chirps.                                                           | None                           |
| `GET /api/chirps/{chirpID}`  | GET    | Retrieves a specific chirp by its ID.                                       | None                           |
| `DELETE /api/chirps/{chirpID}`| DELETE | Deletes a specific chirp.                                                   | JWT Bearer Token (author only) |
| `POST /api/polka/webhooks`   | POST   | Mock webhook endpoint for "Polka" service events.                           | `Authorization: ApiKey YOUR_POLKA_KEY` |
| `GET /admin/metrics`         | GET    | Displays the number of hits to the static file server in HTML format.       | None                           |
| `POST /admin/reset`          | POST   | Resets hit counter and clears database (only if `PLATFORM=dev`).            | None                           |
| `/app/*`                     | GET    | Serves static files from the root directory.                                | None                           |

## Database Schema

The database schema is managed by `goose` migration files located in the `sql/schema/` directory. The primary tables are:

*   **`users`**:
    *   `id` (UUID, Primary Key)
    *   `created_at` (TIMESTAMP)
    *   `updated_at` (TIMESTAMP)
    *   `email` (TEXT, Unique)
    *   `hashed_password` (TEXT)
    *   `is_chirpy_red` (BOOLEAN, default: `FALSE`) - Indicates premium status.
*   **`chirps`**:
    *   `id` (UUID, Primary Key)
    *   `created_at` (TIMESTAMP)
    *   `updated_at` (TIMESTAMP)
    *   `body` (TEXT) - Content of the chirp (1-280 characters).
    *   `user_id` (UUID, Foreign Key to `users.id` on delete cascade)
*   **`refresh_tokens`**:
    *   `token` (TEXT, Primary Key) - The refresh token string.
    *   `created_at` (TIMESTAMP)
    *   `updated_at` (TIMESTAMP)
    *   `user_id` (UUID, Foreign Key to `users.id` on delete cascade)
    *   `expires_at` (TIMESTAMP) - When the refresh token becomes invalid.
    *   `revoked_at` (TIMESTAMP, nullable) - When the token was explicitly revoked.

## Setup and Installation

### Prerequisites

*   **Go:** Version 1.21 or later recommended.
*   **PostgreSQL:** A running instance of PostgreSQL.
*   **`goose` (Optional):** For manual database migration management. Install via:
    ```bash
    go install github.com/pressly/goose/v3/cmd/goose@latest
    ```

### Steps

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/yourusername/chirpy.git # Replace with your repo URL
    cd chirpy
    ```

2.  **Database Setup:**
    *   Ensure your PostgreSQL server is running and accessible.
    *   Create a database for Chirpy (e.g., `chirpy_db`).
    *   The application is configured to automatically run database migrations on startup using the connection string from your `.env` file.
    *   **(Optional) Manual Migration:** If you need to run migrations manually or inspect the schema:
        ```bash
        # Ensure goose is installed and you are in the project root
        goose -dir sql/schema postgres "postgresql://USER:PASSWORD@HOST:PORT/DBNAME?sslmode=disable" up
        ```
        Replace `USER`, `PASSWORD`, `HOST`, `PORT`, and `DBNAME` with your PostgreSQL details.

3.  **Configuration:**
    *   Create a `.env` file in the project root. You can copy `.env.example` if provided, or create it from scratch:
        ```env
        DB_URL=postgresql://user:password@localhost:5432/chirpy_db?sslmode=disable
        JWT_SECRET=a_very_secure_secret_key_for_jwt_hs256
        POLKA_KEY=your_polka_service_api_key
        PLATFORM=dev # Use 'dev' for development (enables /admin/reset), 'prod' for production
        ```
    *   Update `DB_URL` with your actual PostgreSQL connection string.
    *   Set a strong, unique `JWT_SECRET`.
    *   If you use the Polka webhook, set your `POLKA_KEY`.

4.  **Build (Optional):**
    To create a standalone executable:
    ```bash
    go build -o chirpy_server
    ```

5.  **Run the Server:**
    *   Using `go run` (for development):
        ```bash
        go run .
        ```
    *   Or, if you built an executable:
        ```bash
        ./chirpy_server
        ```
        (On Windows: `chirpy_server.exe`)

6.  The server will start by default on `http://localhost:8080`.

## Authentication

Chirpy uses JSON Web Tokens (JWTs) for API authentication, involving Access Tokens and Refresh Tokens.

*   **Login (`POST /api/login`):** Users authenticate by sending their email and password (using HTTP Basic Authentication). Upon successful authentication, the server returns:
    *   An **Access Token:** A short-lived JWT used to authenticate subsequent API requests.
    *   A **Refresh Token:** A long-lived token used to obtain a new access token when the current one expires. This token is stored in the database.
*   **Access Token Usage:** For protected endpoints, the client must send the Access Token in the `Authorization` header with the `Bearer` scheme:
    ```
    Authorization: Bearer <your_access_token>
    ```
*   **Refreshing Access Tokens (`POST /api/refresh`):** When an access token expires, clients can request a new one by sending their valid (non-revoked, non-expired) Refresh Token in the `Authorization` header (also as a Bearer token) to the `/api/refresh` endpoint.
*   **Revoking Refresh Tokens (`POST /api/revoke`):** Clients can invalidate their current Refresh Token (e.g., on logout) by sending it to the `/api/revoke` endpoint. This prevents it from being used to obtain new access tokens.
*   **JWT Secret:** The integrity of JWTs is ensured by a secret key, configured via the `JWT_SECRET` environment variable. This key must be kept confidential.

## Webhook Endpoint (`POST /api/polka/webhooks`)

This server includes a mock webhook endpoint at `POST /api/polka/webhooks`. This endpoint is intended as an example of how the server might receive notifications from an external service (referred to as "Polka" in this example).

*   **Purpose:** To simulate processing events from an external system. Currently, it handles `user.upgraded` events to update a user's `is_chirpy_red` status.
*   **Authentication:** Requests to this endpoint must be authenticated by providing the `POLKA_KEY` (as defined in your `.env` file) in the `Authorization` header using the `ApiKey` scheme:
    ```
    Authorization: ApiKey YOUR_POLKA_KEY_VALUE
    ```

## Project Structure

A brief overview of the key directories and files:

*   `main.go`: The main application entry point, responsible for server setup, routing, and initializing dependencies.
*   `internal/auth/`: Contains logic for JWT generation, validation, password hashing, and API key handling.
*   `internal/database/`: Houses the Go code generated by `sqlc` for type-safe database interactions. **Do not edit files in this directory manually.**
*   `sql/schema/`: Contains database migration files (e.g., `001_users.sql`, `002_chirps.sql`) written in SQL and managed by `goose`.
*   `sql/queries/`: Stores the SQL queries (e.g., `users.sql`, `chirps.sql`) that `sqlc` uses to generate the Go code in `internal/database/`.
*   `database_migrate.go`: Includes code to automatically apply pending database migrations when the server starts.
*   `handler_*.go`: A collection of files, each typically containing the HTTP request handler logic for one or more related API endpoints (e.g., `handler_users_create.go`, `handler_chirps_get.go`).
*   `middleware/`: Contains HTTP middleware functions (e.g., for logging, authentication checks, metrics).
*   `go.mod`, `go.sum`: Go module files defining project dependencies.
*   `sqlc.yaml`: The configuration file for the `sqlc` tool.
*   `.gitignore`: Specifies intentionally untracked files that Git should ignore.
*   `README.md`: This file.
*   `api_usage_examples.md`: Contains `curl` examples for interacting with the API.

## API Usage Examples

For detailed examples of how to interact with the API using `curl`, including request formats and expected responses, please refer to the [api_usage_examples.md](api_usage_examples.md) file.

## Development

*   **`PLATFORM` Environment Variable:**
    *   Set `PLATFORM=dev` in your `.env` file during development. This enables the `POST /admin/reset` endpoint, which is useful for clearing the database and metrics.
    *   Set `PLATFORM=prod` (or any other value) for production to disable the reset endpoint.
*   **Database Tools:**
    If you plan to modify database queries or the schema, you'll need `sqlc` and `goose`:
    *   Install `sqlc`: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
    *   Install `goose`: `go install github.com/pressly/goose/v3/cmd/goose@latest`
*   **Workflow:**
    1.  **Modifying Queries:** Edit `.sql` files in `sql/queries/`. Then run `sqlc generate` from the project root.
    2.  **Modifying Schema:** Create a new migration file in `sql/schema/` (e.g., `00X_descriptive_name.sql`) with `Up` and `Down` sections. Apply migrations using `goose` or rely on automatic migration on startup.

## Error Handling

The API uses standard HTTP status codes to indicate the success or failure of requests. Common codes include:

*   `200 OK`: The request was successful.
*   `201 Created`: The resource was successfully created (e.g., a new user or chirp).
*   `204 No Content`: The request was successful, but there is no response body (e.g., successful deletion).
*   `400 Bad Request`: The request was malformed or contained invalid parameters (e.g., missing required fields, invalid chirp length).
*   `401 Unauthorized`: Authentication failed or was not provided (e.g., missing, invalid, or expired JWT; incorrect Basic Auth credentials for login).
*   `403 Forbidden`: The authenticated user does not have permission to perform the requested action (e.g., attempting to delete another user's chirp).
*   `404 Not Found`: The requested resource could not be found (e.g., chirp with a given ID does not exist).
*   `500 Internal Server Error`: An unexpected error occurred on the server.

---

This README provides a comprehensive guide to the Chirpy server. For specific API interaction examples, please see [api_usage_examples.md](api_usage_examples.md). 
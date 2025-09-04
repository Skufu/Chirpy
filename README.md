# Chirpy

A Go HTTP server serving static files, providing a Chirp API, and tracking usage metrics.

# Chirpy Server
A Go HTTP server serving static files, providing a Chirp API, and tracking usage metrics.
## Overview
Chirpy Server is a backend application built with Go. It serves static web content, offers a comprehensive API for managing "chirps" (short messages) and users, handles user authentication via JWTs, and includes administrative features like metrics tracking and database reset for development.
## Features
* **Static File Serving:** Serves static files (HTML, CSS, JS) from the project's root directory via the `/app/*` endpoint.
* **Comprehensive Chirp API:**
  * **User Management:** Create new users (`POST /api/users`), and update existing authenticated users' information (`PUT /api/users`).
  * **Chirp Management:** Create new chirps (`POST /api/chirps`), list all chirps (`GET /api/chirps`), retrieve a single chirp by its ID (`GET /api/chirps/{chirpID}`), and delete chirps (`DELETE /api/chirps/{chirpID}`). Deletion requires the authenticated user to be the author of the chirp.
* **User Authentication:** Secure, token-based authentication using JSON Web Tokens (JWT). Includes endpoints for user login (`POST /api/login`), access token refreshing (`POST /api/refresh`), and refresh token revocation (`POST /api/revoke`).
* **"Chirpy Red" Status:** A field `is_chirpy_red` in the `users` table indicates a premium user status.
* **Webhook Integration Example:** Includes a mock webhook endpoint (`POST /api/polka/webhooks`) for demonstration purposes, simulating integration with an external service like "Polka." This endpoint is secured using an API key (`POLKA_KEY`).
* **Health Check & Metrics:** Provides a health check endpoint (`GET /api/healthz`) to confirm the server is running and an admin endpoint (`GET /admin/metrics`) to view request counts for the static file server.
* **Database Management:** Utilizes `goose` for managing database schema migrations, which are applied automatically on server startup.
* **Development Utilities:** Offers a database reset functionality (`POST /admin/reset`) for development environments (when `PLATFORM=dev`), which clears the database and resets metrics.
## Technologies Used
* **Backend:** Go (utilizing the standard `net/http` library).
* **Database:** PostgreSQL.
* **Database Tools:**
  * **`sqlc`:** Used for generating type-safe Go code from SQL queries, facilitating database interactions.
  * **`goose`:** Employed for managing database schema migrations.
* **Authentication:** JSON Web Tokens (JWT) for secure API authentication.
* **Configuration:** Primarily through environment variables, typically managed with a `.env` file.
## API Endpoints
| Endpoint                     | Method | Description                                                                 | Authentication Required        |
| ---------------------------- | ------ | --------------------------------------------------------------------------- | ------------------------------ |
| `GET /api/healthz`           | GET    | Health check endpoint; returns "OK" when the server is running.             | None                           |
| `POST /api/users`            | POST   | Creates a new user account.                                                 | None                           |

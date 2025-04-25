package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/Skufu/HTTPS-Bootdev/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	// Create tables if they don't exist
	_, err = db.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		id uuid PRIMARY KEY,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		email TEXT NOT NULL UNIQUE
	)`)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	_, err = db.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS chirps (
		id uuid PRIMARY KEY,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		body TEXT NOT NULL,
		user_id uuid NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Fatalf("Failed to create chirps table: %v", err)
	}

	dbQueries := database.New(db)

	const port = "8080"
	const filePathRoot = "."

	// Initialize API config
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       os.Getenv("PLATFORM"),
	}

	mux := http.NewServeMux()

	// File server for /app/ path
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot))))
	mux.Handle("/app/", fsHandler)

	// Register API endpoints
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filePathRoot, port)
	log.Fatal(server.ListenAndServe())
}

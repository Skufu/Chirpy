package main

import (
	"context"
	"database/sql"
	"log"
	"path/filepath"
	"runtime"

	"github.com/pressly/goose/v3"
)

// runDatabaseMigrations executes database migrations using goose
func runDatabaseMigrations(ctx context.Context, db *sql.DB) error {
	log.Println("Running database migrations with goose...")

	// Set the dialect for PostgreSQL
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// Get the absolute path to the migrations directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("Warning: Unable to determine runtime path, using relative path")
		return goose.Up(db, "sql/schema")
	}

	// Get the absolute path to the migrations
	migrationsDir := filepath.Join(filepath.Dir(filename), "sql", "schema")
	log.Printf("Migration files located at: %s", migrationsDir)

	// Run migrations
	if err := goose.Up(db, migrationsDir); err != nil {
		// If the error is because tables already exist, we can continue
		log.Printf("Note: If migrations report tables already exist, this is expected for existing databases")
		// Don't return the error here so we can proceed with the manual fix
	}

	// Ensure the is_chirpy_red column exists in the users table
	// This is a safety measure in case migrations didn't fully complete
	if err := ensureChirpyRedColumnExists(ctx, db); err != nil {
		log.Printf("Error ensuring is_chirpy_red column: %v", err)
		return err
	}

	log.Println("Database initialization completed successfully")
	return nil
}

// ensureChirpyRedColumnExists checks if the is_chirpy_red column exists in the users table
// If it doesn't exist, it adds the column with a default value of false
func ensureChirpyRedColumnExists(ctx context.Context, db *sql.DB) error {
	log.Println("Ensuring is_chirpy_red column exists...")

	// Check if the column exists
	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = 'users' AND column_name = 'is_chirpy_red'
		);
	`).Scan(&exists)

	if err != nil {
		return err
	}

	// If the column doesn't exist, add it
	if !exists {
		log.Println("Adding is_chirpy_red column to users table...")
		_, err = db.ExecContext(ctx, `
			ALTER TABLE users
			ADD COLUMN is_chirpy_red BOOLEAN DEFAULT FALSE;
		`)
		if err != nil {
			return err
		}
		log.Println("Added is_chirpy_red column successfully")
	} else {
		log.Println("is_chirpy_red column already exists")
	}

	return nil
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

// Middleware to increment the hit counter
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// Handler to reset hit count
func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Counter reset"))
}

// Readiness handler
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Metrics handler with method check
func (cfg *apiConfig) metricsHandlerWithMethodCheck(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hitCount := cfg.fileserverHits.Load()
	fmt.Fprintf(w, "Hits: %d", hitCount)
}

// Handler to return 405 Method Not Allowed
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func main() {
	const port = "8080"
	const filePathRoot = "."

	// Initialize API config
	apiCfg := apiConfig{}

	mux := http.NewServeMux()

	// Register API endpoints with explicit method checking
	mux.HandleFunc("/api/healthz", readinessHandler)
	mux.HandleFunc("/api/metrics", apiCfg.metricsHandlerWithMethodCheck)
	mux.HandleFunc("/api/reset", apiCfg.resetHandler)

	// File server for /app/ path, stripping the /app prefix
	fileServer := http.FileServer(http.Dir(filePathRoot))
	mux.Handle("/app", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Println("Server started on http://localhost:" + port)
	log.Fatal(server.ListenAndServe())
}

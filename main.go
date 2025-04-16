package main

import (
	"log"
	"net/http"
)

// Readiness handler
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	const port = "8080"
	const filePathRoot = "."

	mux := http.NewServeMux()

	// Register readiness endpoint
	mux.HandleFunc("/healthz", readinessHandler)

	// File server for /app/ path, stripping the /app prefix
	fileServer := http.FileServer(http.Dir(filePathRoot))
	mux.Handle("/app/", http.StripPrefix("/app", fileServer))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Println("Server started on http://localhost:" + port)
	log.Fatal(server.ListenAndServe())
}

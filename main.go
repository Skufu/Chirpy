package main

import (
	"log"
	"net/http"
)

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Create a file server handler using the current directory
	fileServer := http.FileServer(http.Dir("."))

	// Register the file server handler for the root path
	mux.Handle("/", fileServer)

	// Create a new http.Server struct
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server
	log.Println("Server started on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

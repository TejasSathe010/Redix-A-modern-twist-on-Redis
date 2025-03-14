package main

import (
	"log"
)

func main() {
	// Initialize storage
	store := storage.NewInMemoryStore()

	// Create command handler
	handler := storage.NewCommandHandler(store)

	// Create network server
	server := network.NewServer(":6379", handler)

	// Start server
	log.Println("Starting server on :6379")
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

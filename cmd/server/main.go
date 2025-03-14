package main

import (
	"log"

	"github.com/TejasSathe010/Redix-A-modern-twist-on-Redis/internal/network"
	"github.com/TejasSathe010/Redix-A-modern-twist-on-Redis/storage"
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

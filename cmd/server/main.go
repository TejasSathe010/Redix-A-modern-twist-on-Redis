package main

import (
	"log"

	"github.com/TejasSathe010/Redix-A-modern-twist-on-Redis/internal/network"
	"github.com/TejasSathe010/Redix-A-modern-twist-on-Redis/internal/storage"
)

func main() {
	store := storage.NewInMemoryStore()

	handler := storage.NewCommandHandler(store)

	server := network.NewServer(":6379", handler)

	log.Println("Starting server on :6379")
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/TejasSathe010/Redix-A-modern-twist-on-Redis/internal/network"
	"github.com/TejasSathe010/Redix-A-modern-twist-on-Redis/internal/storage"
)

type Config struct {
	Server struct {
		Addr         string   `json:"addr"`
		Mode         string   `json:"mode"`
		ClusterNodes []string `json:"cluster_nodes"`
	} `json:"server"`
	Storage struct {
		Dir string `json:"dir"`
	} `json:"storage"`
}

func main() {
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("Configuration file path must be provided")
	}

	configData, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to read configuration file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	store := storage.NewInMemoryStore()

	handler := storage.NewCommandHandler(store)

	server := network.NewServer(config.Server.Addr, handler)

	log.Printf("Starting server on %s", config.Server.Addr)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

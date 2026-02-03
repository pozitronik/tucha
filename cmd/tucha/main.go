// Package main is the entry point for the Tucha server.
package main

import (
	"flag"
	"log"
	"net/http"

	"tucha/internal/api"
	"tucha/internal/auth"
	"tucha/internal/config"
	"tucha/internal/content"
	"tucha/internal/storage"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Tucha server starting")
	log.Printf("  Listen: %s", cfg.Addr())
	log.Printf("  External URL: %s", cfg.Server.ExternalURL)
	log.Printf("  User: %s", cfg.User.Email)
	log.Printf("  Database: %s", cfg.Storage.DBPath)
	log.Printf("  Content dir: %s", cfg.Storage.ContentDir)
	log.Printf("  Quota: %d bytes", cfg.Storage.QuotaBytes)

	// Open database and initialize schema.
	db, err := storage.Open(cfg.Storage.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Seed the configured user and root node.
	userID, err := db.SeedUser(cfg.User.Email, cfg.User.Password)
	if err != nil {
		log.Fatalf("Failed to seed user: %v", err)
	}
	log.Printf("  User ID: %d", userID)

	// Initialize stores.
	tokenStore := storage.NewTokenStore(db)
	nodeStore := storage.NewNodeStore(db)
	contentStore := storage.NewContentStore(db)

	contentFS, err := content.NewStore(cfg.Storage.ContentDir)
	if err != nil {
		log.Fatalf("Failed to create content store: %v", err)
	}

	// Initialize auth.
	authenticator := auth.New(tokenStore)

	// Set up handlers and routes.
	handlers := api.NewHandlers(cfg, authenticator, tokenStore, nodeStore, contentStore, contentFS, userID)

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)

	// Start server.
	log.Printf("Tucha server listening on %s", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

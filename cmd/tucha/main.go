// Package main is the entry point for the Tucha server.
// This is the composition root -- the only place that imports infrastructure packages.
package main

import (
	"flag"
	"log"
	"net/http"

	"tucha/internal/application/service"
	"tucha/internal/config"
	"tucha/internal/infrastructure/contentstore"
	"tucha/internal/infrastructure/hasher"
	"tucha/internal/infrastructure/sqlite"
	"tucha/internal/transport/httpapi"
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

	// --- Infrastructure ---

	db, err := sqlite.Open(cfg.Storage.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	diskStore, err := contentstore.NewDiskStore(cfg.Storage.ContentDir)
	if err != nil {
		log.Fatalf("Failed to create content store: %v", err)
	}

	mrCloudHasher := hasher.NewMrCloud()

	// --- Repositories ---

	userRepo := sqlite.NewUserRepository(db)
	tokenRepo := sqlite.NewTokenRepository(db)
	nodeRepo := sqlite.NewNodeRepository(db)
	contentRepo := sqlite.NewContentRepository(db)

	// --- Application services ---

	seedSvc := service.NewSeedService(userRepo, nodeRepo)
	userID, err := seedSvc.Seed(cfg.User.Email, cfg.User.Password)
	if err != nil {
		log.Fatalf("Failed to seed user: %v", err)
	}
	log.Printf("  User ID: %d", userID)

	authSvc := service.NewAuthService(tokenRepo)
	tokenSvc := service.NewTokenService(tokenRepo)
	quotaSvc := service.NewQuotaService(nodeRepo, cfg.Storage.QuotaBytes)
	folderSvc := service.NewFolderService(nodeRepo)
	fileSvc := service.NewFileService(nodeRepo, contentRepo, diskStore, quotaSvc)
	uploadSvc := service.NewUploadService(mrCloudHasher, diskStore, contentRepo)
	downloadSvc := service.NewDownloadService(nodeRepo, diskStore)

	// --- Transport (HTTP handlers) ---

	presenter := httpapi.NewPresenter()

	tokenH := httpapi.NewTokenHandler(tokenSvc, cfg.User.Email, cfg.User.Password, userID)
	csrfH := httpapi.NewCSRFHandler(authSvc, cfg.User.Email)
	dispatchH := httpapi.NewDispatchHandler(authSvc, cfg.Server.ExternalURL, cfg.User.Email)
	folderH := httpapi.NewFolderHandler(authSvc, folderSvc, presenter, cfg.User.Email, userID)
	fileH := httpapi.NewFileHandler(authSvc, fileSvc, presenter, cfg.User.Email, userID)
	uploadH := httpapi.NewUploadHandler(authSvc, uploadSvc)
	downloadH := httpapi.NewDownloadHandler(authSvc, downloadSvc, userID)
	spaceH := httpapi.NewSpaceHandler(authSvc, quotaSvc, cfg.User.Email, userID)

	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, tokenH, csrfH, dispatchH, folderH, fileH, uploadH, downloadH, spaceH)

	// --- Start server ---

	log.Printf("Tucha server listening on %s", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

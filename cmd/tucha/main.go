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
	trashRepo := sqlite.NewTrashRepository(db)
	shareRepo := sqlite.NewShareRepository(db)

	// --- Application services ---

	seedSvc := service.NewSeedService(userRepo, nodeRepo, cfg.Storage.QuotaBytes)
	userID, err := seedSvc.Seed(cfg.User.Email, cfg.User.Password, true)
	if err != nil {
		log.Fatalf("Failed to seed user: %v", err)
	}
	log.Printf("  User ID: %d", userID)

	authSvc := service.NewAuthService(tokenRepo, userRepo)
	tokenSvc := service.NewTokenService(tokenRepo, userRepo)
	quotaSvc := service.NewQuotaService(nodeRepo, userRepo)
	userSvc := service.NewUserService(userRepo, nodeRepo, cfg.Storage.QuotaBytes)
	folderSvc := service.NewFolderService(nodeRepo)
	fileSvc := service.NewFileService(nodeRepo, contentRepo, diskStore, quotaSvc)
	uploadSvc := service.NewUploadService(mrCloudHasher, diskStore, contentRepo)
	downloadSvc := service.NewDownloadService(nodeRepo, diskStore)
	trashSvc := service.NewTrashService(nodeRepo, trashRepo, contentRepo, diskStore)
	publishSvc := service.NewPublishService(nodeRepo, contentRepo)
	shareSvc := service.NewShareService(shareRepo, nodeRepo, contentRepo, userRepo)

	// --- Transport (HTTP handlers) ---

	presenter := httpapi.NewPresenter()

	tokenH := httpapi.NewTokenHandler(tokenSvc)
	csrfH := httpapi.NewCSRFHandler(authSvc)
	dispatchH := httpapi.NewDispatchHandler(authSvc, cfg.Server.ExternalURL)
	folderH := httpapi.NewFolderHandler(authSvc, folderSvc, publishSvc, presenter)
	fileH := httpapi.NewFileHandler(authSvc, fileSvc, trashSvc, presenter)
	uploadH := httpapi.NewUploadHandler(authSvc, uploadSvc)
	downloadH := httpapi.NewDownloadHandler(authSvc, downloadSvc)
	spaceH := httpapi.NewSpaceHandler(authSvc, quotaSvc)
	selfConfigH := httpapi.NewSelfConfigureHandler(cfg.Endpoints)
	userH := httpapi.NewUserHandler(authSvc, userSvc)
	adminH := httpapi.NewAdminHandler()
	trashH := httpapi.NewTrashHandler(authSvc, trashSvc, presenter)
	publishH := httpapi.NewPublishHandler(authSvc, publishSvc, presenter)
	weblinkH := httpapi.NewWeblinkDownloadHandler(publishSvc, downloadSvc, folderSvc, presenter, cfg.Server.ExternalURL)
	shareH := httpapi.NewShareHandler(authSvc, shareSvc, presenter)

	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, tokenH, csrfH, dispatchH, folderH, fileH, uploadH, downloadH, spaceH, selfConfigH, userH, adminH, trashH, publishH, weblinkH, shareH)

	// --- Start server ---

	log.Printf("Tucha server listening on %s", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

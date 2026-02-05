// Package main is the entry point for the Tucha server.
// This is the composition root -- the only place that imports infrastructure packages.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"tucha/internal/application/service"
	"tucha/internal/config"
	"tucha/internal/infrastructure/contentstore"
	"tucha/internal/infrastructure/hasher"
	"tucha/internal/infrastructure/logger"
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

	// --- Logger (created first, used by all components) ---

	appLogger, err := logger.New(cfg.Logging.Level, cfg.Logging.Output, cfg.Logging.File)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer appLogger.Close()

	appLogger.Info("Tucha server starting")
	appLogger.Info("  Listen: %s", cfg.Addr())
	appLogger.Info("  External URL: %s", cfg.Server.ExternalURL)
	appLogger.Info("  Admin: %s", cfg.Admin.Login)
	appLogger.Info("  Database: %s", cfg.Storage.DBPath)
	appLogger.Info("  Content dir: %s", cfg.Storage.ContentDir)
	appLogger.Info("  Quota: %d bytes", cfg.Storage.QuotaBytes)
	appLogger.Info("  Token TTL: %d seconds", cfg.Auth.TokenTTLSeconds)
	appLogger.Debug("  Log level: %s", cfg.Logging.Level)
	appLogger.Debug("  Log output: %s", cfg.Logging.Output)

	// --- Infrastructure ---

	db, err := sqlite.Open(cfg.Storage.DBPath)
	if err != nil {
		appLogger.Error("Failed to open database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	diskStore, err := contentstore.NewDiskStore(cfg.Storage.ContentDir)
	if err != nil {
		appLogger.Error("Failed to create content store: %v", err)
		os.Exit(1)
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

	adminAuthSvc := service.NewAdminAuthService(cfg.Admin.Login, cfg.Admin.Password)
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

	tokenH := httpapi.NewTokenHandler(tokenSvc, cfg.Auth.TokenTTLSeconds, appLogger)
	csrfH := httpapi.NewCSRFHandler(authSvc)
	dispatchH := httpapi.NewDispatchHandler(authSvc, cfg.Server.ExternalURL)
	folderH := httpapi.NewFolderHandler(authSvc, folderSvc, publishSvc, presenter)
	fileH := httpapi.NewFileHandler(authSvc, fileSvc, trashSvc, presenter)
	uploadH := httpapi.NewUploadHandler(authSvc, uploadSvc)
	downloadH := httpapi.NewDownloadHandler(authSvc, downloadSvc)
	spaceH := httpapi.NewSpaceHandler(authSvc, quotaSvc)
	selfConfigH := httpapi.NewSelfConfigureHandler(cfg.Endpoints)
	userH := httpapi.NewUserHandler(adminAuthSvc, userSvc)
	adminH := httpapi.NewAdminHandler(adminAuthSvc)
	trashH := httpapi.NewTrashHandler(authSvc, trashSvc, presenter)
	publishH := httpapi.NewPublishHandler(authSvc, publishSvc, presenter)
	weblinkH := httpapi.NewWeblinkDownloadHandler(publishSvc, downloadSvc, folderSvc, presenter, cfg.Server.ExternalURL)
	shareH := httpapi.NewShareHandler(authSvc, shareSvc, presenter)

	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, tokenH, csrfH, dispatchH, folderH, fileH, uploadH, downloadH, spaceH, selfConfigH, userH, adminH, trashH, publishH, weblinkH, shareH)

	// --- Start server ---

	appLogger.Info("Tucha server listening on %s", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), mux); err != nil {
		appLogger.Error("Server failed: %v", err)
		os.Exit(1)
	}
}

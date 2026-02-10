// Package main is the entry point for the Tucha server.
// This is the composition root -- the only place that imports infrastructure packages.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/cli"
	"github.com/pozitronik/tucha/internal/config"
	"github.com/pozitronik/tucha/internal/infrastructure/contentstore"
	"github.com/pozitronik/tucha/internal/infrastructure/hasher"
	"github.com/pozitronik/tucha/internal/infrastructure/logger"
	"github.com/pozitronik/tucha/internal/infrastructure/sqlite"
	"github.com/pozitronik/tucha/internal/infrastructure/thumbnail"
	"github.com/pozitronik/tucha/internal/transport/httpapi"
)

func main() {
	parsed, err := cli.Parse(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(cli.ExitError)
	}

	switch parsed.Command {
	case cli.CmdHelp:
		fmt.Print(cli.HelpText())
		return

	case cli.CmdVersion:
		fmt.Printf("Tucha version %s\n", version)
		return

	case cli.CmdStatus:
		runStatus(parsed.ConfigPath)

	case cli.CmdStop:
		runStop(parsed.ConfigPath)

	case cli.CmdConfigCheck:
		runConfigCheck(parsed.ConfigPath)

	case cli.CmdUserList, cli.CmdUserAdd, cli.CmdUserRemove, cli.CmdUserPwd, cli.CmdUserQuota, cli.CmdUserSizeLimit, cli.CmdUserHistory, cli.CmdUserInfo:
		runUserCommand(parsed)

	case cli.CmdRun, cli.CmdBackground:
		runServer(parsed)
	}
}

func runStatus(configPath string) {
	pidFile := cli.DefaultPIDFile(configPath)
	running, pid, err := cli.ServerStatus(pidFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(cli.ExitError)
	}

	if running {
		fmt.Printf("Tucha is running (PID: %d)\n", pid)
	} else {
		fmt.Println("Tucha is not running")
		os.Exit(cli.ExitNotRunning)
	}
}

func runStop(configPath string) {
	pidFile := cli.DefaultPIDFile(configPath)
	if err := cli.StopServer(pidFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(cli.ExitNotRunning)
	}
	fmt.Println("Tucha stopped")
}

func runConfigCheck(configPath string) {
	_, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(cli.ExitConfigError)
	}
	fmt.Println("Configuration is valid")
}

func runUserCommand(parsed *cli.CLI) {
	cfg, err := config.Load(parsed.ConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(cli.ExitConfigError)
	}

	db, err := sqlite.Open(cfg.Storage.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(cli.ExitError)
	}
	defer db.Close()

	userRepo := sqlite.NewUserRepository(db)
	nodeRepo := sqlite.NewNodeRepository(db)
	userSvc := service.NewUserService(userRepo, nodeRepo, cfg.Storage.QuotaBytes)
	cmds := cli.NewUserCommands(userSvc, userRepo)

	var cmdErr error
	switch parsed.Command {
	case cli.CmdUserList:
		pattern := ""
		if len(parsed.Args) > 0 {
			pattern = parsed.Args[0]
		}
		cmdErr = cmds.List(os.Stdout, pattern)

	case cli.CmdUserAdd:
		quota := ""
		if len(parsed.Args) > 2 {
			quota = parsed.Args[2]
		}
		cmdErr = cmds.Add(os.Stdout, parsed.Args[0], parsed.Args[1], quota)

	case cli.CmdUserRemove:
		cmdErr = cmds.Remove(os.Stdout, parsed.Args[0])

	case cli.CmdUserPwd:
		cmdErr = cmds.SetPassword(os.Stdout, parsed.Args[0], parsed.Args[1])

	case cli.CmdUserQuota:
		cmdErr = cmds.SetQuota(os.Stdout, parsed.Args[0], parsed.Args[1])

	case cli.CmdUserSizeLimit:
		cmdErr = cmds.SetSizeLimit(os.Stdout, parsed.Args[0], parsed.Args[1])

	case cli.CmdUserHistory:
		cmdErr = cmds.SetHistory(os.Stdout, parsed.Args[0], parsed.Args[1])

	case cli.CmdUserInfo:
		cmdErr = cmds.Info(os.Stdout, parsed.Args[0])
	}

	if cmdErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", cmdErr)
		os.Exit(cli.ExitError)
	}
}

func runServer(parsed *cli.CLI) {
	cfg, err := config.Load(parsed.ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Determine PID file path
	pidFile := cfg.Server.PIDFile
	if pidFile == "" {
		pidFile = cli.DefaultPIDFile(parsed.ConfigPath)
	}

	// Check if already running (for both foreground and background modes)
	if err := cli.CheckNotRunning(pidFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(cli.ExitAlreadyRunning)
	}

	// Background mode: relaunch as daemon
	if parsed.Command == cli.CmdBackground {
		if err := daemonize(parsed.ConfigPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting background process: %v\n", err)
			os.Exit(cli.ExitError)
		}
		fmt.Println("Tucha started in background")
		return
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
	appLogger.Info("  Thumbnail dir: %s", cfg.Storage.ThumbnailDir)
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

	thumbGen, err := thumbnail.NewGenerator(cfg.Storage.ThumbnailDir)
	if err != nil {
		appLogger.Error("Failed to create thumbnail generator: %v", err)
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
	fileVersionRepo := sqlite.NewFileVersionRepository(db)

	// --- Application services ---

	adminAuthSvc := service.NewAdminAuthService(cfg.Admin.Login, cfg.Admin.Password)
	authSvc := service.NewAuthService(tokenRepo, userRepo)
	tokenSvc := service.NewTokenService(tokenRepo, userRepo)
	quotaSvc := service.NewQuotaService(nodeRepo, userRepo)
	userSvc := service.NewUserService(userRepo, nodeRepo, cfg.Storage.QuotaBytes)
	folderSvc := service.NewFolderService(nodeRepo)
	fileSvc := service.NewFileService(nodeRepo, contentRepo, diskStore, quotaSvc, fileVersionRepo)
	uploadSvc := service.NewUploadService(mrCloudHasher, diskStore, contentRepo)
	downloadSvc := service.NewDownloadService(nodeRepo, diskStore)
	thumbnailSvc := service.NewThumbnailService(nodeRepo, diskStore, thumbGen)
	trashSvc := service.NewTrashService(nodeRepo, trashRepo, contentRepo, diskStore)
	publishSvc := service.NewPublishService(nodeRepo, contentRepo)
	shareSvc := service.NewShareService(shareRepo, nodeRepo, contentRepo, userRepo)

	// --- Transport (HTTP handlers) ---

	presenter := httpapi.NewPresenter()

	tokenH := httpapi.NewTokenHandler(tokenSvc, cfg.Auth.TokenTTLSeconds, appLogger)
	csrfH := httpapi.NewCSRFHandler(authSvc)
	dispatchH := httpapi.NewDispatchHandler(authSvc, cfg.Server.ExternalURL)
	folderH := httpapi.NewFolderHandler(authSvc, folderSvc, shareSvc, publishSvc, presenter)
	fileH := httpapi.NewFileHandler(authSvc, fileSvc, trashSvc, presenter)
	uploadH := httpapi.NewUploadHandler(authSvc, uploadSvc, fileSvc)
	downloadH := httpapi.NewDownloadHandler(authSvc, downloadSvc)
	spaceH := httpapi.NewSpaceHandler(authSvc, quotaSvc)
	selfConfigH := httpapi.NewSelfConfigureHandler(cfg.Endpoints)
	userH := httpapi.NewUserHandler(adminAuthSvc, userSvc)
	adminH := httpapi.NewAdminHandler(adminAuthSvc)
	trashH := httpapi.NewTrashHandler(authSvc, trashSvc, presenter)
	publishH := httpapi.NewPublishHandler(authSvc, publishSvc, presenter)
	weblinkH := httpapi.NewWeblinkDownloadHandler(publishSvc, downloadSvc, folderSvc, presenter, cfg.Server.ExternalURL)
	shareH := httpapi.NewShareHandler(authSvc, shareSvc, presenter)
	thumbnailH := httpapi.NewThumbnailHandler(authSvc, thumbnailSvc)
	publicThumbH := httpapi.NewPublicThumbnailHandler(publishSvc, thumbnailSvc)
	videoH := httpapi.NewVideoHandler(publishSvc, downloadSvc, cfg.Server.ExternalURL)

	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux, tokenH, csrfH, dispatchH, folderH, fileH, uploadH, downloadH, spaceH, selfConfigH, userH, adminH, trashH, publishH, weblinkH, shareH, thumbnailH, publicThumbH, videoH)

	// --- Start server with graceful shutdown ---

	appLogger.Info("Tucha server listening on %s", cfg.Addr())
	if err := cli.RunServerWithGracefulShutdown(cfg.Addr(), mux, pidFile, 30*time.Second); err != nil {
		appLogger.Error("Server failed: %v", err)
		os.Exit(1)
	}
	appLogger.Info("Tucha server stopped")
}

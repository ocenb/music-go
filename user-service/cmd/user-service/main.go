package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ocenb/music-go/user-service/internal/app"
	"github.com/ocenb/music-go/user-service/internal/config"
	"github.com/ocenb/music-go/user-service/internal/logger"
	authrepo "github.com/ocenb/music-go/user-service/internal/repos/auth"
	tokenrepo "github.com/ocenb/music-go/user-service/internal/repos/token"
	userrepo "github.com/ocenb/music-go/user-service/internal/repos/user"
	"github.com/ocenb/music-go/user-service/internal/services/auth"
	"github.com/ocenb/music-go/user-service/internal/services/token"
	"github.com/ocenb/music-go/user-service/internal/services/user"
	"github.com/ocenb/music-go/user-service/internal/storage/postgres"
	"github.com/ocenb/music-go/user-service/internal/utils"
)

func main() {
	startTime := time.Now()
	cfg := config.MustLoad()
	log := logger.Setup(cfg)

	log.Info("Connecting to database",
		slog.String("host", cfg.DBHost),
		slog.String("port", cfg.DBPort),
		slog.String("database", cfg.DBName),
	)
	postgres, err := postgres.New(cfg)
	if err != nil {
		log.Error("Failed to connect to postgres", utils.ErrLog(err))
		os.Exit(1)
	}
	defer func() {
		log.Info("Closing database connection")
		err := postgres.Close()
		if err != nil {
			log.Error("Failed to close postgres connection", utils.ErrLog(err))
		}
	}()

	tokenRepo := tokenrepo.NewTokenRepo(postgres)
	userRepo := userrepo.NewUserRepo(postgres)
	authRepo := authrepo.NewAuthRepo(postgres)

	tokenService := token.NewTokenService(cfg, log, tokenRepo)
	userService := user.NewUserService(cfg, log, userRepo)
	authService := auth.NewAuthService(cfg, log, userService, tokenService, authRepo)

	go runTokenCleanup(tokenService, log)

	log.Info("Initializing gRPC server", slog.Int("port", cfg.GRPC.Port))
	grpcApp := app.New(authService, userService, cfg.GRPC.Port, cfg, log)

	go func() {
		grpcApp.Run()
	}()

	log.Info("Service started successfully", slog.Duration("startup_time", time.Since(startTime)))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	shutdownStart := time.Now()

	grpcApp.Stop()

	log.Info("Service shutdown complete",
		slog.Duration("shutdown_time", time.Since(shutdownStart)),
		slog.Duration("uptime", time.Since(startTime)))
}

func runTokenCleanup(tokenService token.TokenServiceInterface, log *slog.Logger) {
	log.Info("Token cleanup scheduled", slog.Duration("interval", 24*time.Hour))
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	tokenService.CleanupExpiredTokens(log)
	for range ticker.C {
		tokenService.CleanupExpiredTokens(log)
	}
}

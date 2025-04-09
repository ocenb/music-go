package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ocenb/music-go/content-service/internal/app"
	"github.com/ocenb/music-go/content-service/internal/config"
	"github.com/ocenb/music-go/content-service/internal/logger"
	"github.com/ocenb/music-go/content-service/internal/storage/postgres"
	"github.com/ocenb/music-go/content-service/internal/utils"
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

	log.Info("Initializing HTTP server", slog.Int("port", cfg.Port))
	httpApp := app.New(postgres, cfg, log)

	go func() {
		httpApp.Run()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	shutdownStart := time.Now()

	httpApp.Stop()

	log.Info("Service shutdown complete",
		slog.Duration("shutdown_time", time.Since(shutdownStart)),
		slog.Duration("uptime", time.Since(startTime)))
}

package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ocenb/music-go/content-service/internal/app"
	"github.com/ocenb/music-go/content-service/internal/clients/cloudinaryclient"
	"github.com/ocenb/music-go/content-service/internal/clients/notificationclient"
	"github.com/ocenb/music-go/content-service/internal/clients/searchclient"
	"github.com/ocenb/music-go/content-service/internal/clients/userclient"
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

	cloudinary, err := cloudinaryclient.NewCloudinaryClient(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryApiKey,
		cfg.CloudinaryApiSecret,
	)
	if err != nil {
		log.Error("Failed to create cloudinary service", utils.ErrLog(err))
		os.Exit(1)
	}

	searchServiceClient, err := searchclient.New(cfg)
	if err != nil {
		log.Error("Failed to connect to search service", utils.ErrLog(err))
		os.Exit(1)
	}
	defer func() {
		log.Info("Closing search service connection")
		err := searchServiceClient.Conn.Close()
		if err != nil {
			log.Error("Failed to close search service connection", utils.ErrLog(err))
		}
	}()

	userServiceClient, err := userclient.New(cfg)
	if err != nil {
		log.Error("Failed to connect to user service", utils.ErrLog(err))
		os.Exit(1)
	}
	defer func() {
		log.Info("Closing user service connection")
		err := userServiceClient.Conn.Close()
		if err != nil {
			log.Error("Failed to close user service connection", utils.ErrLog(err))
		}
	}()

	notificationClient, err := notificationclient.NewNotificationClient(cfg.KafkaBrokers)
	if err != nil {
		log.Error("Failed to create notification client", utils.ErrLog(err))
		os.Exit(1)
	}
	defer func() {
		log.Info("Closing notification client")
		err := notificationClient.Close()
		if err != nil {
			log.Error("Failed to close notification client", utils.ErrLog(err))
		}
	}()

	log.Info("Initializing HTTP server", slog.Int("port", cfg.Port))
	httpApp := app.New(postgres, cfg, log, cloudinary, searchServiceClient, userServiceClient, notificationClient)

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

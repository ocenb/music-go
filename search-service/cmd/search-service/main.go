package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ocenb/music-go/search-service/internal/app"
	"github.com/ocenb/music-go/search-service/internal/clients/elastic"
	"github.com/ocenb/music-go/search-service/internal/clients/userservice"
	"github.com/ocenb/music-go/search-service/internal/config"
	"github.com/ocenb/music-go/search-service/internal/logger"
	"github.com/ocenb/music-go/search-service/internal/service"
	"github.com/ocenb/music-go/search-service/internal/utils"
)

func main() {
	startTime := time.Now()
	cfg := config.MustLoad()
	log := logger.Setup(cfg)

	log.Info("Connecting to elasticsearch",
		slog.String("host", cfg.ElasticHost),
		slog.String("port", cfg.ElasticPort),
	)
	elasticClient, err := elastic.New(cfg, log)
	if err != nil {
		log.Error("Failed to connect to elasticsearch", utils.ErrLog(err))
		os.Exit(1)
	}
	userServiceClient, err := userservice.New(cfg)
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

	searchService := service.NewSearchService(cfg, log, elasticClient)

	log.Info("Initializing gRPC server", slog.Int("port", cfg.GRPC.Port))
	grpcApp := app.New(searchService, userServiceClient, cfg, log)

	go func() {
		grpcApp.Run()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	shutdownStart := time.Now()

	grpcApp.Stop()

	log.Info("Service shutdown complete",
		slog.Duration("shutdown_time", time.Since(shutdownStart)),
		slog.Duration("uptime", time.Since(startTime)))
}

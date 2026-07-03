package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ocenb/music-go/search-service/internal/clients/user"
	"github.com/ocenb/music-go/search-service/internal/config"
	"github.com/ocenb/music-go/search-service/internal/logger"
	"github.com/ocenb/music-go/search-service/internal/logger/logattr"
	searchrepo "github.com/ocenb/music-go/search-service/internal/repos/search"
	"github.com/ocenb/music-go/search-service/internal/server"
	searchservice "github.com/ocenb/music-go/search-service/internal/services/search"
	"github.com/ocenb/music-go/search-service/internal/storage/elastic"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		//nolint:sloglint // default logger is used before initialization
		slog.Error("cannot load config", logattr.Err(err))
		return 1
	}
	log := logger.New(cfg.Log.Level, cfg.Log.Handler, cfg.Environment)
	slog.SetDefault(log)

	defer log.Info("app stopped")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	connectCtx, connectCancel := context.WithTimeout(ctx, cfg.ElasticConnectTimeout)
	defer connectCancel()

	log.Info("connecting to elasticsearch",
		slog.String("host", cfg.Elastic.Host),
		slog.String("port", cfg.Elastic.Port),
	)
	elasticClient, err := elastic.New(connectCtx, cfg.Elastic, log)
	if err != nil {
		log.Error("initialization failed", logattr.Err(err))
		return 1
	}

	userClient, err := user.New(cfg.UserServiceAddress, cfg.GRPC.Timeout, log)
	if err != nil {
		log.Error("failed to create user service client", logattr.Err(err))
		return 1
	}
	defer func() {
		log.Info("closing user service connection")
		if err := userClient.Close(); err != nil {
			log.Error("failed to close user service connection", logattr.Err(err))
		}
	}()

	repo := searchrepo.New(elasticClient)
	svc := searchservice.New(repo)

	log.Info("initializing gRPC server", slog.Int("port", cfg.GRPC.Port))
	gRPCServer := server.New(svc, userClient, log, cfg.GRPC.Port)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- gRPCServer.Start()
	}()

	var serverErr error
	select {
	case serverErr = <-serverErrors:
		if serverErr != nil {
			log.Error("server crashed", logattr.Err(serverErr))
		} else {
			log.Error("server stopped unexpectedly")
		}
	case <-ctx.Done():
		log.Info("received shutdown signal")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	shutdownErr := gRPCServer.Stop(shutdownCtx)
	if shutdownErr != nil {
		log.Error("gRPC server shutdown error", logattr.Err(shutdownErr))
	}

	if shutdownErr != nil || serverErr != nil {
		return 1
	}
	return 0
}

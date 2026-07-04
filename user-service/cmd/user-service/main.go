package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ocenb/music-go/user-service/internal/clients/content"
	"github.com/ocenb/music-go/user-service/internal/clients/notification"
	"github.com/ocenb/music-go/user-service/internal/clients/search"
	"github.com/ocenb/music-go/user-service/internal/config"
	"github.com/ocenb/music-go/user-service/internal/logger"
	"github.com/ocenb/music-go/user-service/internal/logger/logattr"
	tokenrepo "github.com/ocenb/music-go/user-service/internal/repos/token"
	userrepo "github.com/ocenb/music-go/user-service/internal/repos/user"
	"github.com/ocenb/music-go/user-service/internal/server"
	authservice "github.com/ocenb/music-go/user-service/internal/services/auth"
	tokenservice "github.com/ocenb/music-go/user-service/internal/services/token"
	userservice "github.com/ocenb/music-go/user-service/internal/services/user"
	"github.com/ocenb/music-go/user-service/internal/storage/migrator"
	"github.com/ocenb/music-go/user-service/internal/storage/postgres"
	"github.com/ocenb/music-go/user-service/internal/storage/transactor"
	"github.com/ocenb/music-go/user-service/migrations"
)

type tokenCleaner interface {
	CleanupExpiredTokens() error
}

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

	err = func() error {
		log.Info("connecting to database for migrations",
			slog.String("host", cfg.Postgres.Host),
			slog.String("port", cfg.Postgres.Port),
			slog.String("database", cfg.Postgres.Name),
		)
		migrateCtx, migrateCancel := context.WithTimeout(ctx, cfg.DBConnectTimeout)
		defer migrateCancel()
		migrate, err := migrator.New(migrateCtx, cfg.Postgres.DSN, migrations.FS, logger.NewSlogAdapter(log))
		if err != nil {
			return err
		}
		defer func() {
			if err := migrate.Close(); err != nil {
				log.Error("failed to close migrate connection", logattr.Err(err))
			}
		}()

		log.Info("running migrations")
		if err := migrate.Up(); err != nil {
			return err
		}
		log.Info("migrations completed successfully")
		return nil
	}()
	if err != nil {
		log.Error("initialization failed", logattr.Err(err))
		return 1
	}

	log.Info("connecting to database",
		slog.String("host", cfg.Postgres.Host),
		slog.String("port", cfg.Postgres.Port),
		slog.String("database", cfg.Postgres.Name),
	)
	connectCtx, connectCancel := context.WithTimeout(ctx, cfg.DBConnectTimeout)
	defer connectCancel()
	pool, err := postgres.NewPool(connectCtx, cfg.Postgres)
	if err != nil {
		log.Error("initialization failed", logattr.Err(err))
		return 1
	}
	defer pool.Close()

	tm := transactor.New(pool)

	searchClient, err := search.New(cfg.SearchServiceAddress, cfg.GRPC.Timeout, log)
	if err != nil {
		log.Error("failed to create search service client", logattr.Err(err))
		return 1
	}
	defer func() {
		log.Info("closing search service connection")
		if err := searchClient.Close(); err != nil {
			log.Error("failed to close search service connection", logattr.Err(err))
		}
	}()

	notificationClient, err := notification.New(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		log.Error("failed to create notification client", logattr.Err(err))
		return 1
	}
	defer func() {
		log.Info("closing notification client")
		if err := notificationClient.Close(); err != nil {
			log.Error("failed to close notification client", logattr.Err(err))
		}
	}()

	contentClient := content.New(cfg.ContentServiceURL, cfg.InternalServiceSecret, cfg.GRPC.Timeout)

	tokenRepo := tokenrepo.New(tm)
	userRepo := userrepo.New(tm)

	tokenService := tokenservice.New(tokenRepo, cfg.Auth.JWTSecret, cfg.Auth.AccessTokenLiveTime, cfg.Auth.RefreshTokenLiveTime)
	userService := userservice.New(userRepo, tm, searchClient, contentClient)
	authService := authservice.New(userService, tokenService, notificationClient, tm, cfg.Auth.BCryptCost)

	go runTokenCleanup(ctx, tokenService, log, cfg.Auth.TokenCleanupInterval)

	log.Info("initializing gRPC server", slog.Int("port", cfg.GRPC.Port))
	gRPCServer := server.New(authService, userService, log, cfg.GRPC.Port)

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

func runTokenCleanup(ctx context.Context, cleaner tokenCleaner, log *slog.Logger, interval time.Duration) {
	log.Info("token cleanup scheduled", slog.Duration("interval", interval))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	cleanup := func() {
		if err := cleaner.CleanupExpiredTokens(); err != nil {
			log.Error("failed to cleanup expired tokens", logattr.Err(err))
			return
		}
		log.Info("successfully cleaned up expired tokens")
	}

	cleanup()

	for {
		select {
		case <-ctx.Done():
			log.Info("stopping token cleanup")
			return
		case <-ticker.C:
			cleanup()
		}
	}
}

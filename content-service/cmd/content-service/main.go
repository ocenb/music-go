package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ocenb/music-go/content-service/internal/clients/cloudinary"
	"github.com/ocenb/music-go/content-service/internal/clients/notification"
	"github.com/ocenb/music-go/content-service/internal/clients/search"
	"github.com/ocenb/music-go/content-service/internal/clients/user"
	"github.com/ocenb/music-go/content-service/internal/config"
	"github.com/ocenb/music-go/content-service/internal/logger"
	"github.com/ocenb/music-go/content-service/internal/logger/logattr"
	allrepo "github.com/ocenb/music-go/content-service/internal/repos/all"
	historyrepo "github.com/ocenb/music-go/content-service/internal/repos/history"
	playlistrepo "github.com/ocenb/music-go/content-service/internal/repos/playlist"
	playlisttracksrepo "github.com/ocenb/music-go/content-service/internal/repos/playlisttracks"
	trackrepo "github.com/ocenb/music-go/content-service/internal/repos/track"
	"github.com/ocenb/music-go/content-service/internal/server"
	allservice "github.com/ocenb/music-go/content-service/internal/services/all"
	fileservice "github.com/ocenb/music-go/content-service/internal/services/file"
	historyservice "github.com/ocenb/music-go/content-service/internal/services/history"
	playlistservice "github.com/ocenb/music-go/content-service/internal/services/playlist"
	playlisttracksservice "github.com/ocenb/music-go/content-service/internal/services/playlisttracks"
	searchservice "github.com/ocenb/music-go/content-service/internal/services/search"
	trackservice "github.com/ocenb/music-go/content-service/internal/services/track"
	"github.com/ocenb/music-go/content-service/internal/storage/migrator"
	"github.com/ocenb/music-go/content-service/internal/storage/postgres"
	"github.com/ocenb/music-go/content-service/internal/storage/transactor"
	"github.com/ocenb/music-go/content-service/migrations"
)

const defaultClientTimeout = 5 * time.Second

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
		migrate, migrateErr := migrator.New(migrateCtx, cfg.Postgres.DSN, migrations.FS, logger.NewSlogAdapter(log))
		if migrateErr != nil {
			return migrateErr
		}
		defer func() {
			if closeErr := migrate.Close(); closeErr != nil {
				log.Error("failed to close migrate connection", logattr.Err(closeErr))
			}
		}()

		log.Info("running migrations")
		if migrateErr = migrate.Up(); migrateErr != nil {
			return migrateErr
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

	clientTimeout := defaultClientTimeout

	cloudinaryClient, err := cloudinary.New(cfg.Cloudinary.CloudName, cfg.Cloudinary.APIKey, cfg.Cloudinary.APISecret)
	if err != nil {
		log.Error("failed to create cloudinary client", logattr.Err(err))
		return 1
	}

	searchClient, err := search.New(cfg.SearchServiceAddress, clientTimeout, log)
	if err != nil {
		log.Error("failed to create search service client", logattr.Err(err))
		return 1
	}
	defer func() {
		log.Info("closing search service connection")
		if closeErr := searchClient.Close(); closeErr != nil {
			log.Error("failed to close search service connection", logattr.Err(closeErr))
		}
	}()

	userClient, err := user.New(cfg.UserServiceAddress, clientTimeout, log)
	if err != nil {
		log.Error("failed to create user service client", logattr.Err(err))
		return 1
	}
	defer func() {
		log.Info("closing user service connection")
		if closeErr := userClient.Close(); closeErr != nil {
			log.Error("failed to close user service connection", logattr.Err(closeErr))
		}
	}()

	notificationClient, err := notification.New(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		log.Error("failed to create notification client", logattr.Err(err))
		return 1
	}
	defer func() {
		log.Info("closing notification client")
		if closeErr := notificationClient.Close(); closeErr != nil {
			log.Error("failed to close notification client", logattr.Err(closeErr))
		}
	}()

	fileSvc := fileservice.New(cloudinaryClient, log, cfg)
	trackRepo := trackrepo.New(tm)
	playlistRepo := playlistrepo.New(tm)
	playlistTracksRepo := playlisttracksrepo.New(tm)
	historyRepo := historyrepo.New(tm)
	allRepo := allrepo.New(tm)

	trackSvc := trackservice.New(trackRepo, fileSvc, searchClient, notificationClient, tm)
	playlistSvc := playlistservice.New(playlistRepo, fileSvc, tm)
	playlistTracksSvc := playlisttracksservice.New(playlistTracksRepo, playlistRepo, trackRepo, tm)
	historySvc := historyservice.New(historyRepo, trackSvc)
	allSvc := allservice.New(allRepo, fileSvc)
	searchSvc := searchservice.New(searchClient)

	log.Info("initializing HTTP server", slog.Int("port", cfg.HTTP.Port))
	httpServer := server.New(
		trackSvc,
		playlistSvc,
		playlistTracksSvc,
		historySvc,
		searchSvc,
		allSvc,
		userClient,
		cfg.InternalServiceSecret,
		cfg,
		log,
	)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- httpServer.Start()
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

	shutdownErr := httpServer.Stop(shutdownCtx)
	if shutdownErr != nil {
		log.Error("HTTP server shutdown error", logattr.Err(shutdownErr))
	}

	if shutdownErr != nil || serverErr != nil {
		return 1
	}
	return 0
}

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ocenb/music-go/notification-service/internal/clients/smtp"
	"github.com/ocenb/music-go/notification-service/internal/config"
	"github.com/ocenb/music-go/notification-service/internal/logger"
	"github.com/ocenb/music-go/notification-service/internal/logger/logattr"
	"github.com/ocenb/music-go/notification-service/internal/server"
	notificationservice "github.com/ocenb/music-go/notification-service/internal/services/notification"
	kafkastorage "github.com/ocenb/music-go/notification-service/internal/storage/kafka"
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

	log.Info("connecting to kafka",
		slog.String("topic", cfg.Kafka.Topic),
		slog.String("group_id", cfg.Kafka.GroupID),
	)
	kafkaConsumer, err := kafkastorage.New(cfg.Kafka)
	if err != nil {
		log.Error("initialization failed", logattr.Err(err))
		return 1
	}
	defer func() {
		log.Info("closing kafka consumer")
		if err := kafkaConsumer.Close(); err != nil {
			log.Error("failed to close kafka consumer", logattr.Err(err))
		}
	}()

	log.Info("connecting to smtp",
		slog.String("host", cfg.SMTP.Host),
		slog.Int("port", cfg.SMTP.Port),
	)
	smtpClient := smtp.New(cfg.SMTP)
	svc := notificationservice.New(smtpClient)

	srv := server.New(svc, kafkaConsumer, log)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- srv.Start(ctx)
	}()

	var serverErr error
	select {
	case serverErr = <-serverErrors:
		if serverErr != nil {
			log.Error("consumer crashed", logattr.Err(serverErr))
		} else {
			log.Info("consumer stopped")
		}
	case <-ctx.Done():
		log.Info("received shutdown signal")
	}

	if err := srv.Stop(); err != nil {
		log.Error("consumer shutdown error", logattr.Err(err))
	}

	if serverErr != nil {
		return 1
	}
	return 0
}

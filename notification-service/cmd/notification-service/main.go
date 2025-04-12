package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ocenb/music-go/notification-service/internal/config"
	"github.com/ocenb/music-go/notification-service/internal/handlers"
	"github.com/ocenb/music-go/notification-service/internal/kafka"
	"github.com/ocenb/music-go/notification-service/internal/logger"
	"github.com/ocenb/music-go/notification-service/internal/services/email"
	"github.com/ocenb/music-go/notification-service/internal/utils"
)

func main() {
	startTime := time.Now()
	cfg := config.MustLoad()
	log := logger.Setup(cfg)

	log.Info("Connecting to smtp",
		slog.String("host", cfg.SMTPHost),
		slog.Int("port", cfg.SMTPPort),
	)
	emailService := email.NewEmailService(cfg)

	log.Info("Connecting to kafka",
		slog.String("brokers", strings.Join(cfg.KafkaBrokers, ",")),
		slog.String("topic", cfg.KafkaTopic),
		slog.String("group_id", cfg.KafkaGroupID),
	)

	consumer, err := kafka.NewConsumer(cfg)
	if err != nil {
		log.Error("Failed to create Kafka consumer", utils.ErrLog(err))
		os.Exit(1)
	}
	defer func() {
		log.Info("Closing kafka consumer")
		if err := consumer.Close(); err != nil {
			log.Error("Failed to close kafka consumer", utils.ErrLog(err))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("Starting notification service...")
	err = consumer.Consume(ctx, handlers.EmailNotificationHandler(log, emailService))
	if err != nil {
		log.Error("Error consuming messages", utils.ErrLog(err))
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	shutdownStart := time.Now()
	cancel()

	log.Info("Service shutdown complete",
		slog.Duration("shutdown_time", time.Since(shutdownStart)),
		slog.Duration("uptime", time.Since(startTime)))
}

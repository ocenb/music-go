package server

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ocenb/music-go/notification-service/internal/handlers"
	"github.com/ocenb/music-go/notification-service/internal/logger"
	"github.com/ocenb/music-go/notification-service/internal/logger/logattr"
	"github.com/ocenb/music-go/notification-service/internal/models"
)

type KafkaConsumer interface {
	Read(ctx context.Context) (*models.EmailNotification, error)
	Close() error
}

type Server struct {
	notificationService handlers.NotificationService
	consumer            KafkaConsumer
	log                 *slog.Logger
}

func New(
	notificationService handlers.NotificationService,
	consumer KafkaConsumer,
	log *slog.Logger,
) *Server {
	return &Server{
		notificationService: notificationService,
		consumer:            consumer,
		log:                 log,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.log.Info("starting notification consumer")

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		notification, err := s.consumer.Read(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			s.log.Error("failed to read kafka message", logattr.Err(err))
			continue
		}

		msgCtx := logger.IntoContext(ctx, s.log)
		if err := handlers.HandleEmailNotification(msgCtx, s.notificationService, notification); err != nil {
			s.log.Error("failed to handle notification", logattr.Err(err))
		}
	}
}

func (s *Server) Stop() error {
	s.log.Info("stopping notification consumer")
	return nil
}

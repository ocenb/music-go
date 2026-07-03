package handlers

import (
	"context"
	"log/slog"

	"github.com/ocenb/music-go/notification-service/internal/logger"
	"github.com/ocenb/music-go/notification-service/internal/logger/logattr"
	"github.com/ocenb/music-go/notification-service/internal/models"
)

func HandleEmailNotification(
	ctx context.Context,
	notificationService NotificationService,
	notification *models.EmailNotification,
) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "received notification for email", slog.String("email", notification.Email))

	if err := notificationService.SendEmailNotification(ctx, notification.Email, notification.Msg); err != nil {
		log.ErrorContext(ctx, "failed to send email notification",
			logattr.Err(err),
			slog.String("email", notification.Email),
		)
		return err
	}

	return nil
}

package handlers

import (
	"log/slog"

	"github.com/ocenb/music-go/notification-service/internal/kafka"
	"github.com/ocenb/music-go/notification-service/internal/services/email"
	"github.com/ocenb/music-go/notification-service/internal/utils"
)

func EmailNotificationHandler(log *slog.Logger, emailService *email.EmailService) func(notification kafka.EmailNotification) error {
	return func(notification kafka.EmailNotification) error {
		log.Info("Received notification for email", slog.String("email", notification.Email))
		err := emailService.SendEmailNotification(notification.Email, notification.Msg)
		if err != nil {
			log.Error("Failed to send email notification",
				slog.String("email", notification.Email),
				utils.ErrLog(err))
			return err
		}
		return nil
	}
}

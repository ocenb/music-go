package handlers

import "context"

type NotificationService interface {
	SendEmailNotification(ctx context.Context, email, msg string) error
}

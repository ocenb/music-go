package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/ocenb/music-go/notification-service/internal/errs"
)

type SMTPClient interface {
	Send(ctx context.Context, to, htmlBody string) error
}

type Service struct {
	smtp SMTPClient
}

func New(smtp SMTPClient) *Service {
	return &Service{smtp: smtp}
}

func (s *Service) SendEmailNotification(ctx context.Context, email, msg string) error {
	const (
		maxRetries             = 3
		retryBackoffMultiplier = 2
	)
	var lastErr error

	htmlBody := fmt.Sprintf(`<html><body><p>%s</p></body></html>`, msg)

	for i := range maxRetries {
		if i > 0 {
			time.Sleep(time.Second * time.Duration(i*retryBackoffMultiplier))
		}

		err := s.smtp.Send(ctx, email, htmlBody)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("%w: %w", errs.ErrSendFailed, fmt.Errorf("NotificationService.SendEmailNotification: %w", lastErr))
}

package notification_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ocenb/music-go/notification-service/internal/errs"
	"github.com/ocenb/music-go/notification-service/internal/services/notification"
)

type mockSMTPClient struct {
	sendFn func(ctx context.Context, to, htmlBody string) error
}

func (m *mockSMTPClient) Send(ctx context.Context, to, htmlBody string) error {
	return m.sendFn(ctx, to, htmlBody)
}

func TestService_SendEmailNotification_Success(t *testing.T) {
	var sentTo, sentBody string
	smtpClient := &mockSMTPClient{
		sendFn: func(_ context.Context, to, htmlBody string) error {
			sentTo = to
			sentBody = htmlBody
			return nil
		},
	}

	svc := notification.New(smtpClient)
	err := svc.SendEmailNotification(context.Background(), "user@example.com", "hello")
	require.NoError(t, err)
	require.Equal(t, "user@example.com", sentTo)
	require.Contains(t, sentBody, "hello")
}

func TestService_SendEmailNotification_RetryAndFail(t *testing.T) {
	attempts := 0
	smtpClient := &mockSMTPClient{
		sendFn: func(context.Context, string, string) error {
			attempts++
			return errors.New("smtp down")
		},
	}

	svc := notification.New(smtpClient)
	err := svc.SendEmailNotification(context.Background(), "user@example.com", "hello")
	require.Error(t, err)
	require.True(t, errors.Is(err, errs.ErrSendFailed))
	require.Equal(t, 3, attempts)
}

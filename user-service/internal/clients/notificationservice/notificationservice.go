package notificationservice

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

const (
	verificationEmailTopic = "verification-emails"
)

type NotificationServiceInterface interface {
	SendVerificationEmail(email, token string) error
	Close() error
}

type NotificationService struct {
	writer *kafka.Writer
}

type VerificationEmail struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func NewNotificationService(brokers []string) (*NotificationService, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        verificationEmailTopic,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &NotificationService{
		writer: writer,
	}, nil
}

func (s *NotificationService) SendVerificationEmail(email, token string) error {
	msg := VerificationEmail{
		Email: email,
		Token: token,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal verification email: %w", err)
	}

	err = s.writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: payload,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (s *NotificationService) Close() error {
	return s.writer.Close()
}

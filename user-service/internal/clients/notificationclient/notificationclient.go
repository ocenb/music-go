package notificationclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

const (
	notificationEmailTopic = "email-notifications"
)

type NotificationClientInterface interface {
	SendEmailNotification(email, msg string) error
	Close() error
}

type NotificationClient struct {
	writer *kafka.Writer
}

type EmailNotification struct {
	Email string `json:"email"`
	Msg   string `json:"msg"`
}

func NewNotificationClient(brokers []string) (NotificationClientInterface, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        notificationEmailTopic,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &NotificationClient{
		writer: writer,
	}, nil
}

func (s *NotificationClient) SendEmailNotification(email, msg string) error {
	emailNotification := EmailNotification{
		Email: email,
		Msg:   msg,
	}

	payload, err := json.Marshal(emailNotification)
	if err != nil {
		return fmt.Errorf("failed to marshal email notification: %w", err)
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

func (s *NotificationClient) Close() error {
	return s.writer.Close()
}

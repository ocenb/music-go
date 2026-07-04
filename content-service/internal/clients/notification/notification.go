package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type Client struct {
	writer *kafka.Writer
}

type EmailNotification struct {
	Email string `json:"email"`
	Msg   string `json:"msg"`
}

func New(brokers []string, topic string) (*Client, error) {
	if len(brokers) == 0 {
		return nil, errors.New("kafka brokers list is empty")
	}
	if topic == "" {
		return nil, errors.New("kafka topic is empty")
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &Client{writer: writer}, nil
}

func (c *Client) SendEmailNotification(ctx context.Context, email, msg string) error {
	payload, err := json.Marshal(EmailNotification{
		Email: email,
		Msg:   msg,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal email notification: %w", err)
	}

	if err := c.writer.WriteMessages(ctx, kafka.Message{Value: payload}); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	return c.writer.Close()
}

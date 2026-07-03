package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/segmentio/kafka-go"

	"github.com/ocenb/music-go/notification-service/internal/config"
	"github.com/ocenb/music-go/notification-service/internal/errs"
	"github.com/ocenb/music-go/notification-service/internal/models"
)

type Consumer struct {
	reader *kafka.Reader
}

func New(cfg config.KafkaConfig) (*Consumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers list is empty")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	})

	return &Consumer{reader: reader}, nil
}

func (c *Consumer) Read(ctx context.Context) (*models.EmailNotification, error) {
	message, err := c.reader.ReadMessage(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read kafka message: %w", err)
	}

	var notification models.EmailNotification
	if err := json.Unmarshal(message.Value, &notification); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidNotification, err)
	}

	if notification.Email == "" || notification.Msg == "" {
		return nil, errs.ErrInvalidNotification
	}

	return &notification, nil
}

func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close kafka consumer: %w", err)
	}
	return nil
}

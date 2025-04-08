package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ocenb/music-go/notification-service/internal/config"
	"github.com/segmentio/kafka-go"
)

type EmailNotification struct {
	Email            string `json:"email"`
	VerificationLink string `json:"verification_link"`
}

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(cfg *config.Config) (*Consumer, error) {
	if len(cfg.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("kafka brokers list is empty")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
		GroupID: cfg.KafkaGroupID,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := reader.ReadMessage(ctx)
	if err != nil && err != context.DeadlineExceeded {
		return nil, fmt.Errorf("failed to connect to kafka: %w", err)
	}

	return &Consumer{
		reader: reader,
	}, nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) Consume(ctx context.Context, handler func(EmailNotification) error) error {
	for {
		message, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				return nil
			}
			log.Printf("Error reading message: %v", err)
			continue
		}

		var notification EmailNotification
		if err := json.Unmarshal(message.Value, &notification); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		if err := handler(notification); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}
}

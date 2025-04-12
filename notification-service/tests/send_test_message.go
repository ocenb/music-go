package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type EmailNotification struct {
	Email string `json:"email"`
	Msg   string `json:"msg"`
}

func main() {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"host.docker.internal:29092"},
		Topic:    "email-notifications",
		Balancer: &kafka.LeastBytes{},
	})

	notification := EmailNotification{
		Email: "test@example.com",
		Msg:   "Test message from notification service",
	}

	message, err := json.Marshal(notification)
	if err != nil {
		log.Fatalf("Failed to marshal message: %v", err)
	}

	err = w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte("test-key"),
			Value: message,
		},
	)
	if err != nil {
		log.Fatalf("Failed to write message: %v", err)
	}

	fmt.Println("Message sent successfully!")
}

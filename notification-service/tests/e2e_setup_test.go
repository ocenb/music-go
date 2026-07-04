package tests_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"

	"github.com/ocenb/music-go/notification-service/internal/config"
	"github.com/ocenb/music-go/notification-service/internal/logger"
	"github.com/ocenb/music-go/notification-service/internal/models"
	"github.com/ocenb/music-go/notification-service/internal/server"
	notificationservice "github.com/ocenb/music-go/notification-service/internal/services/notification"
	kafkastorage "github.com/ocenb/music-go/notification-service/internal/storage/kafka"
)

type testEnv struct {
	Broker string
	Topic  string
	smtp   *recordingSMTPClient
}

type recordingSMTPClient struct {
	mu       sync.Mutex
	messages []recordedEmail
	notify   chan struct{}
}

type recordedEmail struct {
	To   string
	Body string
}

func (r *recordingSMTPClient) Send(_ context.Context, to, htmlBody string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, recordedEmail{To: to, Body: htmlBody})
	select {
	case r.notify <- struct{}{}:
	default:
	}
	return nil
}

func (r *recordingSMTPClient) waitForMessage(ctx context.Context) (recordedEmail, error) {
	select {
	case <-ctx.Done():
		return recordedEmail{}, ctx.Err()
	case <-r.notify:
		r.mu.Lock()
		defer r.mu.Unlock()
		if len(r.messages) == 0 {
			return recordedEmail{}, errors.New("notification channel signaled but no messages recorded")
		}
		msg := r.messages[len(r.messages)-1]
		return msg, nil
	}
}

func setupTestEnv(ctx context.Context, t *testing.T) testEnv {
	t.Helper()

	_ = godotenv.Load("tests/.env.test")

	topic := envOr("KAFKA_TOPIC", "email-notifications")
	groupID := fmt.Sprintf("notification-e2e-%d", time.Now().UnixNano())

	kafkaContainer, err := tckafka.Run(ctx, "confluentinc/confluent-local:7.5.0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = kafkaContainer.Terminate(ctx) })

	brokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, brokers)

	broker := brokers[0]

	require.NoError(t, createTopic(ctx, broker, topic))

	kafkaCfg := config.KafkaConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	}

	consumer, err := kafkastorage.New(kafkaCfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = consumer.Close() })

	log := logger.New(-4, "text", "test")
	recorder := &recordingSMTPClient{notify: make(chan struct{}, 1)}
	svc := notificationservice.New(recorder)
	srv := server.New(svc, consumer, log)

	runCtx, cancel := context.WithCancel(ctx)
	serverReady := make(chan struct{})
	go func() {
		close(serverReady)
		_ = srv.Start(runCtx)
	}()
	<-serverReady
	time.Sleep(100 * time.Millisecond)

	t.Cleanup(func() {
		cancel()
		_ = srv.Stop()
	})

	return testEnv{
		Broker: broker,
		Topic:  topic,
		smtp:   recorder,
	}
}

func publishNotification(ctx context.Context, t *testing.T, broker, topic, email, msg string) {
	t.Helper()

	writer := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		RequiredAcks: kafka.RequireAll,
	}
	t.Cleanup(func() { _ = writer.Close() })

	payload, err := json.Marshal(models.EmailNotification{Email: email, Msg: msg})
	require.NoError(t, err)

	err = writer.WriteMessages(ctx, kafka.Message{Value: payload})
	require.NoError(t, err)
}

func createTopic(ctx context.Context, broker, topic string) error {
	conn, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

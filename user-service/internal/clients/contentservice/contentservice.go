package contentservice

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ocenb/music-go/user-service/internal/config"
)

type ContentServiceClient struct {
	cfg    *config.Config
	client *http.Client
	log    *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) *ContentServiceClient {
	return &ContentServiceClient{
		cfg:    cfg,
		client: &http.Client{},
		log:    log,
	}
}

func (c *ContentServiceClient) DeleteUserContent(ctx context.Context, userID int64) error {
	url := fmt.Sprintf("%s/all", c.cfg.ContentServiceURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		c.log.Error("Failed to create request to content service", "error", err, "user_id", userID)
		return fmt.Errorf("failed to create request to content service: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("Failed to delete user content", "error", err, "user_id", userID)
		return fmt.Errorf("failed to delete user content: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			c.log.Error("Failed to close response body", "error", err, "user_id", userID)
		}
	}()

	if resp.StatusCode != http.StatusNoContent {
		c.log.Error("Content service returned non-success status code", "status_code", resp.StatusCode, "user_id", userID)
		return fmt.Errorf("content service returned status code %d", resp.StatusCode)
	}

	c.log.Info("Successfully deleted user content", "user_id", userID)
	return nil
}

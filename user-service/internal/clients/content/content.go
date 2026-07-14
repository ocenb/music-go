package content

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const internalSecretHeader = "X-Internal-Secret" //nolint:gosec // HTTP header name, not a credential

type Client struct {
	baseURL        string
	internalSecret string
	client         *http.Client
}

func New(baseURL, internalSecret string, timeout time.Duration) *Client {
	return &Client{
		baseURL:        strings.TrimSuffix(baseURL, "/"),
		internalSecret: internalSecret,
		client:         &http.Client{Timeout: timeout},
	}
}

func (c *Client) DeleteUserContent(ctx context.Context, userID int64) error {
	url := fmt.Sprintf("%s/api/content/internal/users/%d", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request to content service: %w", err)
	}
	req.Header.Set(internalSecretHeader, c.internalSecret)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete user content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("content service returned status code %d", resp.StatusCode)
	}

	return nil
}

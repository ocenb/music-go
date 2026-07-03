package elastic

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"

	"github.com/ocenb/music-go/search-service/internal/config"
)

const (
	UsersIndexName  = "users"
	AlbumsIndexName = "albums"
	TracksIndexName = "tracks"
)

type Client struct {
	typed *elasticsearch.TypedClient
}

func New(ctx context.Context, cfg config.ElasticConfig, log *slog.Logger) (*Client, error) {
	address := fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port)

	es, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{address},
		Username:  cfg.User,
		Password:  cfg.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	_, err = es.Info().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}

	client := &Client{typed: es}

	if err := client.createIndices(ctx); err != nil {
		return nil, fmt.Errorf("failed to create indices: %w", err)
	}

	log.Info("connected to elasticsearch", slog.String("address", address))

	return client, nil
}

func (c *Client) Typed() *elasticsearch.TypedClient {
	return c.typed
}

func (c *Client) createIndices(ctx context.Context) error {
	if err := c.createIndex(ctx, UsersIndexName, map[string]types.Property{
		"id":       types.NewKeywordProperty(),
		"username": types.NewTextProperty(),
	}); err != nil {
		return err
	}

	if err := c.createIndex(ctx, AlbumsIndexName, map[string]types.Property{
		"id":    types.NewKeywordProperty(),
		"title": types.NewTextProperty(),
	}); err != nil {
		return err
	}

	if err := c.createIndex(ctx, TracksIndexName, map[string]types.Property{
		"id":    types.NewKeywordProperty(),
		"title": types.NewTextProperty(),
	}); err != nil {
		return err
	}

	return nil
}

func (c *Client) createIndex(ctx context.Context, name string, properties map[string]types.Property) error {
	exists, err := c.typed.Indices.Exists(name).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if %s index exists: %w", name, err)
	}
	if exists {
		return nil
	}

	mapping := types.TypeMapping{Properties: properties}
	_, err = c.typed.Indices.Create(name).
		Request(&create.Request{Mappings: &mapping}).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create %s index: %w", name, err)
	}

	return nil
}

package search

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/ocenb/music-protos/gen/searchservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client searchservice.SearchServiceClient
	conn   *grpc.ClientConn
}

func New(address string, requestTimeout time.Duration, log *slog.Logger) (*Client, error) {
	log.Info("creating search service client", slog.String("address", address))

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(timeout.UnaryClientInterceptor(requestTimeout)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create search service gRPC client: %w", err)
	}

	return &Client{
		client: searchservice.NewSearchServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) AddTrack(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	return c.client.AddTrack(ctx, req)
}

func (c *Client) UpdateTrack(ctx context.Context, req *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	return c.client.UpdateTrack(ctx, req)
}

func (c *Client) DeleteTrack(ctx context.Context, req *searchservice.DeleteRequest) (*searchservice.SuccessResponse, error) {
	return c.client.DeleteTrack(ctx, req)
}

func (c *Client) SearchUsers(ctx context.Context, req *searchservice.SearchRequest) (*searchservice.SearchResponse, error) {
	return c.client.SearchUsers(ctx, req)
}

func (c *Client) SearchTracks(ctx context.Context, req *searchservice.SearchRequest) (*searchservice.SearchResponse, error) {
	return c.client.SearchTracks(ctx, req)
}

func (c *Client) Close() error {
	return c.conn.Close()
}

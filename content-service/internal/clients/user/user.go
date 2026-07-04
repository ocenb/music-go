package user

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/ocenb/music-protos/gen/userservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Client struct {
	client userservice.UserServiceClient
	conn   *grpc.ClientConn
}

func New(address string, requestTimeout time.Duration, log *slog.Logger) (*Client, error) {
	log.Info("creating user service client", slog.String("address", address))

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(timeout.UnaryClientInterceptor(requestTimeout)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user service gRPC client: %w", err)
	}

	return &Client{
		client: userservice.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) CheckAuth(ctx context.Context, authorizationHeader string) (*userservice.CheckAuthResponse, error) {
	outCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", authorizationHeader))

	res, err := c.client.CheckAuth(outCtx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("user service check auth: %w", err)
	}
	return res, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

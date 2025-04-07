package userservice

import (
	"fmt"
	"net"
	"time"

	"github.com/ocenb/music-go/search-service/internal/config"
	"github.com/ocenb/music-protos/gen/userservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserServiceClient struct {
	Client userservice.UserServiceClient
	Conn   *grpc.ClientConn
}

func New(cfg *config.Config) (*UserServiceClient, error) {
	tcpConn, tcpErr := net.DialTimeout("tcp", cfg.UserServiceAddress, 5*time.Second)
	if tcpErr != nil {
		fmt.Printf("TCP connection test failed: %v\n", tcpErr)
	} else {
		err := tcpConn.Close()
		if err != nil {
			fmt.Printf("TCP connection close failed: %v\n", err)
		}
		fmt.Println("TCP connection test successful")
	}

	conn, err := grpc.NewClient(
		cfg.UserServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	if tcpErr == nil {
		fmt.Println("gRPC connection created successfully")
	}

	return &UserServiceClient{
		Client: userservice.NewUserServiceClient(conn),
		Conn:   conn,
	}, nil
}

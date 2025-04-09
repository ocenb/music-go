package searchclient

import (
	"fmt"
	"net"
	"time"

	"github.com/ocenb/music-go/content-service/internal/config"
	"github.com/ocenb/music-protos/gen/searchservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SearchServiceClient struct {
	Client searchservice.SearchServiceClient
	Conn   *grpc.ClientConn
}

func New(cfg *config.Config) (*SearchServiceClient, error) {
	tcpConn, tcpErr := net.DialTimeout("tcp", cfg.SearchServiceAddress, 5*time.Second)
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
		cfg.SearchServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	if tcpErr == nil {
		fmt.Println("gRPC connection created successfully")
	}

	return &SearchServiceClient{
		Client: searchservice.NewSearchServiceClient(conn),
		Conn:   conn,
	}, nil
}

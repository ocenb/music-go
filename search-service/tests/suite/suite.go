package suite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ocenb/music-protos/gen/searchservice"
	"github.com/ocenb/music-protos/gen/userservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	AdminEmail    = "admin@example.com"
	AdminPassword = "Password123!"
)

type Suite struct {
	*testing.T
	SearchClient searchservice.SearchServiceClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*1)

	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	userServiceConn, err := grpc.NewClient(
		"localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc server connection error: %v", err)
	}
	userClient := userservice.NewUserServiceClient(userServiceConn)

	loginResp, err := userClient.Login(ctx, &userservice.LoginRequest{
		Email:    AdminEmail,
		Password: AdminPassword,
	})
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	outMD := metadata.New(map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", loginResp.AccessToken),
	})
	outCtx := metadata.NewOutgoingContext(ctx, outMD)

	searchServiceConn, err := grpc.NewClient(
		"localhost:9091",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc server connection error: %v", err)
	}

	return outCtx, &Suite{
		T:            t,
		SearchClient: searchservice.NewSearchServiceClient(searchServiceConn),
	}
}

package suite

import (
	"context"
	"testing"
	"time"

	"github.com/ocenb/music-protos/gen/userservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	UserClient userservice.UserServiceClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*1)

	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	conn, err := grpc.NewClient(
		"localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc server connection error: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		UserClient: userservice.NewUserServiceClient(conn),
	}
}

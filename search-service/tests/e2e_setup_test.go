package tests_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	protos "github.com/ocenb/music-protos/gen/searchservice"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/ocenb/music-go/search-service/internal/config"
	"github.com/ocenb/music-go/search-service/internal/logger"
	searchrepo "github.com/ocenb/music-go/search-service/internal/repos/search"
	"github.com/ocenb/music-go/search-service/internal/server"
	searchsvc "github.com/ocenb/music-go/search-service/internal/services/search"
	"github.com/ocenb/music-go/search-service/internal/storage/elastic"
)

type testEnv struct {
	Client   protos.SearchServiceClient
	grpcConn *grpc.ClientConn
}

type noopUserClient struct{}

func (noopUserClient) CheckAuth(context.Context, string) error { return nil }

func setupTestEnv(ctx context.Context, t *testing.T) testEnv {
	t.Helper()

	_ = godotenv.Load("tests/.env.test")

	esContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "docker.elastic.co/elasticsearch/elasticsearch:8.12.0",
			ExposedPorts: []string{"9200/tcp"},
			Env: map[string]string{
				"discovery.type":         "single-node",
				"xpack.security.enabled": "false",
				"ES_JAVA_OPTS":           "-Xms512m -Xmx512m",
			},
			WaitingFor: wait.ForHTTP("/").WithPort("9200/tcp").WithStartupTimeout(90 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(esContainer) })

	host, err := esContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := esContainer.MappedPort(ctx, "9200/tcp")
	require.NoError(t, err)

	elasticCfg := config.ElasticConfig{
		Host:     host,
		Port:     mappedPort.Port(),
		User:     envOr("ELASTIC_USER", "elastic"),
		Password: envOr("ELASTIC_PASSWORD", ""),
	}

	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log := logger.New(-4, "text", "test")
	elasticClient, err := elastic.New(connectCtx, elasticCfg, log)
	require.NoError(t, err)

	repo := searchrepo.New(elasticClient)
	svc := searchsvc.New(repo)

	port := pickFreePort(t)
	srv := server.New(svc, noopUserClient{}, log, port)

	serverReady := make(chan struct{})
	go func() {
		close(serverReady)
		_ = srv.Start()
	}()
	<-serverReady
	time.Sleep(100 * time.Millisecond)

	t.Cleanup(func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = srv.Stop(shutdownCtx)
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return testEnv{
		Client:   protos.NewSearchServiceClient(conn),
		grpcConn: conn,
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func pickFreePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

func authContext(ctx context.Context) context.Context {
	md := metadata.Pairs("authorization", "Bearer test-token")
	return metadata.NewOutgoingContext(ctx, md)
}

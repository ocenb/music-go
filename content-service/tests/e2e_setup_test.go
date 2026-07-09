package tests_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/ocenb/music-protos/gen/searchservice"
	"github.com/ocenb/music-protos/gen/userservice"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ocenb/music-go/content-service/internal/config"
	"github.com/ocenb/music-go/content-service/internal/logger"
	allrepo "github.com/ocenb/music-go/content-service/internal/repos/all"
	historyrepo "github.com/ocenb/music-go/content-service/internal/repos/history"
	playlistrepo "github.com/ocenb/music-go/content-service/internal/repos/playlist"
	playlisttracksrepo "github.com/ocenb/music-go/content-service/internal/repos/playlisttracks"
	trackrepo "github.com/ocenb/music-go/content-service/internal/repos/track"
	"github.com/ocenb/music-go/content-service/internal/server"
	allservice "github.com/ocenb/music-go/content-service/internal/services/all"
	fileservice "github.com/ocenb/music-go/content-service/internal/services/file"
	historyservice "github.com/ocenb/music-go/content-service/internal/services/history"
	playlistservice "github.com/ocenb/music-go/content-service/internal/services/playlist"
	playlisttracksservice "github.com/ocenb/music-go/content-service/internal/services/playlisttracks"
	searchsvc "github.com/ocenb/music-go/content-service/internal/services/search"
	trackservice "github.com/ocenb/music-go/content-service/internal/services/track"
	"github.com/ocenb/music-go/content-service/internal/storage/migrator"
	postgresstorage "github.com/ocenb/music-go/content-service/internal/storage/postgres"
	"github.com/ocenb/music-go/content-service/internal/storage/transactor"
	"github.com/ocenb/music-go/content-service/migrations"
)

const (
	testInternalSecret = "test-internal-secret"
	testUserID         = int64(1)
	testUsername       = "testuser"
)

type testEnv struct {
	BaseURL    string
	HTTPClient *http.Client
}

type mockUserClient struct{}

func (mockUserClient) CheckAuth(_ context.Context, _ string) (*userservice.CheckAuthResponse, error) {
	return &userservice.CheckAuthResponse{
		User: &userservice.UserPrivateModel{
			Id:       testUserID,
			Username: testUsername,
			Email:    "test@example.com",
		},
	}, nil
}

type noopCloudinary struct{}

func (noopCloudinary) Upload(_ context.Context, _, _, _, _ string) error { return nil }
func (noopCloudinary) Delete(_ context.Context, _, _ string) error       { return nil }

type noopSearchClient struct{}

func (noopSearchClient) AddTrack(_ context.Context, _ *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	return &searchservice.SuccessResponse{}, nil
}
func (noopSearchClient) UpdateTrack(_ context.Context, _ *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	return &searchservice.SuccessResponse{}, nil
}
func (noopSearchClient) DeleteTrack(_ context.Context, _ *searchservice.DeleteRequest) (*searchservice.SuccessResponse, error) {
	return &searchservice.SuccessResponse{}, nil
}
func (noopSearchClient) SearchUsers(_ context.Context, _ *searchservice.SearchRequest) (*searchservice.SearchResponse, error) {
	return &searchservice.SearchResponse{}, nil
}
func (noopSearchClient) SearchTracks(_ context.Context, _ *searchservice.SearchRequest) (*searchservice.SearchResponse, error) {
	return &searchservice.SearchResponse{}, nil
}
func (noopSearchClient) Close() error { return nil }

type noopNotificationClient struct{}

func (noopNotificationClient) SendEmailNotification(_ context.Context, _, _ string) error { return nil }
func (noopNotificationClient) Close() error                                               { return nil }

func setupTestEnv(ctx context.Context, t *testing.T) testEnv {
	t.Helper()

	_ = godotenv.Load("tests/.env.test")

	pgUser := envOr("POSTGRES_USER", "postgres")
	pgPassword := envOr("POSTGRES_PASSWORD", "postgres")
	pgDB := envOr("POSTGRES_DB", "content-service-db")

	pgContainer, err := postgres.Run(ctx,
		"postgres:18.3-alpine3.23",
		postgres.WithDatabase(pgDB),
		postgres.WithUsername(pgUser),
		postgres.WithPassword(pgPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(pgContainer) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	mgr, err := migrator.New(ctx, connStr, migrations.FS, goose.NopLogger())
	require.NoError(t, err)
	require.NoError(t, mgr.Up(ctx))
	require.NoError(t, mgr.Close())

	poolCfg := config.PostgresConfig{
		DSN:             connStr,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Hour,
	}
	pool, err := postgresstorage.NewPool(ctx, poolCfg)
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	tm := transactor.New(pool)
	log := logger.New(-4, "text", "test")

	cfg := &config.Config{
		Environment:           "test",
		InternalServiceSecret: testInternalSecret,
		ImageFileLimit:        10 * 1024 * 1024,
		AudioFileLimit:        50 * 1024 * 1024,
		HTTP: config.HTTPConfig{
			Port:         0,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}

	var cloudinaryClient fileservice.CloudinaryClient = noopCloudinary{}
	fileSvc := fileservice.New(cloudinaryClient, log, cfg)
	trackRepo := trackrepo.New(tm)
	playlistRepo := playlistrepo.New(tm)
	playlistTracksRepo := playlisttracksrepo.New(tm)
	historyRepo := historyrepo.New(tm)
	allRepo := allrepo.New(tm)

	var searchClient noopSearchClient
	var notificationClient noopNotificationClient

	trackSvc := trackservice.New(trackRepo, fileSvc, searchClient, notificationClient, tm)
	playlistSvc := playlistservice.New(playlistRepo, fileSvc, tm)
	playlistTracksSvc := playlisttracksservice.New(playlistTracksRepo, playlistRepo, trackRepo, tm)
	historySvc := historyservice.New(historyRepo, trackSvc)
	allSvc := allservice.New(allRepo, fileSvc)
	searchSvc := searchsvc.New(searchClient)

	port := freePort(t)
	cfg.HTTP.Port = port

	srv := server.New(
		trackSvc,
		playlistSvc,
		playlistTracksSvc,
		historySvc,
		searchSvc,
		allSvc,
		mockUserClient{},
		testInternalSecret,
		cfg,
		log,
	)

	go func() {
		_ = srv.Start()
	}()
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Stop(shutdownCtx)
	})

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	require.Eventually(t, func() bool {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 10*time.Second, 100*time.Millisecond)

	return testEnv{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

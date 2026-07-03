package tests_test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/ocenb/music-protos/gen/searchservice"
	"github.com/ocenb/music-protos/gen/userservice"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/ocenb/music-go/user-service/internal/config"
	"github.com/ocenb/music-go/user-service/internal/logger"
	tokenrepo "github.com/ocenb/music-go/user-service/internal/repos/token"
	userrepo "github.com/ocenb/music-go/user-service/internal/repos/user"
	"github.com/ocenb/music-go/user-service/internal/server"
	authservice "github.com/ocenb/music-go/user-service/internal/services/auth"
	tokenservice "github.com/ocenb/music-go/user-service/internal/services/token"
	usersvc "github.com/ocenb/music-go/user-service/internal/services/user"
	"github.com/ocenb/music-go/user-service/internal/storage/migrator"
	postgresstorage "github.com/ocenb/music-go/user-service/internal/storage/postgres"
	"github.com/ocenb/music-go/user-service/internal/storage/transactor"
	"github.com/ocenb/music-go/user-service/migrations"
)

const (
	adminUsername      = "admin"
	adminEmail         = "admin@example.com"
	adminPassword      = "Password123!"
	toChangeUsername   = "tochange"
	toChangeEmail      = "tochange@example.com"
	toChangePassword   = "Password123!"
	changeNameEmail    = "changename@example.com"
	changeNamePassword = "Password123!"
	toFollowUsername   = "tofollow"
	toFollowEmail      = "tofollow@example.com"
	toFollowPassword   = "Password123!"
	toDeleteUsername   = "todelete"
	toDeleteEmail      = "todelete@example.com"
	toDeletePassword   = "Password123!"
)

type testEnv struct {
	Client       userservice.UserServiceClient
	DBConnString string
	grpcConn     *grpc.ClientConn
}

type testEnvOptions struct {
	SeedFile string
}

func setupTestEnv(ctx context.Context, t *testing.T) testEnv {
	return setupTestEnvWithSeed(ctx, t, "seed_users.sql")
}

func setupTestEnvWithSeed(ctx context.Context, t *testing.T, seedFile string) testEnv {
	return setupTestEnvWithOptions(ctx, t, testEnvOptions{SeedFile: seedFile})
}

func setupTestEnvWithOptions(ctx context.Context, t *testing.T, opts testEnvOptions) testEnv {
	t.Helper()

	_ = godotenv.Load(".env.test")

	if opts.SeedFile == "" {
		opts.SeedFile = "seed_users.sql"
	}

	pgUser := envOr("POSTGRES_USER", "postgres")
	pgPassword := envOr("POSTGRES_PASSWORD", "postgres")
	pgDB := envOr("POSTGRES_DB", "user-service-db")

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
	require.NoError(t, mgr.Up())
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

	seedBytes, err := readSeedFile(opts.SeedFile)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, string(seedBytes))
	require.NoError(t, err)

	port := pickFreePort(t)
	log := logger.New(-4, "text", "test")

	tm := transactor.New(pool)
	tokenRepo := tokenrepo.New(tm)
	userRepo := userrepo.New(tm)

	jwtSecret := envOr("JWT_SECRET", "test-jwt-secret-for-e2e")
	bcryptCost := 12

	tokenService := tokenservice.New(tokenRepo, jwtSecret, time.Hour, 30*24*time.Hour)
	userService := usersvc.New(userRepo, tm, noopSearchClient{}, noopContentClient{})
	authService := authservice.New(userService, tokenService, noopNotificationClient{}, tm, bcryptCost)

	srv := server.New(authService, userService, log, port)

	serverReady := make(chan struct{})
	go func() {
		close(serverReady)
		_ = srv.Start()
	}()
	<-serverReady
	time.Sleep(100 * time.Millisecond)

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Stop(shutdownCtx)
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return testEnv{
		Client:       userservice.NewUserServiceClient(conn),
		DBConnString: connStr,
		grpcConn:     conn,
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func readSeedFile(name string) ([]byte, error) {
	switch name {
	case "seed_users.sql":
		return os.ReadFile("seed_users.sql")
	default:
		return nil, fmt.Errorf("unsupported seed file: %s", name)
	}
}

func pickFreePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

func authContext(ctx context.Context, accessToken string) context.Context {
	md := metadata.Pairs("authorization", "Bearer "+accessToken)
	return metadata.NewOutgoingContext(ctx, md)
}

func login(ctx context.Context, t *testing.T, client userservice.UserServiceClient, email, password string) *userservice.LoginResponse {
	t.Helper()

	resp, err := client.Login(ctx, &userservice.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.AccessToken)
	require.NotEmpty(t, resp.RefreshToken)

	return resp
}

func logout(ctx context.Context, t *testing.T, client userservice.UserServiceClient, accessToken string) {
	t.Helper()

	resp, err := client.Logout(authContext(ctx, accessToken), &emptypb.Empty{})
	require.NoError(t, err)
	require.True(t, resp.Success)
}

func init() {
	if err := gofakeit.Seed(0); err != nil {
		panic(err)
	}
}

func validUsername() string {
	starters := "abcdefghijklmnopqrstuvwxyz0123456789"
	chars := "abcdefghijklmnopqrstuvwxyz0123456789_-"
	length := rand.Intn(13) + 3

	var b strings.Builder
	b.WriteByte(starters[rand.Intn(len(starters))])
	for i := 1; i < length; i++ {
		b.WriteByte(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func validPassword() string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*?-"
	length := rand.Intn(9) + 8

	var b strings.Builder
	for range length {
		b.WriteByte(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func fakeRegisterRequest() (string, string, string) {
	return validUsername(), gofakeit.Email(), validPassword()
}

func fakeRefreshToken() string {
	return uuid.NewString() + uuid.NewString()
}

func fakeAccessToken() string {
	claims := jwt.MapClaims{
		"userId":  gofakeit.Int64(),
		"tokenId": uuid.NewString(),
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("wrong-secret"))
	return tokenString
}

type noopSearchClient struct{}

func (noopSearchClient) AddUser(context.Context, *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	return &searchservice.SuccessResponse{Success: true}, nil
}

func (noopSearchClient) UpdateUser(context.Context, *searchservice.AddOrUpdateRequest) (*searchservice.SuccessResponse, error) {
	return &searchservice.SuccessResponse{Success: true}, nil
}

func (noopSearchClient) DeleteUser(context.Context, *searchservice.DeleteRequest) (*searchservice.SuccessResponse, error) {
	return &searchservice.SuccessResponse{Success: true}, nil
}

type noopContentClient struct{}

func (noopContentClient) DeleteUserContent(context.Context, int64) error {
	return nil
}

type noopNotificationClient struct{}

func (noopNotificationClient) SendEmailNotification(context.Context, string, string) error {
	return nil
}

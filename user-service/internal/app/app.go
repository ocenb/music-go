package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/bufbuild/protovalidate-go"
	authmw "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	protovalidatemw "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/ocenb/music-go/user-service/internal/config"
	"github.com/ocenb/music-go/user-service/internal/handlers"
	"github.com/ocenb/music-go/user-service/internal/services/auth"
	"github.com/ocenb/music-go/user-service/internal/services/user"
	"github.com/ocenb/music-go/user-service/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type App struct {
	gRPCServer *grpc.Server
	port       int
	log        *slog.Logger
}

func New(
	authService auth.AuthServiceInterface,
	userService user.UserServiceInterface,
	port int,
	cfg *config.Config,
	log *slog.Logger,
) *App {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	validator, err := protovalidate.New()
	if err != nil {
		panic(fmt.Errorf("protovalidate error: %w", err))
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(loggerInterceptor(log), loggingOpts...),
		protovalidatemw.UnaryServerInterceptor(validator),
		authmw.UnaryServerInterceptor(authFunc(authService)),
	))

	handlers.NewUserServer(gRPCServer, cfg, authService, userService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func loggerInterceptor(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func checkAuth(ctx context.Context) bool {
	noAuthMethods := map[string]bool{
		"/userservice.UserService/Register":        true,
		"/userservice.UserService/Login":           true,
		"/userservice.UserService/Refresh":         true,
		"/userservice.UserService/Verify":          true,
		"/userservice.UserService/NewVerification": true,
	}

	fullMethod, ok := grpc.Method(ctx)

	if ok && noAuthMethods[fullMethod] {
		return false
	}
	return true
}

func authFunc(authService auth.AuthServiceInterface) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		if !checkAuth(ctx) {
			return ctx, nil
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		authHeader, ok := md["authorization"]
		if !ok || len(authHeader) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization header is not provided")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		if token == authHeader[0] {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header format")
		}

		user, tokenId, err := authService.ValidateAccessToken(ctx, token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "%v", err)
		}

		newCtx := context.WithValue(context.WithValue(ctx, utils.UserKey{}, user), utils.TokenIdKey{}, tokenId)

		return newCtx, nil
	}
}

func (a *App) Run() {
	const op = "app.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	err = a.gRPCServer.Serve(l)
	if err != nil {
		panic(fmt.Errorf("%s: %w", op, err))
	}
}

func (a *App) Stop() {
	const op = "app.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}

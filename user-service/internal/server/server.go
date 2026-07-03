package server

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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ocenb/music-go/user-service/internal/errs"
	"github.com/ocenb/music-go/user-service/internal/handlers"
	"github.com/ocenb/music-go/user-service/internal/logger"
	"github.com/ocenb/music-go/user-service/internal/middlewares"
)

var noAuthMethods = map[string]struct{}{
	"/userservice.UserService/Register":        {},
	"/userservice.UserService/Login":           {},
	"/userservice.UserService/Refresh":         {},
	"/userservice.UserService/Verify":          {},
	"/userservice.UserService/NewVerification": {},
}

type Server struct {
	gRPCServer *grpc.Server
	port       int
	log        *slog.Logger
}

func New(
	authService handlers.AuthService,
	userService handlers.UserService,
	log *slog.Logger,
	port int,
) *Server {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
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
		middlewares.LoggingUnaryInterceptor(log),
		logging.UnaryServerInterceptor(loggerInterceptor(), loggingOpts...),
		protovalidatemw.UnaryServerInterceptor(validator),
		authmw.UnaryServerInterceptor(authFunc(authService)),
	))

	handlers.NewUserServer(gRPCServer, authService, userService)

	return &Server{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func loggerInterceptor() logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		logger.FromContext(ctx).Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func checkAuth(ctx context.Context) bool {
	fullMethod, ok := grpc.Method(ctx)
	if !ok {
		return true
	}
	_, skip := noAuthMethods[fullMethod]
	return !skip
}

func authFunc(authService handlers.AuthService) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		if !checkAuth(ctx) {
			return ctx, nil
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errs.ToGRPC(errs.Unauthenticated("metadata is not provided"))
		}

		authHeader, ok := md["authorization"]
		if !ok || len(authHeader) == 0 {
			return nil, errs.ToGRPC(errs.Unauthenticated("authorization header is not provided"))
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		if token == authHeader[0] {
			return nil, errs.ToGRPC(errs.Unauthenticated("invalid authorization header format"))
		}

		user, tokenID, err := authService.ValidateAccessToken(ctx, token)
		if err != nil {
			if domainErr, ok := errs.As(err); ok && domainErr.Kind() == errs.KindUnauthenticated {
				return nil, errs.ToGRPC(err)
			}
			return nil, errs.ToGRPC(errs.Unauthenticated("unauthenticated"))
		}

		newCtx := middlewares.ContextWithAuth(ctx, user, tokenID)

		return newCtx, nil
	}
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}

	s.log.Info("gRPC server started", slog.String("addr", l.Addr().String()))

	err = s.gRPCServer.Serve(l)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("stopping gRPC server")

	stopped := make(chan struct{})
	go func() {
		s.gRPCServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		s.gRPCServer.Stop()
		return ctx.Err()
	}
}

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	authmw "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ocenb/music-go/search-service/internal/errs"
	"github.com/ocenb/music-go/search-service/internal/handlers"
	"github.com/ocenb/music-go/search-service/internal/logger"
	"github.com/ocenb/music-go/search-service/internal/middlewares"
)

var authRequiredMethods = map[string]struct{}{
	"/searchservice.SearchService/SearchUsers":  {},
	"/searchservice.SearchService/SearchAlbums": {},
	"/searchservice.SearchService/SearchTracks": {},
}

type UserClient interface {
	CheckAuth(ctx context.Context, authorizationHeader string) error
}

type Server struct {
	gRPCServer *grpc.Server
	port       int
	log        *slog.Logger
}

func New(
	searchService handlers.SearchService,
	userClient UserClient,
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

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		middlewares.LoggingUnaryInterceptor(log),
		logging.UnaryServerInterceptor(loggerInterceptor(), loggingOpts...),
		authmw.UnaryServerInterceptor(authFunc(userClient)),
	))

	handlers.NewSearchServer(gRPCServer, searchService)

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

func requiresAuth(ctx context.Context) bool {
	fullMethod, ok := grpc.Method(ctx)
	if !ok {
		return false
	}
	_, required := authRequiredMethods[fullMethod]
	return required
}

func authFunc(userClient UserClient) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		if !requiresAuth(ctx) {
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

		if err := userClient.CheckAuth(ctx, authHeader[0]); err != nil {
			return nil, errs.ToGRPC(errs.Unauthenticated("unauthenticated"))
		}

		return ctx, nil
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

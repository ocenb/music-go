package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	authmw "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/ocenb/music-go/search-service/internal/clients/userservice"
	"github.com/ocenb/music-go/search-service/internal/config"
	"github.com/ocenb/music-go/search-service/internal/handlers"
	"github.com/ocenb/music-go/search-service/internal/service"
	"github.com/ocenb/music-go/search-service/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type App struct {
	gRPCServer *grpc.Server
	port       int
	log        *slog.Logger
}

func New(
	searchService service.SearchServiceInterface,
	userServiceClient *userservice.UserServiceClient,
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

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(loggerInterceptor(log), loggingOpts...),
		authmw.UnaryServerInterceptor(authFunc(userServiceClient)),
	))

	handlers.NewSearchServer(gRPCServer, cfg, searchService, log)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       cfg.GRPC.Port,
	}
}

func loggerInterceptor(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func checkAuth(ctx context.Context) bool {
	authMethods := map[string]bool{
		"/searchservice.SearchService/SearchUsers":  true,
		"/searchservice.SearchService/SearchAlbums": true,
		"/searchservice.SearchService/SearchTracks": true,
	}
	fullMethod, ok := grpc.Method(ctx)

	if ok && authMethods[fullMethod] {
		return true
	}
	return false
}

func authFunc(userServiceClient *userservice.UserServiceClient) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		if !checkAuth(ctx) {
			return ctx, nil
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, utils.UnauthenticatedError("metadata is not provided")
		}

		authHeader, ok := md["authorization"]
		if !ok || len(authHeader) == 0 {
			return nil, utils.UnauthenticatedError("authorization header is not provided")
		}

		outMD := metadata.New(map[string]string{
			"authorization": authHeader[0],
		})
		outCtx := metadata.NewOutgoingContext(ctx, outMD)

		_, err := userServiceClient.Client.CheckAuth(outCtx, &emptypb.Empty{})
		if err != nil {
			return nil, utils.UnauthenticatedError(err.Error())
		}
		return ctx, nil
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

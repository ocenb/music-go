package middlewares

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/ocenb/music-go/user-service/internal/logger"
)

func LoggingUnaryInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = logger.IntoContext(ctx, log.With(slog.String("grpc.method", info.FullMethod)))
		return handler(ctx, req)
	}
}

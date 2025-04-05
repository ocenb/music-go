package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/ocenb/music-go/user-service/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserKey struct{}
type TokenIdKey struct{}
type TxKey struct{}

func ErrLog(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func GetInfoFromContext(ctx context.Context) (*models.UserFullModel, string, error) {
	user, ok := ctx.Value(UserKey{}).(*models.UserFullModel)
	if !ok {
		return nil, "", status.Errorf(codes.Internal, "failed to get user from context")
	}
	tokenId, ok := ctx.Value(TokenIdKey{}).(string)
	if !ok {
		return nil, "", status.Errorf(codes.Internal, "failed to get tokenId from context")
	}
	return user, tokenId, nil
}

func GetTxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(TxKey{}).(*sql.Tx)
	return tx, ok
}

func GetDBUrl(DBHost, DBPort, DBUser, DBPassword, DBName, DBSSLMode string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		DBUser,
		DBPassword,
		DBHost,
		DBPort,
		DBName,
		DBSSLMode,
	)
}

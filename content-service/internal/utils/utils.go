package utils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/ocenb/music-protos/gen/userservice"
)

type txKey struct{}

func GetInfoFromContext(c *gin.Context) (*userservice.UserPrivateModel, error) {
	user, ok := c.Value("user").(*userservice.UserPrivateModel)
	if !ok {
		return nil, errors.New("failed to get user from context")
	}
	return user, nil
}

func GetTxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}

func SetTxToContext(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func GetPostgresUrl(host, port, user, password, dbName, sslMode string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user,
		password,
		host,
		port,
		dbName,
		sslMode,
	)
}

func GetRedisUrl(host, port string) string {
	return fmt.Sprintf("%s:%s", host, port)
}

func ErrLog(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

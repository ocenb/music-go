package utils

import (
	"fmt"
	"log/slog"
)

func ErrLog(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func GetElasticUrl(ElasticHost, ElasticPort, ElasticUser, ElasticPassword string) string {
	return fmt.Sprintf(
		"http://%s:%s@%s:%s",
		ElasticUser,
		ElasticPassword,
		ElasticHost,
		ElasticPort,
	)
}

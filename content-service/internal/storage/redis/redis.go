package redis

import (
	"context"

	"github.com/ocenb/music-go/content-service/internal/config"

	"github.com/redis/go-redis/v9"
)

func New(cfg *config.Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisUrl,
		Password: cfg.RedisPassword,
		DB:       0,
		Protocol: 2,
	})

	ctx := context.Background()
	err := db.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return db, nil
}

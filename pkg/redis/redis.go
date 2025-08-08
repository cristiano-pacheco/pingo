package redis

import (
	"context"
	"errors"
	"log/slog"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type Redis interface {
	Client() *redis.Client
}

type redisClient struct {
	redisClient *redis.Client
}

func NewRedis(addr string, password string, db int) Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	err := errors.Join(redisotel.InstrumentTracing(rdb), redisotel.InstrumentMetrics(rdb))
	if err != nil {
		slog.Error("failed to instrument tracing", "error", err) //nolint:sloglint // no lint is required here
		panic(err)
	}

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		slog.Error("failed to ping redis", "error", err) //nolint:sloglint // no lint is required here
		panic(err)
	}

	return &redisClient{redisClient: rdb}
}

func (r *redisClient) Client() *redis.Client {
	return r.redisClient
}

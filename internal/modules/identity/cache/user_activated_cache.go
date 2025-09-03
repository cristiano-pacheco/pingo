package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cristiano-pacheco/pingo/pkg/redis"
	redisClient "github.com/redis/go-redis/v9"
)

const (
	cacheKeyPrefix = "user_activated:"
	cacheTTL       = 24 * time.Hour
)

type UserActivatedCache interface {
	Set(userID uint64) error
	Get(userID uint64) (bool, error)
	Delete(userID uint64) error
}

type userActivatedCache struct {
	redisClient redis.Redis
}

func NewUserActivatedCache(redisClient redis.Redis) UserActivatedCache {
	return &userActivatedCache{
		redisClient: redisClient,
	}
}

func (c *userActivatedCache) Set(userID uint64) error {
	key := c.buildKey(userID)
	ctx := context.Background()

	return c.redisClient.Client().Set(ctx, key, "1", cacheTTL).Err()
}

func (c *userActivatedCache) Get(userID uint64) (bool, error) {
	key := c.buildKey(userID)
	ctx := context.Background()

	result := c.redisClient.Client().Get(ctx, key)
	if err := result.Err(); err != nil {
		if errors.Is(err, redisClient.Nil) {
			return false, nil // Key does not exist, user is not activated
		}
		return false, err
	}

	return true, nil
}

func (c *userActivatedCache) Delete(userID uint64) error {
	key := c.buildKey(userID)
	ctx := context.Background()

	return c.redisClient.Client().Del(ctx, key).Err()
}

func (c *userActivatedCache) buildKey(userID uint64) string {
	return fmt.Sprintf("%s%s", cacheKeyPrefix, strconv.FormatUint(userID, 10))
}

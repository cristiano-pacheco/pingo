package redis

import (
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/pkg/redis"
)

func NewRedis(config config.Config) redis.Redis {
	return redis.NewRedis(config.Redis.Addr, config.Redis.Password, config.Redis.DB)
}

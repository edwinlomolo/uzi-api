package cache

import (
	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var rdb *redis.Client

func NewCache(config config.Redis, logger *logrus.Logger) {
	opts, err := redis.ParseURL(config.Url)
	if err != nil {
		logger.Errorf("%s-%v", "ParseRedisUrlErr", err.Error())
	}

	rdb = redis.NewClient(opts)
}

func GetCache() *redis.Client { return rdb }

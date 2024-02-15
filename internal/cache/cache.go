package cache

import (
	"github.com/edwinlomolo/uzi-api/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	Redis *redis.Client
)

type Cache interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
}

func NewCache(config config.Redis, logger *logrus.Logger) {
	opts, err := redis.ParseURL(config.Url)
	if err != nil {
		logger.Errorf("%s-%v", "new cache", err.Error())
	} else {
		logger.Infoln("Redis cache...OK")
	}

	Redis = redis.NewClient(opts)
}

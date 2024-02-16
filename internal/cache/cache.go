package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type cacheClient struct {
	cache  *redis.Client
	logger *logrus.Logger
}

var Redis *redis.Client
var Rdb *cacheClient
var _ Cache = (*cacheClient)(nil)

type Cache interface {
	Get(ctx context.Context, key string, returnValue any) (any, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
}

func NewCache(config config.Redis, logger *logrus.Logger) {
	opts, err := redis.ParseURL(config.Url)
	if err != nil {
		logger.Fatalf("%s-%v", "new cache", err.Error())
	} else {
		logger.Infoln("Redis cache...OK")
	}

	rdb := redis.NewClient(opts)
	Redis = rdb
	Rdb = &cacheClient{
		rdb,
		logger,
	}
}

func (c *cacheClient) Get(ctx context.Context, key string, returnValue any) (any, error) {
	result, err := c.cache.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		uziErr := fmt.Errorf("%s:%v", "cache get", err)
		c.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	err = json.Unmarshal([]byte(result), returnValue)
	if err != nil {
		c.cache.Del(ctx, key).Err()
		uziErr := fmt.Errorf("%s:%v", "unmarshal cache value", err)
		c.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return returnValue, nil
}

func (c *cacheClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	valueMarshal, err := json.Marshal(value)
	if err != nil {
		uziErr := fmt.Errorf("%s:%v", "marshal cache value", err)
		c.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return c.cache.Set(ctx, key, valueMarshal, expiration).Err()
}

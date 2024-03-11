package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type cacheClient struct {
	cache *redis.Client
	log   *logrus.Logger
}

type Cache interface {
	Get(ctx context.Context, key string, returnValue interface{}) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	GetRedis() *redis.Client
}

func New() Cache {
	log := logger.GetLogger()
	opts, err := redis.ParseURL(config.Config.Database.Redis.Url)
	if err != nil {
		log.WithError(err).Errorf("new cache client")
	} else {
		log.Infoln("Redis cache...OK")
	}

	rdb := redis.NewClient(opts)
	return &cacheClient{
		rdb,
		log,
	}
}

func (c *cacheClient) GetRedis() *redis.Client {
	return c.cache
}

func (c *cacheClient) Get(ctx context.Context, key string, returnValue interface{}) (interface{}, error) {
	result, err := c.cache.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		c.log.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Errorf("get: reading cache value")
		return nil, err
	}

	err = json.Unmarshal([]byte(result), returnValue)
	if err != nil {
		c.cache.Del(ctx, key).Err()
		c.log.WithError(err).Errorf("get: unmarshal cache result")
		return nil, err
	}

	return returnValue, nil
}

func (c *cacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	valueMarshal, err := json.Marshal(value)
	if err != nil {
		c.log.WithFields(logrus.Fields{
			"key":   key,
			"value": value,
			"error": err,
		}).Errorf("set: marshal value")
		return err
	}

	return c.cache.Set(ctx, key, valueMarshal, expiration).Err()
}

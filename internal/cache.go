package internal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	c Cache
)

type cacheClient struct {
	cache *redis.Client
}

type Cache interface {
	Get(ctx context.Context, key string, returnValue interface{}) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	GetRedis() *redis.Client
}

func NewCache() {
	opts, err := redis.ParseURL(config.Config.Database.Redis.Url)
	if err != nil {
		log.WithError(err).Errorf("new cache client")
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.WithError(err).Fatalln("ping redis")
	}

	c = &cacheClient{
		rdb,
	}
}

func GetCache() Cache {
	return c
}

func (c *cacheClient) GetRedis() *redis.Client {
	return c.cache
}

func (c *cacheClient) Get(ctx context.Context, key string, returnValue interface{}) (interface{}, error) {
	result, err := c.cache.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		log.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Errorf("get: reading cache value")
		return nil, err
	}

	err = json.Unmarshal([]byte(result), returnValue)
	if err != nil {
		c.cache.Del(ctx, key).Err()
		log.WithError(err).Errorf("get: unmarshal cache result")
		return nil, err
	}

	return returnValue, nil
}

func (c *cacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	valueMarshal, err := json.Marshal(value)
	if err != nil {
		log.WithFields(logrus.Fields{
			"key":   key,
			"value": value,
			"error": err,
		}).Errorf("set: marshal value")
		return err
	}

	return c.cache.Set(ctx, key, valueMarshal, expiration).Err()
}

func base64Key(key interface{}) string {
	keyString, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(keyString))

	return encoded
}

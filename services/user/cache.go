package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/3dw1nM0535/uzi-api/gql/model"
	"github.com/3dw1nM0535/uzi-api/internal/cache"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type usercacheclient struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func newusercache(redis *redis.Client, logger *logrus.Logger) cache.Cache {
	return &usercacheclient{redis, logger}
}

func (usc *usercacheclient) Get(key string) (interface{}, error) {
	var res model.User

	keyValue, err := usc.redis.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		cacheErr := fmt.Errorf("%s:%v", "get user cache", err)
		usc.logger.Errorf(cacheErr.Error())
		return nil, cacheErr
	}

	if err := json.Unmarshal([]byte(keyValue), &res); err != nil {
		jsonErr := fmt.Errorf("%s:%v", "unmarshal user cache", err)
		usc.logger.Errorf(jsonErr.Error())
		return nil, jsonErr
	}

	return &res, nil
}

func (usc *usercacheclient) Set(key string, value interface{}) error {
	userinfo := value.(*model.User)
	data, err := json.Marshal(userinfo)
	if err != nil {
		cacheErr := fmt.Errorf("%s:%v", "marshal user cache value", err)
		usc.logger.Errorf(cacheErr.Error())
		return cacheErr
	}

	// TODO proper cache shell life to 24hr once ready/tested
	if err := usc.redis.Set(context.Background(), key, data, time.Minute*1).Err(); err != nil {
		cacheErr := fmt.Errorf("%s:%v", "set user cache", err)
		usc.logger.Errorf(cacheErr.Error())
		return cacheErr
	}

	return nil
}

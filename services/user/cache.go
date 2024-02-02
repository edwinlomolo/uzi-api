package user

import (
	"context"
	"encoding/json"
	"time"

	"github.com/3dw1nM0535/uzi-api/internal/cache"
	"github.com/3dw1nM0535/uzi-api/model"
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
		cacheErr := model.UziErr{Err: err.Error(), Message: "getusercache", Code: 500}
		usc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return nil, cacheErr
	}

	if err := json.Unmarshal([]byte(keyValue), &res); err != nil {
		jsonErr := model.UziErr{Err: err.Error(), Message: "getusercachemarshal", Code: 400}
		usc.logger.Errorf("%s: %s", jsonErr.Message, jsonErr.Err)
		return nil, jsonErr
	}

	return &res, nil
}

func (usc *usercacheclient) Set(key string, value interface{}) error {
	userinfo := value.(*model.User)
	data, err := json.Marshal(userinfo)
	if err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "setusercachemarshal", Code: 500}
		usc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return cacheErr
	}

	if err := usc.redis.Set(context.Background(), key, data, time.Minute*1).Err(); err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "setusercache", Code: 500}
		usc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return cacheErr
	}

	return nil
}

package location

import (
	"context"
	"encoding/json"
	"time"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type locationcacheclient struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func newlocationcacheclient(redis *redis.Client, logger *logrus.Logger) locationcacheclient {
	return locationcacheclient{redis, logger}
}

func (lc *locationcacheclient) Get(key string) (interface{}, error) {
	var res model.Geocode

	keyValue, err := lc.redis.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "getlocationcache", Code: 500}
		lc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return nil, cacheErr
	}

	if err := json.Unmarshal([]byte(keyValue), &res); err != nil {
		jsonErr := model.UziErr{Err: err.Error(), Message: "getlocationcachemarshal", Code: 400}
		lc.logger.Errorf("%s: %s", jsonErr.Message, jsonErr.Err)
		return nil, jsonErr
	}

	return &res, nil
}

func (lc *locationcacheclient) placesGetCache(key string) (interface{}, error) {
	var res []*model.Place

	keyValue, err := lc.redis.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "getplacescache", Code: 500}
		lc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return nil, cacheErr
	}

	if err := json.Unmarshal([]byte(keyValue), &res); err != nil {
		jsonErr := model.UziErr{Err: err.Error(), Message: "getplacescationcachemarshal", Code: 400}
		lc.logger.Errorf("%s: %s", jsonErr.Message, jsonErr.Err)
		return nil, jsonErr
	}

	return res, nil
}

func (lc *locationcacheclient) placesSetCache(key string, value interface{}) error {
	locationinfo := value.([]*model.Place)
	data, err := json.Marshal(locationinfo)
	if err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "setplacescationcachemarshal", Code: 500}
		lc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return cacheErr
	}

	if err := lc.redis.Set(context.Background(), key, data, time.Minute*5).Err(); err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "setplacescationcache", Code: 500}
		lc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return cacheErr
	}

	return nil
}

func (lc *locationcacheclient) Set(key string, value interface{}) error {
	locationinfo := value.(*model.Geocode)
	data, err := json.Marshal(locationinfo)
	if err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "setlocationcachemarshal", Code: 500}
		lc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return cacheErr
	}

	if err := lc.redis.Set(context.Background(), key, data, time.Minute*5).Err(); err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "setlocationcache", Code: 500}
		lc.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return cacheErr
	}

	return nil
}

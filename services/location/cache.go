package location

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/3dw1nM0535/uzi-api/gql/model"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type locationCache struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func newlocationCache(redis *redis.Client, logger *logrus.Logger) locationCache {
	return locationCache{redis, logger}
}

func (lc *locationCache) Get(key string) (interface{}, error) {
	var res Geocode

	keyValue, err := lc.redis.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		cacheErr := fmt.Errorf("%s:%v", "get location cache", err)
		lc.logger.Errorf(cacheErr.Error())
		return nil, cacheErr
	}

	if err := json.Unmarshal([]byte(keyValue), &res); err != nil {
		jsonErr := fmt.Errorf("%s:%v", "unmarshal location cache", err)
		lc.logger.Errorf(jsonErr.Error())
		return nil, jsonErr
	}

	return &res, nil
}

func (lc *locationCache) placesGetCache(key string) (interface{}, error) {
	var res []*model.Place

	keyValue, err := lc.redis.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		cacheErr := fmt.Errorf("%s:%v", "get places cache", err)
		lc.logger.Errorf(cacheErr.Error())
		return nil, cacheErr
	}

	if err := json.Unmarshal([]byte(keyValue), &res); err != nil {
		jsonErr := fmt.Errorf("%s:%v", "unmarshal places cache", err)
		lc.logger.Errorf(jsonErr.Error())
		return nil, jsonErr
	}

	return res, nil
}

func (lc *locationCache) placesSetCache(key string, value interface{}) error {
	locationinfo := value.([]*model.Place)
	data, err := json.Marshal(locationinfo)
	if err != nil {
		cacheErr := fmt.Errorf("%s:%v", "marshal places cache", err)
		lc.logger.Errorf(cacheErr.Error())
		return cacheErr
	}

	if err := lc.redis.Set(context.Background(), key, data, time.Minute*5).Err(); err != nil {
		cacheErr := fmt.Errorf("%s:%v", "set place cache", err)
		lc.logger.Errorf(cacheErr.Error())
		return cacheErr
	}

	return nil
}

func (lc *locationCache) Set(key string, value interface{}) error {
	locationinfo := value.(*Geocode)
	data, err := json.Marshal(locationinfo)
	if err != nil {
		cacheErr := fmt.Errorf("%s:%v", "marshal location cache", err)
		lc.logger.Errorf(cacheErr.Error())
		return cacheErr
	}

	if err := lc.redis.Set(context.Background(), key, data, time.Minute*5).Err(); err != nil {
		cacheErr := fmt.Errorf("%s:%v", "set location cache", err)
		lc.logger.Errorf(cacheErr.Error())
		return cacheErr
	}

	return nil
}

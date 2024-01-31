package route

import (
	"context"
	"encoding/json"
	"time"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type routeCache struct {
	cache  *redis.Client
	logger *logrus.Logger
}

func newrouteCache(redis *redis.Client, logger *logrus.Logger) routeCache {
	return routeCache{redis, logger}
}

func (r routeCache) cacheRoute(key string, value interface{}) error {
	routeinfo := value.(*model.TripRoute)
	data, err := json.Marshal(routeinfo)
	if err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "setrouteinfocachemarshal", Code: 500}
		r.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return cacheErr
	}

	if err := r.cache.Set(context.Background(), key, data, time.Hour*24).Err(); err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "setrouteinfocache", Code: 500}
		r.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return cacheErr
	}

	return nil
}

func (r routeCache) getRouteCache(key string) (interface{}, error) {
	var res *model.TripRoute

	keyValue, err := r.cache.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		cacheErr := model.UziErr{Err: err.Error(), Message: "getrouteinfocache", Code: 500}
		r.logger.Errorf("%s: %s", cacheErr.Message, cacheErr.Err)
		return nil, cacheErr
	}

	if err := json.Unmarshal([]byte(keyValue), &res); err != nil {
		jsonErr := model.UziErr{Err: err.Error(), Message: "getrouteinfocachemarshal", Code: 400}
		r.logger.Errorf("%s: %s", jsonErr.Message, jsonErr.Err)
		return nil, jsonErr
	}

	return res, nil
}

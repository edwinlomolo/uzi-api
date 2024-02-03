package route

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/3dw1nM0535/uzi-api/gql/model"
	"github.com/3dw1nM0535/uzi-api/internal/cache"
	"github.com/3dw1nM0535/uzi-api/internal/logger"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type routeCache struct {
	cache  *redis.Client
	logger *logrus.Logger
}

func newCache() routeCache {
	return routeCache{cache.Redis, logger.Logger}
}

func (r routeCache) cacheRoute(key string, value interface{}) error {
	routeinfo := value.(*model.TripRoute)
	data, err := json.Marshal(routeinfo)
	if err != nil {
		r.logger.Errorf(err.Error())
		return fmt.Errorf("%s:%v", "setrouteinfocachemarshal", err.Error())
	}

	if err := r.cache.Set(context.Background(), key, data, time.Hour*24).Err(); err != nil {
		r.logger.Errorf(err.Error())
		return fmt.Errorf("%s:%v", "setrouteinfocache", err.Error())
	}

	return nil
}

func (r routeCache) getRouteCache(key string) (interface{}, error) {
	var res *model.TripRoute

	keyValue, err := r.cache.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		r.logger.Errorf(err.Error())
		return nil, fmt.Errorf("%s:%v", "getrouteinfocache", err.Error())
	}

	if err := json.Unmarshal([]byte(keyValue), &res); err != nil {
		r.logger.Errorf(err.Error())
		return nil, fmt.Errorf("%s:%v", "getrouteinfocachemarshal", err.Error())
	}

	return res, nil
}

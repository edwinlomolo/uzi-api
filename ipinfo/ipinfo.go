package ipinfo

import (
	"context"
	"encoding/json"
	"net"
	"time"

	redisCache "github.com/edwinlomolo/uzi-api/cache"
	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/ipinfo/go/v2/ipinfo"
	"github.com/ipinfo/go/v2/ipinfo/cache"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var log = logger.GetLogger()

type IpInfoService interface {
	GetIpinfo(ip string) (*ipinfo.Core, error)
}

type ipinfoClient struct {
	config config.Ipinfo
	log    *logrus.Logger
	client *ipinfo.Client
}

func New(redis redisCache.Cache) IpInfoService {
	cache := newCache(redis)
	c := ipinfo.NewCache(cache)
	client := ipinfo.NewClient(
		nil,
		c,
		config.Config.Ipinfo.ApiKey,
	)

	log.Infoln("Ipinfo service...OK")
	return &ipinfoClient{
		config.Config.Ipinfo,
		log,
		client,
	}
}

func (ipc *ipinfoClient) GetIpinfo(
	ip string,
) (*ipinfo.Core, error) {
	info, err := ipc.client.GetIPInfo(net.ParseIP(ip))
	if err != nil {
		ipc.log.WithFields(logrus.Fields{
			"ip":    ip,
			"error": err,
		}).Errorf("get ip info")
		return nil, err
	}

	return info, nil
}

type ipinfoCache struct {
	log   *logrus.Logger
	redis *redis.Client
}

func (ipc *ipinfoCache) Get(
	key string,
) (interface{}, error) {
	var res ipinfo.Core

	keyValue, err := ipc.redis.Get(
		context.Background(),
		key,
	).Result()
	if err != redis.Nil && err != nil {
		ipc.log.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Errorf("get: ipinfo cache value")
		return nil, err
	}

	if err := json.Unmarshal(
		[]byte(keyValue),
		&res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (ipc *ipinfoCache) Set(
	key string,
	value interface{},
) error {
	ipcinfo := value.(*ipinfo.Core)
	data, err := json.Marshal(ipcinfo)
	if err != nil {
		return err
	}

	if err := ipc.redis.Set(
		context.Background(),
		key,
		data,
		time.Hour*24).Err(); err != nil {
		ipc.log.WithFields(logrus.Fields{
			"key":   key,
			"value": value,
			"error": err,
		}).Errorf("set: ipinfo cache value")
		return err
	}

	return nil
}

func newCache(c redisCache.Cache) cache.Interface {
	return &ipinfoCache{
		logger.New(),
		c.Redis(),
	}
}

package services

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/ipinfo/go/v2/ipinfo"
	"github.com/ipinfo/go/v2/ipinfo/cache"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var ipinfoService IpInfo

type IpInfo interface {
	GetIpinfo(ip string) (*ipinfo.Core, error)
}

type ipinfoClient struct {
	config config.Ipinfo
	logger *logrus.Logger
	client *ipinfo.Client
}

func NewIpinfoService(redis *redis.Client, config config.Ipinfo, logger *logrus.Logger) {
	cache := newipinfocache(redis, logger, config)
	c := ipinfo.NewCache(cache)
	client := ipinfo.NewClient(nil, c, config.ApiKey)

	ipinfoService = &ipinfoClient{config, logger, client}
}

func GetIpinfoService() IpInfo {
	return ipinfoService
}

func (ipc *ipinfoClient) GetIpinfo(ip string) (*ipinfo.Core, error) {
	info, err := ipc.client.GetIPInfo(net.ParseIP(ip))
	if err != nil {
		ipc.logger.Errorf("%s-%v", "IPInfoErr", err.Error())
		return nil, err
	}

	return info, nil
}

type ipinfocacheClient struct {
	logger *logrus.Logger
	redis  *redis.Client
}

func (ipc *ipinfocacheClient) Get(key string) (interface{}, error) {
	var res ipinfo.Core

	keyValue, err := ipc.redis.Get(context.Background(), key).Result()
	if err != redis.Nil && err != nil {
		ipc.logger.Errorf("%s-%v", "IpinfoGetValueErr", err.Error())
		return nil, err
	}

	if err := json.Unmarshal([]byte(keyValue), &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (ipc *ipinfocacheClient) Set(key string, value interface{}) error {
	ipcinfo := value.(*ipinfo.Core)
	data, err := json.Marshal(ipcinfo)
	if err != nil {
		return err
	}

	if err := ipc.redis.Set(context.Background(), key, data, time.Hour*24*7).Err(); err != nil {
		ipc.logger.Errorf("%s-%v", "IpinfoCacheSetErr", err.Error())
		return err
	}

	return nil
}

func newipinfocache(redis *redis.Client, logger *logrus.Logger, config config.Ipinfo) cache.Interface {
	return &ipinfocacheClient{logger, redis}
}

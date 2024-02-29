package ipinfo

import (
	"context"
	"encoding/json"
	"fmt"
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

var (
	IpInfo IpInfoService
)

type IpInfoService interface {
	GetIpinfo(ip string) (*ipinfo.Core, error)
}

type ipinfoClient struct {
	config config.Ipinfo
	logger *logrus.Logger
	client *ipinfo.Client
}

func NewIpinfoService() {
	cache := newCache()
	c := ipinfo.NewCache(cache)
	client := ipinfo.NewClient(
		nil,
		c,
		config.Config.Ipinfo.ApiKey,
	)

	IpInfo = &ipinfoClient{
		config.Config.Ipinfo,
		logger.Logger,
		client,
	}
	logger.Logger.Infoln("Ipinfo service...OK")
}

func (ipc *ipinfoClient) GetIpinfo(
	ip string,
) (*ipinfo.Core, error) {
	info, err := ipc.client.GetIPInfo(net.ParseIP(ip))
	if err != nil {
		ipErr := fmt.Errorf("%s:%v", "ipinfo", err)
		ipc.logger.Errorf(ipErr.Error())
		return nil, ipErr
	}

	return info, nil
}

type ipinfoCache struct {
	logger *logrus.Logger
	redis  *redis.Client
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
		uziErr := fmt.Errorf("%s:%v", "ipinfo cache", err)
		ipc.logger.Errorf(uziErr.Error())
		return nil, uziErr
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
		uziErr := fmt.Errorf("%s:%v", "ipinfo cache", err.Error())
		ipc.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return nil
}

func newCache() cache.Interface {
	return &ipinfoCache{
		logger.Logger,
		redisCache.Redis,
	}
}

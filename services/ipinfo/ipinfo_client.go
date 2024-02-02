package ipinfo

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/3dw1nM0535/uzi-api/config"
	redisCache "github.com/3dw1nM0535/uzi-api/internal/cache"
	"github.com/3dw1nM0535/uzi-api/internal/logger"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/ipinfo/go/v2/ipinfo"
	"github.com/ipinfo/go/v2/ipinfo/cache"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var ipinfoService IpInfo

type ipinfoClient struct {
	config config.Ipinfo
	logger *logrus.Logger
	client *ipinfo.Client
}

func NewIpinfoService() {
	log := logger.GetLogger()
	cfg := config.GetConfig().Ipinfo
	cache := newipinfocache(redisCache.GetCache(), log, cfg)
	c := ipinfo.NewCache(cache)
	client := ipinfo.NewClient(nil, c, cfg.ApiKey)

	ipinfoService = &ipinfoClient{cfg, log, client}
	log.Infoln("Ipinfo service...OK")
}

func GetIpinfoService() IpInfo {
	return ipinfoService
}

func (ipc *ipinfoClient) GetIpinfo(ip string) (*ipinfo.Core, error) {
	info, err := ipc.client.GetIPInfo(net.ParseIP(ip))
	if err != nil {
		ipErr := model.UziErr{Err: err.Error(), Message: "IpinfoErr", Code: 500}
		ipc.logger.Errorf("%s: %s", ipErr.Message, ipErr.Err)
		return nil, ipErr
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
		ipc.logger.Errorf("%s: %v", "IpinfoGetValueErr", err.Error())
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

	if err := ipc.redis.Set(context.Background(), key, data, time.Hour*24*365).Err(); err != nil {
		ipc.logger.Errorf("%s: %v", "IpinfoCacheSetErr", err.Error())
		return err
	}

	return nil
}

func newipinfocache(redis *redis.Client, logger *logrus.Logger, config config.Ipinfo) cache.Interface {
	return &ipinfocacheClient{logger, redis}
}

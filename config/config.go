package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/3dw1nM0535/uzi-api/pkg/util"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Configuration struct {
	Server   Server
	Database Database
	Ipinfo   Ipinfo
	Jwt      Jwt
	Aws      Aws
}

// Env - load env
func Env() {
	// Load os env
	err := godotenv.Load()
	if err != nil {
		logrus.Errorf("Error loading .env variable: %v", err)
	}
}

var configAll *Configuration

// LoadConfig - load all configuration
func LoadConfig() *Configuration {
	var configuration Configuration

	configuration.Server = serverConfig()
	configuration.Database = databaseConfig()
	configuration.Ipinfo = ipinfoConfig()
	configuration.Jwt = jwtConfig()
	configuration.Aws = awsConfig()

	configAll = &configuration

	return configAll
}

// GetConfig - get configurations
func GetConfig() *Configuration {
	return configAll
}

// serverConfig - load server configuration
func serverConfig() Server {
	var serverConfig Server

	Env()

	serverConfig.Env = strings.TrimSpace(os.Getenv("SERVERENV"))
	serverConfig.Port = strings.TrimSpace(os.Getenv("SERVERPORT"))

	return serverConfig
}

// rdbmsConfig - get relational database management system config
func rdbmsConfig() RDBMS {
	var rdbmsConfig RDBMS

	Env()

	rdbmsConfig.Postal.Uri = strings.TrimSpace(os.Getenv("POSTAL_DATABASE_URI"))
	rdbmsConfig.Uri = strings.TrimSpace(os.Getenv("DATABASE_URI"))
	rdbmsConfig.Env.Driver = strings.TrimSpace(os.Getenv("DBDRIVER"))

	return rdbmsConfig
}

// databaseConfig - load database configurations
func databaseConfig() Database {
	var databaseConfig Database

	Env()

	databaseConfig.Rdbms = rdbmsConfig()
	databaseConfig.Redis = redisConfig()

	forceMigration, err := strconv.ParseBool(strings.TrimSpace(os.Getenv("FORCE_MIGRATION")))
	if err != nil {
		panic(err)
	}

	databaseConfig.ForceMigration = forceMigration

	return databaseConfig
}

// ipinfoConfig - load ipinfo config
func ipinfoConfig() Ipinfo {
	var ipInfo Ipinfo

	Env()

	ipInfo.ApiKey = strings.TrimSpace(os.Getenv("IPINFO_API_KEY"))

	return ipInfo
}

// redisConfig - load redis configs
func redisConfig() Redis {
	var redis Redis

	Env()

	redis.Url = strings.TrimSpace(os.Getenv("REDIS_ENDPOINT"))

	return redis
}

// jwtConfig - get jwt configs
func jwtConfig() Jwt {
	var jwtConfig Jwt

	Env()

	jwtExpires, err := util.ParseDuration(strings.TrimSpace(os.Getenv("JWTEXPIRE")))
	if err != nil {
		panic(err)
	}

	jwtConfig.Expires = jwtExpires
	jwtConfig.Secret = strings.TrimSpace(os.Getenv("JWTSECRET"))

	return jwtConfig
}

// awsConfig - get aws config
func awsConfig() Aws {
	var awsConfig Aws

	Env()

	awsConfig.AccessKey = strings.TrimSpace(os.Getenv("ACCESS_KEY"))
	awsConfig.SecretAccessKey = strings.TrimSpace(os.Getenv("SECRET_ACCESS_KEY"))
	awsConfig.S3.Buckets.Media = strings.TrimSpace(os.Getenv("S3_BUCKET"))

	return awsConfig
}

func IsDev() bool {
	return configAll.Server.Env == "development"
}

func IsProd() bool {
	return configAll.Server.Env == "production"
}

func IsStaging() bool {
	return configAll.Server.Env == "staging"
}

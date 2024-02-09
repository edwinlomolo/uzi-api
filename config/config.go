package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/edwinlomolo/uzi-api/internal/util"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Configuration struct {
	Server     Server
	Database   Database
	Ipinfo     Ipinfo
	Jwt        Jwt
	Aws        Aws
	GoogleMaps GoogleMaps
	Pricer     Pricer
}

// Env - load env
func Env() {
	// Load os env
	err := godotenv.Load()
	if err != nil {
		logrus.Errorf(".env: %v", err)
	}
}

var Config *Configuration

// LoadConfig - load all configuration
func LoadConfig() {
	var configuration Configuration

	configuration.Server = serverConfig()
	configuration.Database = databaseConfig()
	configuration.Ipinfo = ipinfoConfig()
	configuration.Jwt = jwtConfig()
	configuration.Aws = awsConfig()
	configuration.GoogleMaps = googleMapsConfig()
	configuration.Pricer = pricerConfig()

	Config = &configuration
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
	rdbmsConfig.MigrationUrl = strings.TrimSpace(os.Getenv("MIGRATION_URL"))

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

// googleMapsConfig - get google map config
func googleMapsConfig() GoogleMaps {
	var googleMapsConfig GoogleMaps

	Env()

	googleMapsConfig.GooglePlacesApiKey = strings.TrimSpace(os.Getenv("MAPS_PLACES_API_KEY"))
	googleMapsConfig.GoogleGeocodeApiKey = strings.TrimSpace(os.Getenv("MAPS_GEOCODE_API_KEY"))
	googleMapsConfig.GoogleRoutesApiKey = strings.TrimSpace(os.Getenv("MAPS_ROUTES_API_KEY"))

	return googleMapsConfig
}

// pricerConfig - get pricing config
func pricerConfig() Pricer {
	var pricingConfig Pricer

	Env()

	hourlyWage, err := strconv.Atoi(strings.TrimSpace(os.Getenv("MINIMUM_HOURLY_WAGE")))
	if err != nil {
		panic(err)
	}

	pricingConfig.HourlyWage = hourlyWage

	return pricingConfig
}

func IsDev() bool {
	return Config.Server.Env == "development"
}

func IsProd() bool {
	return Config.Server.Env == "production"
}

func IsStaging() bool {
	return Config.Server.Env == "staging"
}

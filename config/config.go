package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type Configuration struct {
	Server     Server
	Database   Database
	Ipinfo     Ipinfo
	Jwt        Jwt
	Aws        Aws
	GoogleMaps GoogleMaps
	Pricer     Pricer
	Sentry     Sentry
}

// Env - load env
func Env() {
	// Load os env
	godotenv.Load()
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
	configuration.Sentry = sentryConfig()

	Config = &configuration
}

// serverConfig - load server configuration
func serverConfig() Server {
	var config Server

	Env()

	config.Env = strings.TrimSpace(os.Getenv("SERVERENV"))
	config.Port = strings.TrimSpace(os.Getenv("SERVERPORT"))

	return config
}

// rdbmsConfig - get relational database management system config
func rdbmsConfig() RDBMS {
	var config RDBMS

	Env()

	config.Postal.Uri = strings.TrimSpace(os.Getenv("POSTAL_DATABASE_URI"))
	config.Uri = strings.TrimSpace(os.Getenv("DATABASE_URI"))
	config.Env.Driver = strings.TrimSpace(os.Getenv("DBDRIVER"))
	config.MigrationUrl = strings.TrimSpace(os.Getenv("MIGRATION_URL"))
	forceMigrate, err := strconv.ParseBool(strings.TrimSpace(os.Getenv("FORCE_MIGRATION")))
	if err != nil {
		log.WithError(err).Fatalln("force migrate env")
	}
	config.ForceMigrate = forceMigrate

	return config
}

// databaseConfig - load database configurations
func databaseConfig() Database {
	var config Database

	Env()

	config.Rdbms = rdbmsConfig()
	config.Redis = redisConfig()

	return config
}

// ipinfoConfig - load ipinfo config
func ipinfoConfig() Ipinfo {
	var config Ipinfo

	Env()

	config.ApiKey = strings.TrimSpace(os.Getenv("IPINFO_API_KEY"))

	return config
}

// redisConfig - load redis configs
func redisConfig() Redis {
	var config Redis

	Env()

	config.Url = strings.TrimSpace(os.Getenv("REDIS_ENDPOINT"))

	return config
}

// jwtConfig - get jwt configs
func jwtConfig() Jwt {
	var config Jwt

	Env()

	duration, err := time.ParseDuration(strings.TrimSpace(os.Getenv("JWTEXPIRE")))
	if err != nil {
		log.WithError(err).Fatalln("jwt expire parsing")
	}

	config.Expires = duration
	config.Secret = strings.TrimSpace(os.Getenv("JWTSECRET"))

	return config
}

// awsConfig - get aws config
func awsConfig() Aws {
	var config Aws

	Env()

	config.AccessKey = strings.TrimSpace(os.Getenv("ACCESS_KEY"))
	config.SecretAccessKey = strings.TrimSpace(os.Getenv("SECRET_ACCESS_KEY"))
	config.S3.Buckets.Media = strings.TrimSpace(os.Getenv("S3_BUCKET"))

	return config
}

// config - get google map config
func googleMapsConfig() GoogleMaps {
	var config GoogleMaps

	Env()

	config.GooglePlacesApiKey = strings.TrimSpace(os.Getenv("MAPS_PLACES_API_KEY"))
	config.GoogleGeocodeApiKey = strings.TrimSpace(os.Getenv("MAPS_GEOCODE_API_KEY"))
	config.GoogleRoutesApiKey = strings.TrimSpace(os.Getenv("MAPS_ROUTES_API_KEY"))

	return config
}

// pricerConfig - get pricing config
func pricerConfig() Pricer {
	var config Pricer

	Env()

	hourlyWage, err := strconv.Atoi(strings.TrimSpace(os.Getenv("MINIMUM_HOURLY_WAGE")))
	if err != nil {
		log.WithError(err).Fatalln("hourly wage env")
	}

	config.HourlyWage = hourlyWage

	return config
}

// sentryConfig - get sentry config
func sentryConfig() Sentry {
	var config Sentry

	Env()

	config.Dsn = strings.TrimSpace(os.Getenv("SENTRY_DSN"))

	return config
}

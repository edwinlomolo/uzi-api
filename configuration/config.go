package configuration

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Configuration struct {
	Server   Server
	Database Database
}

// env - load env
func env() {
	// Load os env
	err := godotenv.Load()
	if err != nil {
		logrus.Errorf("Error loading .env variable: %v", err)
	}
}

var configAll *Configuration

// LoadConfig - load all configuration
func LoadConfig() {
	var configuration Configuration

	configuration.Server = serverConfig()
	configuration.Database = databaseConfig()

	configAll = &configuration
}

// GetConfig - get configurations
func GetConfig() *Configuration {
	return configAll
}

// serverConfig - load server configuration
func serverConfig() Server {
	var serverConfig Server

	env()

	serverConfig.Env = os.Getenv("SERVERENV")
	serverConfig.Port = os.Getenv("SERVERPORT")

	return serverConfig
}

// rdbmsConfig - get relational database management system config
func rdbmsConfig() RDBMS {
	var rdbmsConfig RDBMS

	env()

	rdbmsConfig.Postal.Uri = os.Getenv("POSTAL_DATABASE_URI")
	rdbmsConfig.Uri = os.Getenv("DATABASE_URI")
	rdbmsConfig.Env.Driver = os.Getenv("DBDRIVER")

	return rdbmsConfig
}

// databaseConfig - load database configurations
func databaseConfig() Database {
	var databaseConfig Database

	env()

	databaseConfig.Rdbms = rdbmsConfig()

	forceMigration, err := strconv.ParseBool(os.Getenv("FORCE_MIGRATION"))
	if err != nil {
		panic(err)
	}

	databaseConfig.ForceMigration = forceMigration

	return databaseConfig
}

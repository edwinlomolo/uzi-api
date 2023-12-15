package store

import (
	"database/sql"
	"fmt"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
)

var dbClient *sql.DB

func InitializeStorage(logger *logrus.Logger, migrationUrl string) (*sql.DB, error) {
	configs := config.GetConfig()
	databaseconfigs := configs.Database.Rdbms
	isDevelopment := configs.Server.Env != "production"
	forceMigrate := configs.Database.ForceMigration

	db, err := sql.Open(databaseconfigs.Env.Driver, databaseconfigs.Uri)
	if err != nil {
		logrus.Errorf("%s:%v", "DatabaseError", err)
		return nil, err
	}
	db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %q;", "uuid-ossp"))
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis;")
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis_rasters; --OPTIONAL")
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis_topology; --OPTIONAL")

	if err := db.Ping(); err != nil {
		logrus.Errorf("%s:%v", "DatabasePingError", err.Error())
	} else if err == nil {
		logrus.Infoln("Database connected!")
	}

	dbClient = db

	// Setup database schema
	if err := runDatabaseMigration(dbClient, logger, isDevelopment, forceMigrate, migrationUrl); err != nil {
		logger.Errorf("%s:%v", "ApplyingMigrationErr", err.Error())
		return nil, err
	} else if err == nil {
		logger.Infoln("Database migration applied")
	}

	return dbClient, nil

}

// runDbMigration - setup database tables
func runDatabaseMigration(db *sql.DB, logger *logrus.Logger, isDevelopment, forceMigrate bool, migrationUrl string) error {
	migrationErr := "DatabaseMigrationErr"

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Errorf("%s: %s", migrationErr, err)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", migrationUrl), "postgres", driver)
	if err != nil {
		logger.Errorf("%s: %s", migrationErr, err)
		return err
	}

	// Apply migration(s)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Errorf("%s: %s", migrationErr, err)
		return err
	}

	// Drop stuff in dev mode if forced
	if isDevelopment && forceMigrate {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			logger.Errorf("%s: %s", migrationErr, err)
			return err
		}
	}

	return nil
}

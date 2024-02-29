package store

import (
	"database/sql"
	"fmt"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/logger"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
)

var (
	DB *sqlStore.Queries
)

func InitializeStorage() error {
	log := logger.Logger
	configs := config.Config
	rdbmsConfig := configs.Database.Rdbms
	isDevelopment := config.IsDev()
	forceMigrate := configs.Database.ForceMigration

	db, err := sql.Open(rdbmsConfig.Env.Driver, rdbmsConfig.Uri)
	if err != nil {
		log.Errorf("%s:%v", "DatabaseError", err)
		return err
	}

	db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %q;", "uuid-ossp"))
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis;")
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis_rasters; --OPTIONAL")
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis_topology; --OPTIONAL")

	if err := db.Ping(); err != nil {
		log.Errorf("%s:%v", "DatabasePingError", err.Error())
		return err
	} else if err == nil {
		log.Infoln("Database connection...OK")
	}

	DB = sqlStore.New(db)

	// Setup database schema
	if err := runDatabaseMigration(
		db,
		log,
		isDevelopment,
		rdbmsConfig.MigrationUrl,
		forceMigrate); err == nil {
		log.Infoln("Database migration...DONE")
	}

	return nil
}

// runDbMigration - setup database tables
func runDatabaseMigration(
	db *sql.DB,
	logger *logrus.Logger,
	isDevelopment bool,
	migrationUrl string,
	forceMigrate bool,
) error {
	migrationErr := "DatabaseMigrationErr"

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Errorf("%s: %s", migrationErr, err)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(migrationUrl, "postgres", driver)
	if err != nil {
		logger.Errorf("%s: %s", migrationErr, err)
		return err
	}

	if forceMigrate {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			logger.Errorf("%s:%v", "ResetMigration", err)
		}
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Errorf("%s: %s", migrationErr, err)
		return err
	}

	return nil
}

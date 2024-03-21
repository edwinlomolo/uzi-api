package store

import (
	"database/sql"
	"fmt"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/internal"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var (
	log = internal.GetLogger()
	dB  *sqlStore.Queries
)

func InitializeStorage() error {
	db, err := sql.Open(config.Config.Database.Rdbms.Env.Driver, config.Config.Database.Rdbms.Uri)
	if err != nil {
		log.WithError(err).Errorf("open database connection")
		return err
	}

	db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %q;", "uuid-ossp"))
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis;")
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis_rasters; --OPTIONAL")
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis_topology; --OPTIONAL")

	if err := db.Ping(); err != nil {
		log.WithError(err).Fatalln("ping database connection")
		return err
	} else if err == nil {
		log.Infoln("Database connection...OK")
	}

	// Setup database schema
	if err := runDatabaseMigration(db); err == nil {
		log.Infoln("Database migration...DONE")
	}

	dB = sqlStore.New(db)

	return nil
}

func GetDb() *sqlStore.Queries {
	return dB
}

// runDbMigration - setup database tables
func runDatabaseMigration(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.WithError(err).Errorf("migration driver")
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(config.Config.Database.Rdbms.MigrationUrl, "postgres", driver)
	if err != nil {
		log.WithError(err).Errorf("new migrate instance")
		return err
	}

	if config.Config.Database.Rdbms.ForceMigrate {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.WithError(err).Errorf("reset migration tables")
		}
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.WithError(err).Errorf("setup database schema")
		return err
	}

	return nil
}

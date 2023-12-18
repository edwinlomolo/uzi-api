package config

type Database struct {
	Rdbms          RDBMS
	Redis          Redis
	ForceMigration bool
}

package config

type RDBMS struct {
	Uri    string
	Postal struct {
		Uri string
	}
	Env struct {
		Driver string
	}
}

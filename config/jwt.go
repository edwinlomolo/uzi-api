package config

import "time"

type Jwt struct {
	Secret  string
	Expires time.Duration
}

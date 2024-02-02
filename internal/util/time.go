package util

import (
	"time"
)

func ParseDuration(duration string) (time.Duration, error) {
	t, err := time.ParseDuration(duration)
	if err != nil {
		panic(err)
	}

	return t, nil
}

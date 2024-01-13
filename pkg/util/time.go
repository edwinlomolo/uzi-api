package util

import (
	"time"
)

func ParseDuration(duration string) (time.Time, error) {
	t, err := time.ParseDuration(duration)
	if err != nil {
		panic(err)
	}

	return time.Now().Add(t), nil
}

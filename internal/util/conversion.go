package util

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"

	"github.com/edwinlomolo/uzi-api/gql/model"
)

type point struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

func FloatToString(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func Base64Key(key interface{}) string {
	keyString, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(keyString))

	return encoded
}

func ParsePostgisLocation(p interface{}) *model.Gps {
	var location *point

	if p != nil {
		json.Unmarshal([]byte((p).(string)), &location)

		lat := &location.Coordinates[1]
		lng := &location.Coordinates[0]
		return &model.Gps{
			Lat: *lat,
			Lng: *lng,
		}
	} else {
		return nil
	}
}

func ParseDuration(
	duration string,
) (time.Duration, error) {
	t, err := time.ParseDuration(duration)
	if err != nil {
		panic(err)
	}

	return t, nil
}

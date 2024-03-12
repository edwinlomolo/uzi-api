package model

import "encoding/json"

type Geocode struct {
	PlaceID          string
	FormattedAddress string
	Location         Gps
}

type point struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

func ParsePostgisLocation(p interface{}) *Gps {
	var location *point

	if p != nil {
		json.Unmarshal([]byte((p).(string)), &location)

		lat := &location.Coordinates[1]
		lng := &location.Coordinates[0]
		return &Gps{
			Lat: *lat,
			Lng: *lng,
		}
	} else {
		return nil
	}
}

package model

type Geocode struct {
	PlaceID          string
	FormattedAddress string
	Location         Gps
}

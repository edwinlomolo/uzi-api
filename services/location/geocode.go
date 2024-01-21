package location

import "github.com/3dw1nM0535/uzi-api/model"

type Geocode interface {
	GeocodeLatLng(input model.GpsInput) (*model.Geocode, error)
	GetPlaceDetails(placeID string) (*model.Geocode, error)
}

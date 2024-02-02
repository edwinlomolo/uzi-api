package location

import "github.com/3dw1nM0535/uzi-api/gql/model"

type Location interface {
	GeocodeLatLng(input model.GpsInput) (*Geocode, error)
	AutocompletePlace(query string) ([]*model.Place, error)
	GetPlaceDetails(placeID string) (*Geocode, error)
}

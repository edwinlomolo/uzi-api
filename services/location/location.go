package location

import "github.com/3dw1nM0535/uzi-api/model"

type Location interface {
	AutoCompletePlace(searchQuery string) ([]*model.Place, error)
	GeocodeLatLng(input model.GpsInput) (*model.Place, error)
}

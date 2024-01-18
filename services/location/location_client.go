package location

import (
	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
)

type locationClient struct {
	config config.GoogleMaps
}

var locationService Location

func GetLocationService() Location { return locationService }

func NewLocationService(cfg config.GoogleMaps) Location {
	locationService = &locationClient{cfg}
	return locationService
}

func (l *locationClient) AutoCompletePlace(searchQuery string) ([]*model.Place, error) {
	return make([]*model.Place, 0), nil
}

func (l *locationClient) GeocodeLatLng(input model.GpsInput) (*model.Place, error) {
	return nil, nil
}

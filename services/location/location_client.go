package location

import (
	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/sirupsen/logrus"
)

type locationClient struct {
	config config.GoogleMaps
	logger *logrus.Logger
}

var locationService Location

func GetLocationService() Location { return locationService }

func NewLocationService(cfg config.GoogleMaps, logger *logrus.Logger) Location {
	locationService = &locationClient{cfg, logger}
	return locationService
}

func (l *locationClient) AutoCompletePlace(searchQuery string) ([]*model.Place, error) {
	return make([]*model.Place, 0), nil
}

func (l *locationClient) GeocodeLatLng(input model.GpsInput) (*model.Place, error) {
	return nil, nil
}

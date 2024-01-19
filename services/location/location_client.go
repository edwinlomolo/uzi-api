package location

import (
	"context"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
)

type locationClient struct {
	places, geocode *maps.Client
	config          config.GoogleMaps
	logger          *logrus.Logger
}

var locationService Location

func GetLocationService() Location { return locationService }

func NewLocationService(cfg config.GoogleMaps, logger *logrus.Logger) Location {
	places, placesErr := maps.NewClient(maps.WithAPIKey(cfg.GooglePlacesApiKey))
	if placesErr != nil {
		logger.Errorf("%s:%v", "new places", placesErr.Error())
	} else {
		logger.Infoln("Places service...OK")
	}

	geocode, geocodeErr := maps.NewClient(maps.WithAPIKey(cfg.GoogleGeocodeApiKey))
	if geocodeErr != nil {
		logger.Errorf("%s: %v", "new geocode err", geocodeErr.Error())
	} else {
		logger.Infoln("Geocode service...OK")
	}

	locationService = &locationClient{places, geocode, cfg, logger}

	return locationService
}

func (l *locationClient) AutocompletePlace(searchQuery string) ([]*model.Place, error) {
	componentsFilter := map[maps.Component][]string{
		maps.ComponentCountry: {"KE"},
	}

	req := &maps.PlaceAutocompleteRequest{
		Input:      searchQuery,
		Components: componentsFilter,
	}

	places, err := l.places.PlaceAutocomplete(context.Background(), req)
	if err != nil {
		placesErr := model.UziErr{Err: err.Error(), Message: "placeautocomplete", Code: 500}
		l.logger.Errorf(placesErr.Error())
		return nil, placesErr
	}
	l.logger.Infoln(places)

	return make([]*model.Place, 0), nil
}

func (l *locationClient) GeocodeLatLng(input model.GpsInput) (*model.Place, error) {
	reqFilter := map[maps.Component]string{
		maps.ComponentCountry: "KE",
	}

	req := &maps.GeocodingRequest{
		LatLng:     &maps.LatLng{Lat: input.Lat, Lng: input.Lng},
		Components: reqFilter,
	}

	res, err := l.geocode.ReverseGeocode(context.Background(), req)
	if err != nil {
		geocodeErr := model.UziErr{Err: err.Error(), Message: "reversegeocode", Code: 500}
		l.logger.Errorf(geocodeErr.Error())
		return nil, geocodeErr
	}
	l.logger.Infoln(res)

	return nil, nil
}

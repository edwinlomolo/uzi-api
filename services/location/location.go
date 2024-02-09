package location

import (
	"context"
	"fmt"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal/logger"
	"github.com/edwinlomolo/uzi-api/internal/util"
	"github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
)

const (
	nominatimApi = "https://nominatim.openstreetmap.org"
)

var Location LocationService

type LocationService interface {
	GeocodeLatLng(input model.GpsInput) (*Geocode, error)
	AutocompletePlace(query string) ([]*model.Place, error)
	GetPlaceDetails(placeID string) (*Geocode, error)
}

type Geocode struct {
	PlaceID          string
	FormattedAddress string
	Location         model.Gps
}

type locationClient struct {
	nominatim       nominatim
	places, geocode *maps.Client
	config          config.GoogleMaps
	logger          *logrus.Logger
	cache           locationCache
}

func NewLocationService() {
	apiKey := config.Config.GoogleMaps.GooglePlacesApiKey

	places, placesErr := maps.NewClient(maps.WithAPIKey(apiKey))
	if placesErr != nil {
		logger.Logger.Errorf("%s:%v", "new places", placesErr.Error())
		panic(placesErr)
	} else {
		logger.Logger.Infoln("Places service...OK")
	}

	geocode, geocodeErr := maps.NewClient(maps.WithAPIKey(apiKey))
	if geocodeErr != nil {
		logger.Logger.Errorf("%s: %v", "new geocode err", geocodeErr.Error())
		panic(geocodeErr)
	} else {
		logger.Logger.Infoln("Geocode service...OK")
	}

	lc := newCache()

	Location = &locationClient{newNominatimService(lc), places, geocode, config.Config.GoogleMaps, logger.Logger, lc}
}

func (l *locationClient) AutocompletePlace(searchQuery string) ([]*model.Place, error) {
	placesCache, placesCacheErr := l.cache.placesGetCache(searchQuery)
	if placesCacheErr != nil {
		return nil, placesCacheErr
	}

	if placesCache != nil {
		return (placesCache).([]*model.Place), nil
	}

	componentsFilter := map[maps.Component][]string{
		maps.ComponentCountry: {"KE"}, // TODO use session ipinfo to get country
	}

	req := &maps.PlaceAutocompleteRequest{
		Input:      searchQuery,
		Components: componentsFilter,
	}

	places, err := l.places.PlaceAutocomplete(context.Background(), req)
	if err != nil {
		placesErr := fmt.Errorf("%s:%v", "place autocomplete", err)
		l.logger.Errorf(placesErr.Error())
		return nil, placesErr
	}

	p := make([]*model.Place, 0)
	for _, item := range places.Predictions {
		place := model.Place{
			ID:            item.PlaceID,
			MainText:      item.StructuredFormatting.MainText,
			SecondaryText: item.StructuredFormatting.SecondaryText,
		}

		p = append(p, &place)
	}

	if err := l.cache.placesSetCache(searchQuery, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (l *locationClient) GeocodeLatLng(input model.GpsInput) (*Geocode, error) {
	cacheKey := util.FloatToString(input.Lat) + util.FloatToString(input.Lng)

	geocodeCache, geocodeCacheErr := l.cache.Get(cacheKey)
	if geocodeCacheErr != nil {
		return nil, geocodeCacheErr
	}

	if geocodeCache != nil {
		geo := (geocodeCache).(*Geocode)
		return geo, nil
	}

	return l.nominatim.ReverseGeocode(input)
}

func (l *locationClient) GetPlaceDetails(placeID string) (*Geocode, error) {
	placeCache, placeCacheErr := l.cache.Get(placeID)
	if placeCacheErr != nil {
		return nil, placeCacheErr
	}

	if placeCache != nil {
		return (placeCache).(*Geocode), nil
	}

	req := &maps.PlaceDetailsRequest{
		PlaceID: placeID,
		Fields:  []maps.PlaceDetailsFieldMask{"geometry"},
	}

	res, placeDetailsErr := l.places.PlaceDetails(context.Background(), req)
	if placeDetailsErr != nil {
		placeDetailsErr := fmt.Errorf("%s:%v", "get place details", placeDetailsErr)
		l.logger.Errorf(placeDetailsErr.Error())
		return nil, placeDetailsErr
	}

	placeDetails := &Geocode{
		Location: model.Gps{Lat: res.Geometry.Location.Lat, Lng: res.Geometry.Location.Lng},
	}

	if err := l.cache.Set(placeDetails.PlaceID, placeDetails); err != nil {
		return nil, err
	}

	return placeDetails, nil
}

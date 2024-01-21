package location

import (
	"context"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/util"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
)

type locationClient struct {
	nominatim       nominatim
	places, geocode *maps.Client
	config          config.GoogleMaps
	logger          *logrus.Logger
	cache           locationCache
}

var locationService Location

func GetLocationService() Location { return locationService }

func NewLocationService(cfg config.GoogleMaps, logger *logrus.Logger, redis *redis.Client) Location {
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

	cache := newlocationCache(redis, logger)

	locationService = &locationClient{newNominatimService(logger, cache), places, geocode, cfg, logger, cache}

	return locationService
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

	p := make([]*model.Place, 0)
	for _, item := range places.Predictions {
		place := model.Place{}
		place.ID = item.PlaceID
		place.MainText = item.StructuredFormatting.MainText
		place.SecondaryText = item.StructuredFormatting.SecondaryText

		p = append(p, &place)
	}

	if err := l.cache.placesSetCache(searchQuery, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (l *locationClient) GeocodeLatLng(input model.GpsInput) (*model.Geocode, error) {
	cacheKey := util.FloatToString(input.Lat) + util.FloatToString(input.Lng)

	geocodeCache, geocodeCacheErr := l.cache.Get(cacheKey)
	if geocodeCacheErr != nil {
		return nil, geocodeCacheErr
	}

	if geocodeCache != nil {
		geo := (geocodeCache).(*model.Geocode)
		return geo, nil
	}

	return l.nominatim.ReverseGeocode(input)
}

func (l *locationClient) GetPlaceDetails(placeID string) (*model.Geocode, error) {
	/*
		placeCache, placeCacheErr := l.cache.Get(placeID)
		if placeCacheErr != nil {
			return nil, placeCacheErr
		}

		if placeCache != nil {
			return (placeCache).(*model.Geocode), nil
		}
	*/

	req := &maps.PlaceDetailsRequest{
		PlaceID: placeID,
		Fields:  []maps.PlaceDetailsFieldMask{"adr_address", "name", "place_id", "formatted_address", "geometry"},
	}

	res, placeDetailsErr := l.places.PlaceDetails(context.Background(), req)
	if placeDetailsErr != nil {
		placeDetailsErr := model.UziErr{Err: placeDetailsErr.Error(), Message: "placedetails", Code: 500}
		l.logger.Errorf(placeDetailsErr.Error())
		return nil, placeDetailsErr
	}

	placeDetails := &model.Geocode{
		PlaceID:          res.PlaceID,
		FormattedAddress: res.FormattedAddress,
		Location:         model.Gps{Lat: res.Geometry.Location.Lat, Lng: res.Geometry.Location.Lng},
	}

	if err := l.cache.Set(placeDetails.PlaceID, placeDetails); err != nil {
		return nil, err
	}

	return placeDetails, nil
}

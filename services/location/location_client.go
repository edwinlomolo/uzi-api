package location

import (
	"context"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/internal/cache"
	"github.com/3dw1nM0535/uzi-api/internal/logger"
	"github.com/3dw1nM0535/uzi-api/internal/util"
	"github.com/3dw1nM0535/uzi-api/model"
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

func NewLocationService() {
	log := logger.GetLogger()
	apiKey := config.GetConfig().GoogleMaps.GooglePlacesApiKey
	c := cache.GetCache()

	places, placesErr := maps.NewClient(maps.WithAPIKey(apiKey))
	if placesErr != nil {
		log.Errorf("%s:%v", "new places", placesErr.Error())
	} else {
		log.Infoln("Places service...OK")
	}

	geocode, geocodeErr := maps.NewClient(maps.WithAPIKey(apiKey))
	if geocodeErr != nil {
		log.Errorf("%s: %v", "new geocode err", geocodeErr.Error())
	} else {
		log.Infoln("Geocode service...OK")
	}

	lc := newlocationCache(c, log)

	locationService = &locationClient{newNominatimService(lc), places, geocode, config.GetConfig().GoogleMaps, log, lc}
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
		placesErr := model.UziErr{Err: err.Error(), Message: "placeautocomplete", Code: 500}
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
	placeCache, placeCacheErr := l.cache.Get(placeID)
	if placeCacheErr != nil {
		return nil, placeCacheErr
	}

	if placeCache != nil {
		return (placeCache).(*model.Geocode), nil
	}

	req := &maps.PlaceDetailsRequest{
		PlaceID: placeID,
		Fields:  []maps.PlaceDetailsFieldMask{"geometry"},
	}

	res, placeDetailsErr := l.places.PlaceDetails(context.Background(), req)
	if placeDetailsErr != nil {
		placeDetailsErr := model.UziErr{Err: placeDetailsErr.Error(), Message: "placedetails", Code: 500}
		l.logger.Errorf(placeDetailsErr.Error())
		return nil, placeDetailsErr
	}

	placeDetails := &model.Geocode{
		Location: model.Gps{Lat: res.Geometry.Location.Lat, Lng: res.Geometry.Location.Lng},
	}

	if err := l.cache.Set(placeDetails.PlaceID, placeDetails); err != nil {
		return nil, err
	}

	return placeDetails, nil
}

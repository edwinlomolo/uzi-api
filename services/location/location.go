package location

import (
	"context"
	"fmt"
	"time"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal/cache"
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
	cache           cache.Cache
}

func NewLocationService() {
	apiKey := config.Config.GoogleMaps.GooglePlacesApiKey

	places, placesErr := maps.NewClient(maps.WithAPIKey(apiKey))
	if placesErr != nil {
		logger.Logger.Errorf("%s:%v", "new places", placesErr.Error())
		logger.Logger.Fatal(placesErr)
	} else {
		logger.Logger.Infoln("Places service...OK")
	}

	geocode, geocodeErr := maps.NewClient(maps.WithAPIKey(apiKey))
	if geocodeErr != nil {
		logger.Logger.Errorf("%s: %v", "new geocode", geocodeErr.Error())
		logger.Logger.Fatal(geocodeErr)
	} else {
		logger.Logger.Infoln("Geocode service...OK")
	}

	Location = &locationClient{
		newNominatimService(cache.Rdb),
		places,
		geocode,
		config.Config.GoogleMaps,
		logger.Logger,
		cache.Rdb,
	}
}

func (l *locationClient) AutocompletePlace(
	searchQuery string,
) ([]*model.Place, error) {
	var pls []*model.Place
	cacheKey := util.Base64Key(searchQuery)
	placesCache, placesCacheErr := l.cache.Get(context.Background(), cacheKey, &[]*model.Place{})
	if placesCacheErr != nil {
		return nil, placesCacheErr
	}

	if placesCache != nil {
		v := (placesCache).(*[]*model.Place)
		return *v, nil
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

	for _, item := range places.Predictions {
		place := model.Place{
			ID:            item.PlaceID,
			MainText:      item.StructuredFormatting.MainText,
			SecondaryText: item.StructuredFormatting.SecondaryText,
		}

		pls = append(pls, &place)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		l.cache.Set(context.Background(), cacheKey, pls, time.Hour*24)
	}()
	<-done

	return pls, nil
}

func (l *locationClient) GeocodeLatLng(
	input model.GpsInput,
) (*Geocode, error) {
	var geo *Geocode
	cacheKey := util.FloatToString(input.Lat) + util.FloatToString(input.Lng)

	geocodeCache, geocodeCacheErr := l.cache.Get(context.Background(), cacheKey, geo)
	if geocodeCacheErr != nil {
		return nil, geocodeCacheErr
	}

	if geocodeCache != nil {
		v := (geocodeCache).(**Geocode)
		return *v, nil
	}

	return l.nominatim.ReverseGeocode(input)
}

func (l *locationClient) GetPlaceDetails(
	placeID string,
) (*Geocode, error) {
	var placeDetails *Geocode
	placeCache, placeCacheErr := l.cache.Get(context.Background(), placeID, &Geocode{})
	if placeCacheErr != nil {
		return nil, placeCacheErr
	}

	if placeCache != nil {
		p := (placeCache).(*Geocode)
		return p, nil
	}

	req := &maps.PlaceDetailsRequest{
		PlaceID: placeID,
		Fields:  []maps.PlaceDetailsFieldMask{"geometry"},
	}

	res, resErr := l.places.PlaceDetails(context.Background(), req)
	if resErr != nil {
		uziErr := fmt.Errorf("%s:%v", "place details", resErr)
		l.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	placeDetails = &Geocode{
		Location: model.Gps{
			Lat: res.Geometry.Location.Lat,
			Lng: res.Geometry.Location.Lng,
		},
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		l.cache.Set(context.Background(), placeDetails.PlaceID, placeDetails, time.Hour*24)
	}()
	<-done

	return placeDetails, nil
}

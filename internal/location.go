package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
)

const nominatimApi = "https://nominatim.openstreetmap.org"

var (
	lctn LocationService
)

type LocationService interface {
	GeocodeLatLng(input model.GpsInput) (*model.Geocode, error)
	AutocompletePlace(query string) ([]*model.Place, error)
	GetPlaceDetails(placeID string) (*model.Geocode, error)
}

type locationClient struct {
	nominatim       nominatim
	places, geocode *maps.Client
	config          config.Google
	log             *logrus.Logger
}

func NewLocationService() {
	places, placesErr := maps.NewClient(maps.WithAPIKey(config.Config.Google.GooglePlacesApiKey))
	if placesErr != nil {
		log.WithError(placesErr).Errorf("new places client")
	}

	geocode, geocodeErr := maps.NewClient(maps.WithAPIKey(config.Config.Google.GoogleGeocodeApiKey))
	if geocodeErr != nil {
		log.WithError(geocodeErr).Errorf("new geocode client")
	}

	lctn = &locationClient{
		newNominatim(),
		places,
		geocode,
		config.Config.Google,
		log,
	}
}

func GetLocationService() LocationService {
	return lctn
}

func (l *locationClient) AutocompletePlace(
	searchQuery string,
) ([]*model.Place, error) {
	var pls []*model.Place
	cacheKey := base64Key(searchQuery)
	placesCache, placesCacheErr := c.Get(context.Background(), cacheKey, &[]*model.Place{})
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
		l.log.WithFields(logrus.Fields{
			"search_query": searchQuery,
			"error":        err,
		}).Errorf("place autocomplete")
		return nil, err
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
		c.Set(context.Background(), cacheKey, pls, time.Hour*24)
	}()
	<-done

	return pls, nil
}

func (l *locationClient) GeocodeLatLng(
	input model.GpsInput,
) (*model.Geocode, error) {
	var geo *model.Geocode
	cacheKey := base64Key(input)

	geocodeCache, geocodeCacheErr := c.Get(context.Background(), cacheKey, geo)
	if geocodeCacheErr != nil {
		return nil, geocodeCacheErr
	}

	if geocodeCache != nil {
		v := (geocodeCache).(**model.Geocode)
		return *v, nil
	}

	return l.nominatim.ReverseGeocode(input)
}

func (l *locationClient) GetPlaceDetails(
	placeID string,
) (*model.Geocode, error) {
	var placeDetails *model.Geocode
	placeCache, placeCacheErr := c.Get(context.Background(), placeID, &model.Geocode{})
	if placeCacheErr != nil {
		return nil, placeCacheErr
	}

	if placeCache != nil {
		p := (placeCache).(*model.Geocode)
		return p, nil
	}

	req := &maps.PlaceDetailsRequest{
		PlaceID: placeID,
		Fields:  []maps.PlaceDetailsFieldMask{"geometry"},
	}

	res, resErr := l.places.PlaceDetails(context.Background(), req)
	if resErr != nil {
		uziErr := fmt.Errorf("%s:%v", "place details", resErr)
		l.log.WithFields(logrus.Fields{
			"error":    resErr,
			"place_id": placeID,
		}).Errorf("place details")
		return nil, uziErr
	}

	placeDetails = &model.Geocode{
		Location: model.Gps{
			Lat: res.Geometry.Location.Lat,
			Lng: res.Geometry.Location.Lng,
		},
	}

	go func() {
		c.Set(context.Background(), placeDetails.PlaceID, placeDetails, time.Hour*24)
	}()

	return placeDetails, nil
}

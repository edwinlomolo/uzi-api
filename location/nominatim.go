package location

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/edwinlomolo/uzi-api/cache"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/sirupsen/logrus"
)

type nominatimresponse struct {
	PlaceID     int      `json:"place_id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	BoundingBox []string `json:"boundingbox"`
	Type        string   `json:"type"`
	Address     address  `json:"address"`
}

type address struct {
	Village     string `json:"village,omitempty"`
	County      string `json:"state,omitempty"`
	Region      string `json:"region,omitempty"`
	City        string `json:"city,omitempty"`
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
}

type nominatim interface {
	ReverseGeocode(model.GpsInput) (*Geocode, error)
}

type nominatimClient struct {
	logger *logrus.Logger
	cache  cache.Cache
}

func newNominatimService(cache cache.Cache) nominatim {
	return &nominatimClient{
		logger.Logger,
		cache,
	}
}

func (n nominatimClient) ReverseGeocode(
	input model.GpsInput,
) (*Geocode, error) {
	cacheKey := base64Key(input)

	var nominatimRes nominatimresponse
	geo := &Geocode{}

	cValue, err := n.cache.Get(context.Background(), cacheKey, &Geocode{})
	if err != nil {
		return nil, err
	}

	if cValue != nil {
		n := (cValue).(*Geocode)
		return n, nil
	}

	url := fmt.Sprintf(
		"%s/reverse?format=jsonv2&lat=%f&lon=%f",
		nominatimApi,
		input.Lat,
		input.Lng,
	)

	res, err := http.Get(url)
	if err != nil {
		uziErr := fmt.Errorf("%s:%v", "reverse geocode", err)
		n.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&nominatimRes); err != nil {
		uziErr := fmt.Errorf("%s:%v", "unmarshal", err)
		n.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	geo.PlaceID = strconv.Itoa(nominatimRes.PlaceID)
	if nominatimRes.Name == "" {
		geo.FormattedAddress = nominatimRes.DisplayName
	} else {
		geo.FormattedAddress = nominatimRes.Name
	}

	lat, parseErr := strconv.ParseFloat(nominatimRes.Lat, 64)
	if parseErr != nil {
		uziErr := fmt.Errorf("%s:%v", "parse lat", parseErr)
		n.logger.Errorf(uziErr.Error())
		return nil, err
	}
	lng, parseErr := strconv.ParseFloat(nominatimRes.Lon, 64)
	if parseErr != nil {
		uziErr := fmt.Errorf("%s:%v", "parse lng", parseErr)
		n.logger.Errorf(uziErr.Error())
		return nil, err
	}

	geo.Location = model.Gps{
		Lat: lat,
		Lng: lng,
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		n.cache.Set(context.Background(), cacheKey, geo, time.Hour)
	}()
	<-done

	return geo, nil
}
package location

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/edwinlomolo/uzi-api/gql/model"
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
}

func newNominatim() nominatim {
	return &nominatimClient{}
}

func (n nominatimClient) ReverseGeocode(
	input model.GpsInput,
) (*Geocode, error) {
	var nominatimRes nominatimresponse
	geo := &Geocode{}

	url := fmt.Sprintf(
		"%s/reverse?format=jsonv2&lat=%f&lon=%f",
		nominatimApi,
		input.Lat,
		input.Lng,
	)

	res, err := http.Get(url)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"cords": input,
		}).Errorf("reverse geocode")
		return nil, err
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&nominatimRes); err != nil {
		log.WithError(err).Errorf("unmarshal reverse geocode res")
		return nil, err
	}

	geo.PlaceID = strconv.Itoa(nominatimRes.PlaceID)
	if nominatimRes.Name == "" {
		geo.FormattedAddress = nominatimRes.DisplayName
	} else {
		geo.FormattedAddress = nominatimRes.Name
	}

	lat, parseErr := strconv.ParseFloat(nominatimRes.Lat, 64)
	if parseErr != nil {
		log.WithFields(logrus.Fields{
			"error": parseErr,
			"lat":   lat,
		}).Errorf("parse latitude")
		return nil, err
	}
	lng, parseErr := strconv.ParseFloat(nominatimRes.Lon, 64)
	if parseErr != nil {
		log.WithFields(logrus.Fields{
			"error":     parseErr,
			"longitude": lng,
		}).Errorf("parse longitude")
		return nil, err
	}

	geo.Location = model.Gps{
		Lat: lat,
		Lng: lng,
	}

	return geo, nil
}

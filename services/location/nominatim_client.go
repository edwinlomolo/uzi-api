package location

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/3dw1nM0535/uzi-api/internal/logger"
	"github.com/3dw1nM0535/uzi-api/internal/util"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/sirupsen/logrus"
)

type nominatim interface {
	ReverseGeocode(model.GpsInput) (*model.Geocode, error)
}

type nominatimClient struct {
	logger *logrus.Logger
	cache  locationCache
}

func newNominatimService(cache locationCache) nominatim {
	return &nominatimClient{logger.GetLogger(), cache}
}

func (n nominatimClient) ReverseGeocode(input model.GpsInput) (*model.Geocode, error) {
	cacheKey := util.FloatToString(input.Lat) + util.FloatToString(input.Lng)

	var nominatimRes model.NominatimResponse
	geo := &model.Geocode{}

	url := fmt.Sprintf("%s/reverse?format=jsonv2&lat=%f&lon=%f", nominatimApi, input.Lat, input.Lng)

	res, err := http.Get(url)
	if err != nil {
		uziErr := model.UziErr{Err: err.Error(), Message: "reversegeocode", Code: 500}
		n.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&nominatimRes); err != nil {
		uziErr := model.UziErr{Err: err.Error(), Message: "reversegeocodedecode", Code: 500}
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
		uziErr := model.UziErr{Err: parseErr.Error(), Message: "parselatitude", Code: 500}
		n.logger.Errorf(uziErr.Error())
		return nil, err
	}
	lng, parseErr := strconv.ParseFloat(nominatimRes.Lon, 64)
	if parseErr != nil {
		uziErr := model.UziErr{Err: parseErr.Error(), Message: "parselongitude", Code: 500}
		n.logger.Errorf(uziErr.Error())
		return nil, err
	}

	geo.Location = model.Gps{
		Lat: lat,
		Lng: lng,
	}

	if err := n.cache.Set(cacheKey, geo); err != nil {
		return nil, err
	}

	return geo, nil
}

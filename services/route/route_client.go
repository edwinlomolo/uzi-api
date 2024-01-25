package route

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/config"
	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/services/location"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/sirupsen/logrus"
)

const (
	routeV2 = "https://routes.googleapis.com/directions/v2:computeRoutes"
)

var routeService Route

type routeClient struct {
	logger *logrus.Logger
	store  *sqlStore.Queries
	config config.GoogleMaps
}

func NewRouteService(logger *logrus.Logger, store *sqlStore.Queries, config config.GoogleMaps) {
	routeService = &routeClient{logger, store, config}
}

func GetRouteService() Route { return routeService }

func (r *routeClient) ComputeTripRoute(input model.TripRouteInput) (*model.TripRoute, error) {
	pickup, pickupErr := r.parsePickupDropoff(*input.Pickup)
	if pickupErr != nil {
		return nil, pickupErr
	}

	dropoff, dropoffErr := r.parsePickupDropoff(*input.Dropoff)
	if dropoffErr != nil {
		return nil, dropoffErr
	}

	return r.computeRoute(*pickup, *dropoff)
}

func (r *routeClient) parsePickupDropoff(input model.TripInput) (*model.Geocode, error) {
	// Google place autocomplete select won't have cord in the request
	if input.Location.Lat == 0.0 && input.Location.Lng == 0.0 {
		placedetails, err := location.GetLocationService().GetPlaceDetails(input.PlaceID)
		if err != nil {
			return nil, err
		}

		return &model.Geocode{
			PlaceID:          placedetails.PlaceID,
			FormattedAddress: placedetails.FormattedAddress,
			Location:         model.Gps{Lat: placedetails.Location.Lat, Lng: placedetails.Location.Lng},
		}, nil
	}

	return &model.Geocode{
		PlaceID:          input.PlaceID,
		FormattedAddress: input.FormattedAddress,
		Location:         model.Gps{Lat: input.Location.Lat, Lng: input.Location.Lng},
	}, nil
}

func (r *routeClient) computeRoute(pickup, dropoff model.Geocode) (*model.TripRoute, error) {
	routeResponse := &model.RouteResponse{}
	routeParams := createRouteRequest(
		model.LatLng{
			Lat: pickup.Location.Lat,
			Lng: pickup.Location.Lng,
		},
		model.LatLng{
			Lat: dropoff.Location.Lat,
			Lng: dropoff.Location.Lng,
		},
	)

	reqPayload, payloadErr := json.Marshal(routeParams)
	if payloadErr != nil {
		uziErr := model.UziErr{Err: payloadErr.Error(), Message: "computeroutepayloadmarshal", Code: 500}
		r.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	req, reqErr := http.NewRequest("POST", routeV2, bytes.NewBuffer(reqPayload))
	if reqErr != nil {
		uziErr := model.UziErr{Err: reqErr.Error(), Message: "computerouterequest", Code: 500}
		r.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Goog-Api-Key", r.config.GoogleRoutesApiKey)
	req.Header.Add("X-Goog-FieldMask", "routes.duration,routes.distanceMeters,routes.polyline.encodedPolyline,routes.staticDuration")

	c := &http.Client{}
	res, resErr := c.Do(req)
	if resErr != nil {
		uziErr := model.UziErr{Err: resErr.Error(), Message: "computerouteresponse", Code: 500}
		r.logger.Errorf(uziErr.Error())
		return nil, resErr
	}

	if err := json.NewDecoder(res.Body).Decode(&routeResponse); err != nil {
		uziErr := model.UziErr{Err: err.Error(), Message: "computerouteresunmarshal", Code: 500}
		r.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	if routeResponse.Error.Code > 0 {
		uziErr := model.UziErr{Err: routeResponse.Error.Message, Message: routeResponse.Error.Status, Code: routeResponse.Error.Code}
		r.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return nil, nil
}

func createRouteRequest(pickup, dropoff model.LatLng) model.RouteRequest {
	return model.RouteRequest{
		Origin: model.Origin{
			RoutePoint: model.RoutePoint{
				Location: pickup,
			},
		},
		Destination: model.Destination{
			RoutePoint: model.RoutePoint{
				Location: dropoff,
			},
		},
		TravelMode:             "DRIVE",
		ComputeAlternateRoutes: false,
		RoutePreference:        "TRAFFIC_AWARE_OPTIMAL",
		RouteModifiers: model.RouteModifiers{
			AvoidTolls:    false,
			AvoidHighways: false,
			AvoidFerries:  false,
		},
		PolylineQuality: "HIGH_QUALITY",
		Language:        "en-US",
		Units:           "IMPERIAL",
		RegionCode:      "KE",
	}
}

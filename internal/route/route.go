package route

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal/cache"
	"github.com/edwinlomolo/uzi-api/internal/logger"
	"github.com/edwinlomolo/uzi-api/internal/util"
	"github.com/edwinlomolo/uzi-api/services/location"
	"github.com/edwinlomolo/uzi-api/services/trip"
	"github.com/edwinlomolo/uzi-api/store"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	routeV2 = "https://routes.googleapis.com/directions/v2:computeRoutes"
)

var (
	Routing         Route
	ErrComputeRoute = errors.New("compute route")
)

type Route interface {
	ComputeTripRoute(input model.TripRouteInput) (*model.TripRoute, error)
}

type routeClient struct {
	redis  *redis.Client
	logger *logrus.Logger
	store  *sqlStore.Queries
	config config.GoogleMaps
	cache  routeCache
}

func NewRouteService() {
	Routing = &routeClient{
		cache.Redis,
		logger.Logger,
		store.DB,
		config.Config.GoogleMaps,
		newCache(),
	}
}

func (r *routeClient) ComputeTripRoute(
	input model.TripRouteInput,
) (*model.TripRoute, error) {
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

func (r *routeClient) parsePickupDropoff(
	input model.TripInput,
) (*location.Geocode, error) {
	// Google place autocomplete select won't have cord in the request
	if input.Location.Lat == 0.0 && input.Location.Lng == 0.0 {
		placedetails, err := location.Location.GetPlaceDetails(input.PlaceID)
		if err != nil {
			return nil, err
		}

		return &location.Geocode{
			PlaceID:          placedetails.PlaceID,
			FormattedAddress: placedetails.FormattedAddress,
			Location: model.Gps{
				Lat: placedetails.Location.Lat,
				Lng: placedetails.Location.Lng,
			},
		}, nil
	}

	return &location.Geocode{
		PlaceID:          input.PlaceID,
		FormattedAddress: input.FormattedAddress,
		Location: model.Gps{
			Lat: input.Location.Lat,
			Lng: input.Location.Lng,
		},
	}, nil
}

func (r *routeClient) computeRoute(
	pickup,
	dropoff location.Geocode,
) (*model.TripRoute, error) {
	routeResponse := &routeresponse{}

	tripRoute := &model.TripRoute{}

	routeParams := createRouteRequest(
		latlng{
			Lat: pickup.Location.Lat,
			Lng: pickup.Location.Lng,
		},
		latlng{
			Lat: dropoff.Location.Lat,
			Lng: dropoff.Location.Lng,
		},
	)

	cacheKey := util.Base64Key(routeParams)

	tripInfo, tripInfoErr := r.cache.getRouteCache(cacheKey)
	if tripInfoErr != nil {
		return nil, tripInfoErr
	}

	if tripInfo == nil {
		routeRes, routeResErr := r.requestGoogleRoute(routeParams, routeResponse)
		if routeResErr != nil {
			return nil, routeResErr
		}

		tripRoute.Polyline = routeRes.Routes[0].Polyline.EncodedPolyline
		tripRoute.Distance = routeRes.Routes[0].Distance

		go r.cache.cacheRoute(cacheKey, tripRoute)
	} else {
		route := (tripInfo).(*model.TripRoute)
		tripRoute.Polyline = route.Polyline
		tripRoute.Distance = route.Distance
	}

	nearbyParams := sqlStore.GetNearbyAvailableCourierProductsParams{
		Point: fmt.Sprintf(
			"SRID=4326;POINT(%.8f %.8f)",
			pickup.Location.Lng,
			pickup.Location.Lat,
		),
		Radius: 2000,
	}
	nearbyProducts, nearbyErr := trip.Trip.GetNearbyAvailableProducts(
		nearbyParams,
		tripRoute.Distance,
	)
	if nearbyErr != nil {
		return nil, nearbyErr
	}
	tripRoute.AvailableProducts = nearbyProducts

	return tripRoute, nil
}

func (r *routeClient) requestGoogleRoute(
	routeParams routerequest,
	routeResponse *routeresponse,
) (*routeresponse, error) {
	reqPayload, payloadErr := json.Marshal(routeParams)
	if payloadErr != nil {
		err := fmt.Errorf("%s:%v", "marshal", payloadErr.Error())
		r.logger.Errorf(err.Error())
		return nil, err
	}

	req, reqErr := http.NewRequest("POST", routeV2, bytes.NewBuffer(reqPayload))
	if reqErr != nil {
		err := fmt.Errorf("%s:%v", "new request", reqErr.Error())
		r.logger.Errorf(err.Error())
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Goog-Api-Key", r.config.GoogleRoutesApiKey)
	req.Header.Add(
		"X-Goog-FieldMask",
		"routes.duration,routes.distanceMeters,routes.polyline.encodedPolyline,routes.staticDuration",
	)

	c := &http.Client{}
	res, resErr := c.Do(req)
	if resErr != nil {
		err := fmt.Errorf("%s:%v", ErrComputeRoute.Error(), resErr.Error())
		r.logger.Errorf(err.Error())
		return nil, err
	}

	if err := json.NewDecoder(res.Body).Decode(&routeResponse); err != nil {
		jsonErr := fmt.Errorf("%s:%v", ErrComputeRoute.Error(), err.Error())
		r.logger.Errorf(jsonErr.Error())
		return nil, jsonErr
	}

	if routeResponse.Error.Code > 0 {
		resErr := fmt.Errorf(
			"%s:%v",
			routeResponse.Error.Status,
			routeResponse.Error.Message,
		)
		r.logger.Errorf(resErr.Error())
		return nil, resErr
	}

	return routeResponse, nil
}

func createRouteRequest(pickup, dropoff latlng) routerequest {
	return routerequest{
		origin: origin{
			routepoint: routepoint{
				Location: pickup,
			},
		},
		destination: destination{
			routepoint: routepoint{
				Location: dropoff,
			},
		},
		TravelMode:             "DRIVE",
		ComputeAlternateRoutes: false,
		RoutePreference:        "TRAFFIC_AWARE_OPTIMAL",
		RouteModifiers: routemodifiers{
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

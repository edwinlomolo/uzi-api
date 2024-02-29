package routing

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/edwinlomolo/uzi-api/cache"
	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/location"
	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/edwinlomolo/uzi-api/pricer"
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
	ParsePickupDropoff(input model.TripInput) (*location.Geocode, error)
	GetNearbyAvailableProducts(
		params sqlStore.GetNearbyAvailableCourierProductsParams,
		tripDistance int,
	) ([]*model.Product, error)
}

type routeClient struct {
	redis  *redis.Client
	log    *logrus.Logger
	store  *sqlStore.Queries
	config config.GoogleMaps
	cache  cache.Cache
	mu     sync.Mutex
}

func NewRouteService() {
	Routing = &routeClient{
		cache.Redis,
		logger.Logger,
		store.DB,
		config.Config.GoogleMaps,
		cache.Rdb,
		sync.Mutex{},
	}
}

func (r *routeClient) ComputeTripRoute(
	input model.TripRouteInput,
) (*model.TripRoute, error) {
	pickup, pickupErr := r.ParsePickupDropoff(*input.Pickup)
	if pickupErr != nil {
		return nil, pickupErr
	}

	dropoff, dropoffErr := r.ParsePickupDropoff(*input.Dropoff)
	if dropoffErr != nil {
		return nil, dropoffErr
	}

	return r.computeRoute(*pickup, *dropoff)
}

func (r *routeClient) ParsePickupDropoff(
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

	cacheKey := base64Key(routeParams)

	tripInfo, tripInfoErr := r.cache.Get(context.Background(), cacheKey, tripRoute)
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

		// Let the above fallthrough and shortcircuit here not to super-charge in dev
		if config.IsDev() {
			go func() {
				r.cache.Set(context.Background(), cacheKey, tripRoute, time.Hour*24)
			}()
		}
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
	nearbyProducts, nearbyErr := r.GetNearbyAvailableProducts(
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
		r.log.WithFields(logrus.Fields{
			"route_params": routeParams,
			"error":        payloadErr,
		}).Errorf("marshal route params")
		return nil, payloadErr
	}

	req, reqErr := http.NewRequest("POST", routeV2, bytes.NewBuffer(reqPayload))
	if reqErr != nil {
		r.log.WithError(reqErr).Errorf("compute route request")
		return nil, reqErr
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
		r.log.WithError(resErr).Errorf("call google compute route api")
		return nil, resErr
	}

	if err := json.NewDecoder(res.Body).Decode(&routeResponse); err != nil {
		r.log.WithError(err).Errorf("unmarshal google compute route res")
		return nil, err
	}

	if routeResponse.Error.Code > 0 {
		resErr := fmt.Errorf(
			"%s:%v",
			routeResponse.Error.Status,
			routeResponse.Error.Message,
		)
		r.log.WithFields(logrus.Fields{
			"status":  routeResponse.Error.Status,
			"message": routeResponse.Error.Message,
		}).Errorf("google compute route res error")
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

func (r *routeClient) GetNearbyAvailableProducts(
	params sqlStore.GetNearbyAvailableCourierProductsParams,
	tripDistance int,
) ([]*model.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var nearbyProducts []*model.Product

	nearbys, nearbyErr := r.store.GetNearbyAvailableCourierProducts(
		context.Background(),
		params,
	)
	if nearbyErr == sql.ErrNoRows {
		return make([]*model.Product, 0), nil
	} else if nearbyErr != nil {
		r.log.WithError(nearbyErr).Errorf("nearby courier products")
		return nil, nearbyErr
	}

	for _, item := range nearbys {
		product := &model.Product{
			ID: item.ID_2,
			Price: pricer.Pricer.CalculateTripCost(
				int(item.WeightClass),
				tripDistance,
				item.Name != "UziX",
			),
			Name:        item.Name,
			Description: item.Description,
			IconURL:     item.Icon,
		}

		nearbyProducts = append(nearbyProducts, product)
	}

	return nearbyProducts, nil
}

func base64Key(key interface{}) string {
	keyString, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(keyString))

	return encoded
}

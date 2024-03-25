package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal"
	r "github.com/edwinlomolo/uzi-api/repository"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	tService TripController
)

type TripController interface {
	FindAvailableCourier(pickup model.GpsInput) (*model.Courier, error)
	GetCourierNearPickupPoint(pickup model.GpsInput) ([]*model.Courier, error)
	AssignCourierToTrip(tripID, courierID uuid.UUID) error
	UnassignTrip(courierID uuid.UUID) error
	CreateTrip(sqlStore.CreateTripParams) (*model.Trip, error)
	SetTripStatus(tripID uuid.UUID, status model.TripStatus) error
	MatchCourier(tripID uuid.UUID, pickup model.TripInput)
	CreateTripRecipient(tripID uuid.UUID, input model.TripRecipientInput) error
	GetTripRecipient(tripID uuid.UUID) (*model.Recipient, error)
	GetTripDetails(tripID uuid.UUID) (*model.Trip, error)
	GetCourierAssignedTrip(courierID uuid.UUID) error
	GetTripCourier(courierID uuid.UUID) (*model.Courier, error)
	ReportTripStatus(tripID uuid.UUID, status model.TripStatus) error
	ComputeTripRoute(input model.TripRouteInput) (*model.TripRoute, error)
	ParsePickupDropoff(input model.TripInput) (*model.Geocode, error)
}

type tripClient struct {
	r     *r.TripRepository
	mu    sync.Mutex
	log   *logrus.Logger
	cache internal.Cache
	p     internal.Pricing
}

func NewTripController(q *sqlc.Queries) {
	t := &r.TripRepository{}
	t.Init(q)
	tService = &tripClient{
		t,
		sync.Mutex{},
		internal.GetLogger(),
		internal.GetCache(),
		internal.GetPricer(),
	}
}

func GetTripController() TripController {
	return tService
}

func (t *tripClient) ParsePickupDropoff(input model.TripInput) (*model.Geocode, error) {
	return t.r.ParsePickupDropoff(input)
}

func (t *tripClient) ComputeTripRoute(input model.TripRouteInput) (*model.TripRoute, error) {
	pickup, pickupErr := t.r.ParsePickupDropoff(*input.Pickup)
	if pickupErr != nil {
		return nil, pickupErr
	}

	dropoff, dropoffErr := t.r.ParsePickupDropoff(*input.Dropoff)
	if dropoffErr != nil {
		return nil, dropoffErr
	}

	return t.computeRoute(*pickup, *dropoff)
}

func (t *tripClient) computeRoute(pickup, dropoff model.Geocode) (*model.TripRoute, error) {
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

	tripInfo, tripInfoErr := t.cache.Get(context.Background(), cacheKey, tripRoute)
	if tripInfoErr != nil {
		return nil, tripInfoErr
	}

	if tripInfo == nil {
		routeRes, routeResErr := t.googleRoutesApi(routeParams, routeResponse)
		if routeResErr != nil {
			return nil, routeResErr
		}

		tripRoute.Polyline = routeRes.Routes[0].Polyline.EncodedPolyline
		tripRoute.Distance = routeRes.Routes[0].Distance

		// Short-circuit google route api with cache here not to super-charge in dev
		if isDev() {
			go func() {
				t.cache.Set(context.Background(), cacheKey, tripRoute, time.Hour*24)
			}()
		}
	} else {
		route := (tripInfo).(*model.TripRoute)
		tripRoute.Polyline = route.Polyline
		tripRoute.Distance = route.Distance
	}

	nearbyParams := sqlc.GetNearbyAvailableCourierProductsParams{
		Point: fmt.Sprintf(
			"SRID=4326;POINT(%.8f %.8f)",
			pickup.Location.Lng,
			pickup.Location.Lat,
		),
		Radius: 2000,
	}
	nearbyProducts, nearbyErr := t.r.GetNearbyAvailableProducts(
		nearbyParams,
		tripRoute.Distance,
	)
	if nearbyErr != nil {
		return nil, nearbyErr
	}
	tripRoute.AvailableProducts = nearbyProducts

	return tripRoute, nil
}

func (t *tripClient) googleRoutesApi(routeParams routerequest, routeResponse *routeresponse) (*routeresponse, error) {
	reqPayload, payloadErr := json.Marshal(routeParams)
	if payloadErr != nil {
		t.log.WithFields(logrus.Fields{
			"route_params": routeParams,
		}).WithError(payloadErr).Errorf("marshal route params")
		return nil, payloadErr
	}

	req, reqErr := http.NewRequest("POST", internal.ComputeRouteApi, bytes.NewBuffer(reqPayload))
	if reqErr != nil {
		t.log.WithError(reqErr).Errorf("google route: compute route request")
		return nil, reqErr
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Goog-Api-Key", config.Config.Google.GoogleRoutesApiKey)
	req.Header.Add(
		"X-Goog-FieldMask",
		"routes.duration,routes.distanceMeters,routes.polyline.encodedPolyline,routes.staticDuration",
	)

	c := &http.Client{}
	res, resErr := c.Do(req)
	if resErr != nil {
		t.log.WithError(resErr).Errorf("call google compute route api")
		return nil, resErr
	}

	if err := json.NewDecoder(res.Body).Decode(&routeResponse); err != nil {
		t.log.WithError(err).Errorf("google route: unmarshal google route")
		return nil, err
	}

	if routeResponse.Error.Code > 0 {
		resErr := fmt.Errorf(
			"%s:%v",
			routeResponse.Error.Status,
			routeResponse.Error.Message,
		)
		t.log.WithError(resErr).Errorf("google route: compute route response")
		return nil, resErr
	}

	return routeResponse, nil
}

func (t *tripClient) FindAvailableCourier(pickup model.GpsInput) (*model.Courier, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.FindAvailableCourier(pickup)
}

func (t *tripClient) AssignCourierToTrip(tripID, courierID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Create trip cost silently
	go t.createTripCost(tripID)

	return t.r.AssignCourierToTrip(tripID, courierID)
}

func (t *tripClient) UnassignTrip(courierID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.UnassignTrip(courierID)
}

func (t *tripClient) CreateTrip(args sqlStore.CreateTripParams) (*model.Trip, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.CreateTrip(args)
}

func (t *tripClient) SetTripStatus(tripID uuid.UUID, status model.TripStatus) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.SetTripStatus(tripID, status)
}

func (t *tripClient) GetCourierNearPickupPoint(pickup model.GpsInput) ([]*model.Courier, error) {
	return t.r.GetCourierNearPickupPoint(pickup)
}

func (t *tripClient) GetCourierAssignedTrip(courierID uuid.UUID) error {
	return t.r.GetCourierAssignedTrip(courierID)
}

func (t *tripClient) CreateTripRecipient(
	tripID uuid.UUID,
	input model.TripRecipientInput,
) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.r.CreateTripRecipient(tripID, input)
}

func (t *tripClient) createTripCost(tripID uuid.UUID) error {
	trip, err := t.r.GetTrip(tripID)
	if err != nil {
		return err
	}

	pickup := trip.StartLocation
	dropoff := trip.EndLocation
	routeRes, routeErr := t.computeRoute(
		model.Geocode{
			Location: *pickup,
		},
		model.Geocode{
			Location: *dropoff,
		},
	)
	if routeErr != nil {
		return routeErr
	}

	product, err := t.r.GetTripProduct(trip.ProductID)
	if err != nil {
		return err
	}

	cost := t.p.CalculateTripCost(product.WeightClass, routeRes.Distance, product.Name != "UziX")

	return t.r.CreateTripCost(tripID, cost)
}

func (t *tripClient) GetTripRecipient(tripID uuid.UUID) (*model.Recipient, error) {
	return t.r.GetTripRecipient(tripID)
}

func (t *tripClient) GetTripDetails(tripID uuid.UUID) (*model.Trip, error) {
	trip, err := t.r.GetTrip(tripID)
	if err != nil {
		return nil, err
	}

	// Can we return trip route also?
	// TODO there is a cheapest way to do this
	// other than re-requesting google routes api
	if trip.CourierID.String() != internal.ZERO_UUID {
		pickup := model.TripInput{}
		dropoff := model.TripInput{}
		courierGps, err := t.r.GetCourierLocation(*trip.CourierID)
		if err != nil {
			return nil, err
		}
		switch trip.Status {
		case model.TripStatusCourierArriving,
			model.TripStatusCourierAssigned:
			pickup.Location = &model.GpsInput{
				Lat: courierGps.Lat,
				Lng: courierGps.Lng,
			}
			dropoff.Location = &model.GpsInput{
				Lat: trip.ConfirmedPickup.Lat,
				Lng: trip.ConfirmedPickup.Lng,
			}
			tripRoute, err := t.ComputeTripRoute(model.TripRouteInput{Pickup: &pickup, Dropoff: &dropoff})
			if err != nil {
				return nil, err
			}
			trip.Route = tripRoute
		case model.TripStatusCourierEnRoute:
			pickup.Location = &model.GpsInput{
				Lat: courierGps.Lat,
				Lng: courierGps.Lng,
			}
			dropoff.Location = &model.GpsInput{
				Lat: trip.EndLocation.Lat,
				Lng: trip.EndLocation.Lng,
			}
			tripRoute, err := t.ComputeTripRoute(model.TripRouteInput{Pickup: &pickup, Dropoff: &dropoff})
			if err != nil {
				return nil, err
			}
			trip.Route = tripRoute
		}
	}

	return trip, nil
}

func (t *tripClient) GetTripCourier(courierID uuid.UUID) (*model.Courier, error) {
	return t.r.GetTripCourier(courierID)
}

func (t *tripClient) MatchCourier(tripID uuid.UUID, pickup model.TripInput) {
	pkp, parseErr := t.r.ParsePickupDropoff(pickup)
	if parseErr != nil {
		t.log.WithFields(logrus.Fields{
			"pickup": pickup,
		}).WithError(parseErr).Errorf("cleanup trip pickup input")
	}

	ctx := context.Background()
	timeout, cancel := context.WithTimeout(ctx, time.Minute)
	go func() {
		defer cancel()
		courierFound := false

		for {
			select {
			case <-timeout.Done():
				if !courierFound {
					t.ReportTripStatus(tripID, model.TripStatusCourierNotFound)
				}
				return
			default:
				time.Sleep(500 * time.Millisecond)

				trip, err := t.r.GetTrip(tripID)
				if err != nil {
					return
				}

				if trip.Status == model.TripStatusCancelled {
					return
				}

				courier, err := t.r.FindAvailableCourier(model.GpsInput{
					Lat: pkp.Location.Lat,
					Lng: pkp.Location.Lng,
				})
				if err != nil {
					return
				}

				if courier != nil && !courierFound {
					courierFound = true
					t.ReportTripStatus(trip.ID, model.TripStatusCourierFound)

					assignErr := t.AssignCourierToTrip(trip.ID, courier.ID)
					if assignErr == nil {
						t.ReportTripStatus(tripID, model.TripStatusCourierAssigned)
						return
					} else if assignErr != nil {
						return
					}
				}
			}
		}
	}()
}

func (t *tripClient) ReportTripStatus(tripID uuid.UUID, status model.TripStatus) error {
	// Are we cancelling trip?
	switch status {
	case model.TripStatusCancelled:
		trip, err := t.r.GetTrip(tripID)
		if err != nil {
			return err
		}

		// Check courier hasn't been assigned yet
		if trip.CourierID.String() == internal.ZERO_UUID {
			t.publishTripUpdate(tripID, model.TripStatusCancelled, getTripStatusChannel(status))
		}
	default:
		t.publishTripUpdate(tripID, status, getTripStatusChannel(status))
	}

	return nil
}

// determine communication channels
func getTripStatusChannel(status model.TripStatus) []string {
	switch status {
	case model.TripStatusCourierArriving,
		model.TripStatusCourierEnRoute,
		model.TripStatusComplete,
		model.TripStatusCourierNotFound:
		return []string{internal.TRIP_UPDATES_CHANNEL}
	case model.TripStatusCourierAssigned:
		return []string{internal.ASSIGN_TRIP_CHANNEL, internal.TRIP_UPDATES_CHANNEL}
	case model.TripStatusCancelled:
		return []string{internal.ASSIGN_TRIP_CHANNEL}
	default:
		return []string{}
	}
}

func (t *tripClient) publishTripUpdate(tripID uuid.UUID, status model.TripStatus, channels []string) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		update := model.TripUpdate{ID: tripID, Status: status}

		t.r.SetTripStatus(tripID, status)

		switch status {
		case model.TripStatusCourierArriving,
			model.TripStatusCourierEnRoute,
			model.TripStatusCourierAssigned,
			model.TripStatusCancelled:
			getTrip, err := t.r.GetTrip(tripID)
			if err != nil {
				return
			}

			tripCourier, courierErr := t.r.GetTripCourier(*getTrip.CourierID)
			if courierErr != nil {
				return
			}

			switch status {
			case model.TripStatusCourierArriving, model.TripStatusCourierEnRoute:
				update.Location = &model.Gps{Lat: tripCourier.Location.Lat, Lng: tripCourier.Location.Lng}
			case model.TripStatusCourierAssigned:
				update.CourierID = getTrip.CourierID
			}
		}

		u, marshalErr := json.Marshal(update)
		if marshalErr != nil {
			t.log.WithError(marshalErr).Errorf("publish trip: marshal trip update")
			return
		}

		for _, channel := range channels {
			pubTripErr := t.cache.GetRedis().Publish(context.Background(), channel, u).Err()
			if pubTripErr != nil {
				t.log.WithError(pubTripErr).Errorf("publish trip: update")
				return
			}
		}
	}()
	<-done

	time.Sleep(3 * time.Second)

	return nil
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

func base64Key(key interface{}) string {
	keyString, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(keyString))

	return encoded
}

func isDev() bool {
	return config.Config.Server.Env == "development"
}

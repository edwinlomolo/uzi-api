package repository

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

	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal"
	sqlStore "github.com/edwinlomolo/uzi-api/store"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	TRIP_UPDATES = "trip_updates"
	ASSIGN_TRIP  = "assign_trip"
	routeV2      = "https://routes.googleapis.com/directions/v2:computeRoutes"
)

var (
	ErrCourierAlreadyAssigned = errors.New("courier has active trip")
	ErrCourierTripNotFound    = errors.New("courier trip not found")
)

type TripRepository struct {
	redis    *redis.Client
	location internal.LocationService
	cache    internal.Cache
	p        internal.Pricing
	mu       sync.Mutex
	store    *sqlc.Queries
	log      *logrus.Logger
}

func (t *TripRepository) Init() {
	pr := &PricerRepository{}
	pr.Init()
	t.redis = internal.GetCache().GetRedis()
	t.location = internal.GetLocationService()
	t.cache = internal.GetCache()
	t.p = internal.GetPricer()
	t.mu = sync.Mutex{}
	t.store = sqlStore.GetDb()
	t.log = internal.GetLogger()
}

func (t *TripRepository) FindAvailableCourier(pickup model.GpsInput) (*model.Courier, error) {
	args := sqlc.FindAvailableCourierParams{
		Point:  fmt.Sprintf("SRID=4326;POINT(%.8f %.8f)", pickup.Lng, pickup.Lat),
		Radius: 2000,
	}
	c, err := t.store.FindAvailableCourier(context.Background(), args)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		t.log.WithError(err).Errorf("find available courier")
		return nil, err
	}

	return &model.Courier{
		ID:        c.ID,
		UserID:    c.UserID.UUID,
		ProductID: c.ProductID.UUID,
		Location:  model.ParsePostgisLocation(c.Location),
	}, nil
}

func (t *TripRepository) AssignCourierToTrip(tripID, courierID uuid.UUID) error {
	err := t.getCourierAssignedTrip(courierID)
	if err != nil {
		t.log.WithFields(logrus.Fields{
			"courier_id": courierID,
			"trip_id":    tripID,
			"error":      err,
		}).Errorf("get assigned trip")
		return err
	}

	args := sqlc.AssignCourierToTripParams{
		ID: courierID,
		TripID: uuid.NullUUID{
			UUID:  tripID,
			Valid: true,
		},
	}
	_, assignCourierErr := t.store.AssignCourierToTrip(
		context.Background(),
		args,
	)
	if assignCourierErr != nil {
		t.log.WithFields(logrus.Fields{
			"courier_id": courierID,
			"trip_id":    args.TripID.UUID,
			"error":      assignCourierErr,
		}).Errorf("assign courier")
		return assignCourierErr
	}

	courierArgs := sqlc.AssignTripToCourierParams{
		ID: tripID,
		CourierID: uuid.NullUUID{
			UUID:  courierID,
			Valid: true,
		},
	}
	_, assignTripErr := t.store.AssignTripToCourier(
		context.Background(),
		courierArgs,
	)
	if assignTripErr != nil {
		t.log.WithFields(logrus.Fields{
			"courier_id": courierArgs.CourierID.UUID,
			"trip_id":    tripID,
			"error":      assignTripErr,
		}).Errorf("assign trip")
		return assignTripErr
	}

	// Calculate trip cost
	go t.CreateTripCost(tripID)

	return nil
}

func (t *TripRepository) UnassignTrip(courierID uuid.UUID) error {
	if _, err := t.store.UnassignCourierTrip(
		context.Background(),
		courierID,
	); err != nil {
		t.log.WithFields(logrus.Fields{
			"courier_id": courierID,
			"error":      err,
		}).Errorf("unassign trip")
		return err
	}

	return nil
}

func (t *TripRepository) CreateTrip(args sqlc.CreateTripParams) (*model.Trip, error) {
	createTrip, err := t.store.CreateTrip(context.Background(), args)
	if err != nil {
		uziErr := fmt.Errorf("%s:%v", "create trip", err)
		t.log.WithFields(logrus.Fields{
			"error":  err,
			"params": args,
		}).Errorf("create trip")
		return nil, uziErr
	}

	return &model.Trip{
		ID:     createTrip.ID,
		Status: model.TripStatus(createTrip.Status),
	}, nil
}

func (t *TripRepository) CreateTripCost(tripID uuid.UUID) error {
	// Get trip details assuming whenever we are calling this a trip already exists
	trip, err := t.store.GetTrip(context.Background(), tripID)
	if err != nil {
		t.log.WithFields(logrus.Fields{
			"error":   err,
			"trip_id": tripID,
		}).Errorf("get trip details for cost calculation")
		return err
	}

	product, err := t.store.GetCourierProductByID(context.Background(), trip.ProductID)
	if err != nil {
		t.log.WithFields(logrus.Fields{
			"error":      err,
			"product_id": trip.ProductID,
		}).Errorf("create trip cost: get courier product")
		return err
	}

	pickup := model.ParsePostgisLocation(trip.StartLocation)
	dropoff := model.ParsePostgisLocation(trip.EndLocation)
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

	cost := t.p.CalculateTripCost(int(product.WeightClass), routeRes.Distance, product.Name != "UziX")
	args := sqlc.CreateTripCostParams{
		ID:   tripID,
		Cost: int32(cost),
	}
	t.mu.Lock()
	if _, err := t.store.CreateTripCost(
		context.Background(),
		args,
	); err != nil {
		uziErr := fmt.Errorf("%s:%v", "trip cost", err)
		t.log.WithFields(logrus.Fields{
			"error":   err,
			"trip_id": tripID,
			"cost":    cost,
		}).Errorf(uziErr.Error())
		return uziErr
	}
	t.mu.Unlock()

	return nil
}

func (t *TripRepository) SetTripStatus(tripID uuid.UUID, status model.TripStatus) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	tripArgs := sqlc.SetTripStatusParams{
		ID:     tripID,
		Status: status.String(),
	}
	if _, err := t.store.SetTripStatus(
		context.Background(),
		tripArgs); err != nil {
		uziErr := fmt.Errorf("%s:%v", "trip status", err)
		t.log.WithFields(logrus.Fields{
			"trip_id": tripID,
			"status":  status.String(),
			"error":   err,
		}).Errorf("trip status")
		return uziErr
	}

	return nil
}

func (t *TripRepository) GetCourierNearPickupPoint(pickup model.GpsInput) ([]*model.Courier, error) {
	var couriers []*model.Courier

	args := sqlc.GetCourierNearPickupPointParams{
		Point: fmt.Sprintf(
			"SRID=4326;POINT(%.8f %.8f)",
			pickup.Lng,
			pickup.Lat,
		),
		Radius: 2000,
	}
	foundCouriers, err := t.store.GetCourierNearPickupPoint(
		context.Background(),
		args,
	)
	if err == sql.ErrNoRows {
		return make([]*model.Courier, 0), nil
	} else if err != nil {
		t.log.WithError(err).Errorf("courier new pickup point")
		return nil, err
	}

	for _, item := range foundCouriers {
		courier := &model.Courier{
			ID:        item.ID,
			ProductID: item.ProductID.UUID,
			Location:  model.ParsePostgisLocation(item.Location),
		}

		couriers = append(couriers, courier)
	}

	return couriers, nil
}

func (t *TripRepository) getCourierAssignedTrip(courierID uuid.UUID) error {
	_, err := t.store.GetCourierAssignedTrip(
		context.Background(),
		courierID,
	)
	if err == sql.ErrNoRows {
		return nil
	}

	return err
}

func (t *TripRepository) GetCourierAssignedTrip(courierID uuid.UUID) error {
	return t.getCourierAssignedTrip(courierID)
}

func (t *TripRepository) ParsePickupDropoff(input model.TripInput) (*model.Geocode, error) {
	// Google place autocomplete select won't have cord in the request
	if input.Location.Lat == 0.0 && input.Location.Lng == 0.0 {
		placedetails, err := t.location.GetPlaceDetails(input.PlaceID)
		if err != nil {
			return nil, err
		}

		return &model.Geocode{
			PlaceID:          placedetails.PlaceID,
			FormattedAddress: placedetails.FormattedAddress,
			Location: model.Gps{
				Lat: placedetails.Location.Lat,
				Lng: placedetails.Location.Lng,
			},
		}, nil
	}

	return &model.Geocode{
		PlaceID:          input.PlaceID,
		FormattedAddress: input.FormattedAddress,
		Location: model.Gps{
			Lat: input.Location.Lat,
			Lng: input.Location.Lng,
		},
	}, nil

}

func (t *TripRepository) MatchCourier(tripID uuid.UUID, pickup model.TripInput) {
	pkp, parseErr := t.ParsePickupDropoff(pickup)
	if parseErr != nil {
		t.log.WithFields(logrus.Fields{
			"pickup": pickup,
			"error":  parseErr,
		}).Errorf("cleanup trip pickup input")
	}

	go func() {
		courierFound := false

		for {
			select {
			case <-time.After(time.Minute):
				if !courierFound {
					t.ReportTripStatus(tripID, model.TripStatusCourierNotFound)
				}
				return
			default:
				time.Sleep(500 * time.Millisecond)

				trip, err := t.GetTrip(tripID)
				if err != nil {
					return
				}

				if trip.Status == model.TripStatusCancelled {
					return
				}

				courier, err := t.FindAvailableCourier(model.GpsInput{
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

func (t *TripRepository) CreateTripRecipient(tripID uuid.UUID, input model.TripRecipientInput) error {
	rArgs := sqlc.CreateRecipientParams{
		Name:     input.Name,
		Phone:    input.Phone,
		TripNote: input.TripNote,
		Building: sql.NullString{
			String: *input.BuildingName,
			Valid:  true,
		},
		Unit: sql.NullString{
			String: *input.UnitName,
			Valid:  true,
		},
		TripID: uuid.NullUUID{
			UUID:  tripID,
			Valid: true,
		},
	}
	if _, err := t.store.CreateRecipient(context.Background(), rArgs); err != nil {
		t.log.WithFields(logrus.Fields{
			"error":     err,
			"recipient": rArgs,
		}).Errorf("create trip recipient")
		return err
	}

	return nil
}

func (t *TripRepository) GetTripRecipient(tripID uuid.UUID) (*model.Recipient, error) {
	r, err := t.store.GetTripRecipient(
		context.Background(),
		uuid.NullUUID{
			UUID:  tripID,
			Valid: true,
		},
	)
	if err != nil {
		t.log.WithFields(logrus.Fields{
			"trip_id": tripID,
			"error":   err,
		}).Errorf("get trip recipient")
		return nil, err
	}

	return &model.Recipient{
		ID:           r.ID,
		Phone:        r.Phone,
		Name:         r.Name,
		BuildingName: &r.Building.String,
		UnitName:     &r.Unit.String,
		TripID:       r.TripID.UUID,
	}, nil
}

func (t *TripRepository) getCourierProduct(productID uuid.UUID) (*model.Product, error) {
	product, err := t.store.GetCourierProductByID(
		context.Background(),
		productID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		t.log.WithFields(logrus.Fields{
			"product_id": productID,
			"error":      err,
		}).Errorf("get courier product")
		return nil, err
	}

	return &model.Product{
		ID:          product.ID,
		IconURL:     product.Icon,
		Name:        product.Name,
		WeightClass: int(product.WeightClass),
	}, nil
}

func (t *TripRepository) GetTrip(tripID uuid.UUID) (*model.Trip, error) {
	trip, err := t.store.GetTrip(context.Background(), tripID)
	if err != nil {
		t.log.WithFields(logrus.Fields{
			"trip_id": tripID,
			"error":   err,
		}).Errorf("get trip")
		return nil, err
	}

	trp := &model.Trip{
		ID:          trip.ID,
		Status:      model.TripStatus(trip.Status),
		CourierID:   &trip.CourierID.UUID,
		ProductID:   trip.ProductID,
		Cost:        int(trip.Cost),
		EndLocation: model.ParsePostgisLocation(trip.EndLocation),
	}

	// Return trip route also
	if trp.CourierID.String() != internal.ZERO_UUID {
		pickup := model.TripInput{}
		dropoff := model.TripInput{}
		courierGps, err := t.store.GetCourierLocation(context.Background(), *trp.CourierID)
		if err != nil {
			return nil, err
		}
		switch trp.Status {
		case model.TripStatusCourierArriving,
			model.TripStatusCourierAssigned:
			pickup.Location = &model.GpsInput{
				Lat: model.ParsePostgisLocation(courierGps).Lat,
				Lng: model.ParsePostgisLocation(courierGps).Lng,
			}
			dropoff.Location = &model.GpsInput{
				Lat: model.ParsePostgisLocation(trip.ConfirmedPickup).Lat,
				Lng: model.ParsePostgisLocation(trip.ConfirmedPickup).Lng,
			}
			tripRoute, err := t.ComputeTripRoute(model.TripRouteInput{Pickup: &pickup, Dropoff: &dropoff})
			if err != nil {
				return nil, err
			}
			trp.Route = tripRoute
		case model.TripStatusCourierEnRoute:
			pickup.Location = &model.GpsInput{
				Lat: model.ParsePostgisLocation(courierGps).Lat,
				Lng: model.ParsePostgisLocation(courierGps).Lng,
			}
			dropoff.Location = &model.GpsInput{
				Lat: trp.EndLocation.Lat,
				Lng: trp.EndLocation.Lng,
			}
			tripRoute, err := t.ComputeTripRoute(model.TripRouteInput{Pickup: &pickup, Dropoff: &dropoff})
			if err != nil {
				return nil, err
			}
			trp.Route = tripRoute
		}
	}

	return trp, nil
}

func (t *TripRepository) publishTripUpdate(tripID uuid.UUID, status model.TripStatus, channels []string) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		update := model.TripUpdate{ID: tripID, Status: status}

		t.SetTripStatus(tripID, status)

		switch status {
		case model.TripStatusCourierArriving,
			model.TripStatusCourierEnRoute,
			model.TripStatusCourierAssigned,
			model.TripStatusCancelled:
			getTrip, err := t.GetTrip(tripID)
			if err != nil {
				return
			}

			tripCourier, courierErr := t.GetTripCourier(*getTrip.CourierID)
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
			t.log.WithError(marshalErr).Errorf("marshal trip update")
			return
		}

		for _, channel := range channels {
			pubTripErr := t.redis.Publish(context.Background(), channel, u).Err()
			if pubTripErr != nil {
				t.log.WithError(pubTripErr).Errorf("redis publish trip update")
				return
			}
		}
	}()
	<-done

	time.Sleep(3 * time.Second)

	return nil
}

func (t *TripRepository) GetTripCourier(courierID uuid.UUID) (*model.Courier, error) {
	courier, err := t.store.GetCourierByID(context.Background(), courierID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		t.log.WithFields(logrus.Fields{
			"courier_id": courierID,
			"error":      err,
		}).Errorf("get trip courier")
		return nil, err
	}

	return &model.Courier{
		ID:       courier.ID,
		TripID:   &courier.TripID.UUID,
		UserID:   courier.UserID.UUID,
		Location: model.ParsePostgisLocation(courier.Location),
	}, nil
}

func (t *TripRepository) ReportTripStatus(tripID uuid.UUID, status model.TripStatus) error {
	// Are we cancelling trip?
	switch status {
	case model.TripStatusCancelled:
		trip, err := t.GetTrip(tripID)
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
		model.TripStatusComplete:
		return []string{TRIP_UPDATES}
	case model.TripStatusCourierAssigned:
		return []string{ASSIGN_TRIP, TRIP_UPDATES}
	case model.TripStatusCancelled:
		return []string{ASSIGN_TRIP}
	default:
		return []string{}
	}
}

func (t *TripRepository) ComputeTripRoute(input model.TripRouteInput) (*model.TripRoute, error) {
	pickup, pickupErr := t.ParsePickupDropoff(*input.Pickup)
	if pickupErr != nil {
		return nil, pickupErr
	}

	dropoff, dropoffErr := t.ParsePickupDropoff(*input.Dropoff)
	if dropoffErr != nil {
		return nil, dropoffErr
	}

	return t.computeRoute(*pickup, *dropoff)
}

func (t *TripRepository) computeRoute(pickup, dropoff model.Geocode) (*model.TripRoute, error) {
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
		routeRes, routeResErr := t.requestGoogleRoute(routeParams, routeResponse)
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
	nearbyProducts, nearbyErr := t.GetNearbyAvailableProducts(
		nearbyParams,
		tripRoute.Distance,
	)
	if nearbyErr != nil {
		return nil, nearbyErr
	}
	tripRoute.AvailableProducts = nearbyProducts

	return tripRoute, nil
}

func (t *TripRepository) requestGoogleRoute(routeParams routerequest, routeResponse *routeresponse) (*routeresponse, error) {
	reqPayload, payloadErr := json.Marshal(routeParams)
	if payloadErr != nil {
		t.log.WithFields(logrus.Fields{
			"route_params": routeParams,
			"error":        payloadErr,
		}).Errorf("marshal route params")
		return nil, payloadErr
	}

	req, reqErr := http.NewRequest("POST", routeV2, bytes.NewBuffer(reqPayload))
	if reqErr != nil {
		t.log.WithError(reqErr).Errorf("compute route request")
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
		t.log.WithError(err).Errorf("unmarshal google compute route res")
		return nil, err
	}

	if routeResponse.Error.Code > 0 {
		resErr := fmt.Errorf(
			"%s:%v",
			routeResponse.Error.Status,
			routeResponse.Error.Message,
		)
		t.log.WithFields(logrus.Fields{
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

func base64Key(key interface{}) string {
	keyString, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(keyString))

	return encoded
}

func (t *TripRepository) getNearbyAvailableCourierProducts(params sqlc.GetNearbyAvailableCourierProductsParams) ([]*model.Product, error) {
	var nearbyProducts []*model.Product

	nearbys, err := t.store.GetNearbyAvailableCourierProducts(context.Background(), params)
	if err == sql.ErrNoRows {
		return make([]*model.Product, 0), nil
	} else if err != nil {
		t.log.WithFields(logrus.Fields{
			"error": err,
			"args":  params,
		}).Errorf("nearby courier products")
		return nil, err
	}

	for _, item := range nearbys {
		product := &model.Product{
			ID:          item.ID_2,
			Name:        item.Name,
			WeightClass: int(item.WeightClass),
			Description: item.Description,
			IconURL:     item.Icon,
		}

		nearbyProducts = append(nearbyProducts, product)
	}
	return nearbyProducts, nil
}

func (t *TripRepository) GetNearbyAvailableProducts(params sqlc.GetNearbyAvailableCourierProductsParams, tripDistance int) ([]*model.Product, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	nearbys, nearbyErr := t.getNearbyAvailableCourierProducts(params)
	if nearbyErr != nil {
		return nil, nearbyErr
	}

	for _, item := range nearbys {
		item.Price = t.p.CalculateTripCost(
			int(item.WeightClass),
			tripDistance,
			item.Name != "UziX",
		)
	}

	return nearbys, nil
}

func isDev() bool {
	return config.Config.Server.Env == "development"
}

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

	"github.com/edwinlomolo/uzi-api/cache"
	"github.com/edwinlomolo/uzi-api/config"
	"github.com/edwinlomolo/uzi-api/constants"
	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/location"
	"github.com/edwinlomolo/uzi-api/pricer"
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
	store    *sqlc.Queries
	redis    *redis.Client
	location location.LocationService
	cache    cache.Cache
	p        pricer.Pricing
	mu       sync.Mutex
}

func (t *TripRepository) Init(store *sqlc.Queries, cache cache.Cache) {
	pr := &PricerRepository{}
	pr.Init(store, cache)
	t.store = store
	t.redis = cache.Redis()
	t.location = location.New(cache)
	t.cache = cache
	t.p = pricer.New()
	t.mu = sync.Mutex{}
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
		log.WithError(err).Errorf("find available courier")
		return nil, err
	}

	return &model.Courier{
		ID:        c.ID,
		UserID:    c.UserID.UUID,
		ProductID: c.ProductID.UUID,
		Location:  location.ParsePostgisLocation(c.Location),
	}, nil
}

func (t *TripRepository) AssignCourierToTrip(tripID, courierID uuid.UUID) error {
	err := t.getCourierAssignedTrip(courierID)
	if err != nil {
		log.WithFields(logrus.Fields{
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
		log.WithFields(logrus.Fields{
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
		log.WithFields(logrus.Fields{
			"courier_id": courierArgs.CourierID.UUID,
			"trip_id":    tripID,
			"error":      assignTripErr,
		}).Errorf("assign trip")
		return assignTripErr
	}

	return nil
}

func (t *TripRepository) UnassignTrip(courierID uuid.UUID) error {
	if _, err := t.store.UnassignCourierTrip(
		context.Background(),
		courierID,
	); err != nil {
		log.WithFields(logrus.Fields{
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
		log.WithFields(logrus.Fields{
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

func (t *TripRepository) CreateTripCost(tripID uuid.UUID, cost int) error {
	args := sqlc.CreateTripCostParams{
		ID:   tripID,
		Cost: sql.NullInt32{Int32: int32(cost), Valid: true},
	}
	if _, err := t.store.CreateTripCost(
		context.Background(),
		args,
	); err != nil {
		uziErr := fmt.Errorf("%s:%v", "trip cost", err)
		log.WithFields(logrus.Fields{
			"error":   err,
			"trip_id": tripID,
			"cost":    cost,
		}).Errorf(uziErr.Error())
		return uziErr
	}

	return nil
}

func (t *TripRepository) SetTripStatus(tripID uuid.UUID, status model.TripStatus) error {
	tripArgs := sqlc.SetTripStatusParams{
		ID:     tripID,
		Status: status.String(),
	}
	if _, err := t.store.SetTripStatus(
		context.Background(),
		tripArgs); err != nil {
		uziErr := fmt.Errorf("%s:%v", "trip status", err)
		log.WithFields(logrus.Fields{
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
		log.WithError(err).Errorf("courier new pickup point")
		return nil, err
	}

	for _, item := range foundCouriers {
		courier := &model.Courier{
			ID:        item.ID,
			ProductID: item.ProductID.UUID,
			Location:  location.ParsePostgisLocation(item.Location),
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

func (t *TripRepository) ParsePickupDropoff(input model.TripInput) (*location.Geocode, error) {
	// Google place autocomplete select won't have cord in the request
	if input.Location.Lat == 0.0 && input.Location.Lng == 0.0 {
		placedetails, err := t.location.GetPlaceDetails(input.PlaceID)
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

func (t *TripRepository) MatchCourier(tripID uuid.UUID, pickup model.TripInput) {
	pkp, parseErr := t.ParsePickupDropoff(pickup)
	if parseErr != nil {
		log.WithFields(logrus.Fields{
			"pickup": pickup,
			"error":  parseErr,
		}).Errorf("cleanup trip pickup input")
	}

	// use a 5/10/15 minute timeout - trick impatiency cancellation from client(user)
	timeoutCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Minute*5,
	)

	go func() {
		defer cancel()
		courierFound := false

		for {
			select {
			case <-timeoutCtx.Done():
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
		log.WithFields(logrus.Fields{
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
		log.WithFields(logrus.Fields{
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

func (t *TripRepository) getTripCost(trip model.Trip, distance int) (int, error) {
	if trip.CourierID.String() == constants.ZERO_UUID {
		return 0, nil
	}

	courier, err := t.GetTripCourier(*trip.CourierID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":    err,
			"distance": distance,
		}).Errorf("get trip courier for trip cost calculation")
		return 0, err
	}

	product, productErr := t.getCourierProduct(courier.ProductID)
	if productErr != nil {
		log.WithFields(logrus.Fields{
			"error":              productErr,
			"courier_product_id": courier.ProductID,
		}).Errorf("get trip courier product for trip cost calculation")
		return 0, productErr
	}

	return t.p.CalculateTripCost(int(product.WeightClass), distance, product.Name != "UziX"), nil
}

func (t *TripRepository) getCourierProduct(productID uuid.UUID) (*model.Product, error) {
	product, err := t.store.GetCourierProductByID(
		context.Background(),
		productID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		log.WithFields(logrus.Fields{
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
		log.WithFields(logrus.Fields{
			"trip_id": tripID,
			"error":   err,
		}).Errorf("get trip")
		return nil, err
	}

	trp := &model.Trip{
		ID:          trip.ID,
		Status:      model.TripStatus(trip.Status),
		CourierID:   &trip.CourierID.UUID,
		Cost:        int(trip.Cost.Int32),
		EndLocation: location.ParsePostgisLocation(trip.EndLocation),
	}

	// Return trip route also
	if trp.CourierID.String() != constants.ZERO_UUID {
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
				Lat: location.ParsePostgisLocation(courierGps).Lat,
				Lng: location.ParsePostgisLocation(courierGps).Lng,
			}
			dropoff.Location = &model.GpsInput{
				Lat: location.ParsePostgisLocation(trip.ConfirmedPickup).Lat,
				Lng: location.ParsePostgisLocation(trip.ConfirmedPickup).Lng,
			}
			tripRoute, err := t.ComputeTripRoute(model.TripRouteInput{Pickup: &pickup, Dropoff: &dropoff})
			if err != nil {
				return nil, err
			}
			trp.Route = tripRoute
			cost, costErr := t.getTripCost(*trp, trp.Route.Distance)
			if costErr != nil {
				return nil, costErr
			}
			trp.Cost = cost
		case model.TripStatusCourierEnRoute:
			pickup.Location = &model.GpsInput{
				Lat: location.ParsePostgisLocation(courierGps).Lat,
				Lng: location.ParsePostgisLocation(courierGps).Lng,
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
			cost, costErr := t.getTripCost(*trp, trp.Route.Distance)
			if costErr != nil {
				return nil, costErr
			}
			trp.Cost = cost

			// Create cost while en-route. There has to be a better way to do costing?
			go func() {
				_, err := t.store.CreateTripCost(context.Background(), sqlc.CreateTripCostParams{
					ID:   tripID,
					Cost: sql.NullInt32{Int32: int32(trp.Cost), Valid: true},
				})
				log.WithFields(logrus.Fields{
					"error":   err,
					"trip_id": tripID,
					"cost":    trp.Cost,
				}).Errorf("create trip cost")
				if err != nil {
					return
				}
			}()
		}
	}

	return trp, nil
}

func (t *TripRepository) publishTripUpdate(tripID uuid.UUID, status model.TripStatus, channel string) error {
	time.Sleep(5 * time.Second)

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
			log.WithError(marshalErr).Errorf("marshal trip update")
			return
		}

		pubTripErr := t.redis.Publish(context.Background(), channel, u).Err()
		if pubTripErr != nil {
			log.WithError(pubTripErr).Errorf("redis publish trip update")
			return
		}
	}()
	<-done
	return nil
}

func (t *TripRepository) GetTripCourier(courierID uuid.UUID) (*model.Courier, error) {
	courier, err := t.store.GetCourierByID(context.Background(), courierID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		log.WithFields(logrus.Fields{
			"courier_id": courierID,
			"error":      err,
		}).Errorf("get trip courier")
		return nil, err
	}

	return &model.Courier{
		ID:       courier.ID,
		TripID:   &courier.TripID.UUID,
		UserID:   courier.UserID.UUID,
		Location: location.ParsePostgisLocation(courier.Location),
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
		if trip.CourierID.String() == constants.ZERO_UUID {
			go t.publishTripUpdate(tripID, model.TripStatusCancelled, getTripStatusChannel(status))
		}
	default:
		go t.publishTripUpdate(tripID, status, getTripStatusChannel(status))
	}

	return nil
}

func getTripStatusChannel(status model.TripStatus) string {
	switch status {
	case model.TripStatusCourierArriving,
		model.TripStatusCourierEnRoute,
		model.TripStatusComplete:
		return TRIP_UPDATES
	case model.TripStatusCourierAssigned,
		model.TripStatusCancelled:
		return ASSIGN_TRIP
	default:
		return ""
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

func (t *TripRepository) computeRoute(pickup, dropoff location.Geocode) (*model.TripRoute, error) {
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
		log.WithFields(logrus.Fields{
			"route_params": routeParams,
			"error":        payloadErr,
		}).Errorf("marshal route params")
		return nil, payloadErr
	}

	req, reqErr := http.NewRequest("POST", routeV2, bytes.NewBuffer(reqPayload))
	if reqErr != nil {
		log.WithError(reqErr).Errorf("compute route request")
		return nil, reqErr
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Goog-Api-Key", config.Config.GoogleMaps.GoogleRoutesApiKey)
	req.Header.Add(
		"X-Goog-FieldMask",
		"routes.duration,routes.distanceMeters,routes.polyline.encodedPolyline,routes.staticDuration",
	)

	c := &http.Client{}
	res, resErr := c.Do(req)
	if resErr != nil {
		log.WithError(resErr).Errorf("call google compute route api")
		return nil, resErr
	}

	if err := json.NewDecoder(res.Body).Decode(&routeResponse); err != nil {
		log.WithError(err).Errorf("unmarshal google compute route res")
		return nil, err
	}

	if routeResponse.Error.Code > 0 {
		resErr := fmt.Errorf(
			"%s:%v",
			routeResponse.Error.Status,
			routeResponse.Error.Message,
		)
		log.WithFields(logrus.Fields{
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
		log.WithFields(logrus.Fields{
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

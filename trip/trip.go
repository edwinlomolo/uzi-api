package trip

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/edwinlomolo/uzi-api/cache"
	"github.com/edwinlomolo/uzi-api/constants"
	"github.com/edwinlomolo/uzi-api/gql/model"
	l "github.com/edwinlomolo/uzi-api/location"
	"github.com/edwinlomolo/uzi-api/logger"
	"github.com/edwinlomolo/uzi-api/pricer"
	"github.com/edwinlomolo/uzi-api/routing"
	"github.com/edwinlomolo/uzi-api/store"
	sqlStore "github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	TRIP_UPDATES = "trip_updates"
	ASSIGN_TRIP  = "assign_trip"
)

var (
	ErrCourierAlreadyAssigned = errors.New("courier has active trip")
	ErrCourierTripNotFound    = errors.New("courier trip not found")
	location                  l.LocationService
)

type TripService interface {
	FindAvailableCourier(pickup model.GpsInput) (*model.Courier, error)
	GetCourierNearPickupPoint(pickup model.GpsInput) ([]*model.Courier, error)
	AssignCourierToTrip(tripID, courierID uuid.UUID) error
	UnassignTrip(courierID uuid.UUID) error
	CreateTrip(sqlStore.CreateTripParams) (*model.Trip, error)
	CreateTripCost(tripID uuid.UUID, cost int) error
	SetTripStatus(tripID uuid.UUID, status model.TripStatus) error
	MatchCourier(tripID uuid.UUID, pickup model.TripInput)
	CreateTripRecipient(tripID uuid.UUID, input model.TripRecipientInput) error
	GetTripRecipient(tripID uuid.UUID) (*model.Recipient, error)
	GetTrip(tripID uuid.UUID) (*model.Trip, error)
	GetCourierAssignedTrip(courierID uuid.UUID) error
	GetTripCourier(courierID uuid.UUID) (*model.Courier, error)
	GetCourierTrip(tripID uuid.UUID) (*model.Trip, error)
	ReportTripStatus(tripID uuid.UUID, status model.TripStatus) error
}

type tripClient struct {
	redis  *redis.Client
	logger *logrus.Logger
	store  *sqlStore.Queries
	mu     sync.Mutex
}

var Trip TripService

func NewTripService() {
	location = l.Location
	Trip = &tripClient{
		cache.Redis,
		logger.Logger,
		store.DB,
		sync.Mutex{},
	}
	logger.Logger.Infoln("Trip service...OK")
}

func (t *tripClient) FindAvailableCourier(pickup model.GpsInput) (*model.Courier, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	args := sqlStore.FindAvailableCourierParams{
		Point:  fmt.Sprintf("SRID=4326;POINT(%.8f %.8f)", pickup.Lng, pickup.Lat),
		Radius: 2000,
	}
	c, err := t.store.FindAvailableCourier(context.Background(), args)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		uziErr := fmt.Errorf("%s: %v", "available courier", err.Error())
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return &model.Courier{
		ID:        c.ID,
		UserID:    c.UserID.UUID,
		ProductID: c.ProductID.UUID,
		Location:  location.ParsePostgisLocation(c.Location),
	}, nil
}

func (t *tripClient) AssignCourierToTrip(tripID, courierID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	err := t.getCourierAssignedTrip(courierID)
	if err != nil {
		uziErr := fmt.Errorf("%s:%v", "assigned trip", err)
		t.logger.Errorf(uziErr.Error())
		return uziErr
	}

	args := sqlStore.AssignCourierToTripParams{
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
		uziErr := fmt.Errorf("%s:%v", "assign courier", assignCourierErr)
		t.logger.Errorf(uziErr.Error())
		return uziErr
	}

	courierArgs := sqlStore.AssignTripToCourierParams{
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
		uziErr := fmt.Errorf("%s:%v", "assign trip", assignTripErr)
		t.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return nil
}

func (t *tripClient) UnassignTrip(courierID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, err := t.store.UnassignCourierTrip(
		context.Background(),
		courierID,
	); err != nil {
		uziErr := fmt.Errorf("%s:%v", "unassign trip", err)
		t.logger.Errorf(uziErr.Error())
		return nil
	}

	return nil
}

func (t *tripClient) CreateTrip(
	args sqlStore.CreateTripParams,
) (*model.Trip, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	createTrip, err := t.store.CreateTrip(context.Background(), args)
	if err != nil {
		uziErr := fmt.Errorf("%s:%v", "create trip", err)
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return &model.Trip{
		ID:     createTrip.ID,
		Status: model.TripStatus(createTrip.Status),
	}, nil
}

func (t *tripClient) CreateTripCost(tripID uuid.UUID, cost int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	args := sqlStore.CreateTripCostParams{
		ID:   tripID,
		Cost: sql.NullInt32{Int32: int32(cost), Valid: true},
	}
	if _, err := t.store.CreateTripCost(
		context.Background(),
		args,
	); err != nil {
		uziErr := fmt.Errorf("%s:%v", "trip cost", err)
		t.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return nil
}

func (t *tripClient) SetTripStatus(
	tripID uuid.UUID,
	status model.TripStatus,
) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	tripArgs := sqlStore.SetTripStatusParams{
		ID:     tripID,
		Status: status.String(),
	}
	if _, err := t.store.SetTripStatus(
		context.Background(),
		tripArgs); err != nil {
		uziErr := fmt.Errorf("%s:%v", "trip status", err)
		t.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return nil
}

func (t *tripClient) GetCourierNearPickupPoint(
	pickup model.GpsInput,
) ([]*model.Courier, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var couriers []*model.Courier

	args := sqlStore.GetCourierNearPickupPointParams{
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
		uziErr := fmt.Errorf("%s: %v", "near pickup", err.Error())
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
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

func (t *tripClient) getCourierAssignedTrip(courierID uuid.UUID) error {
	_, err := t.store.GetCourierAssignedTrip(
		context.Background(),
		courierID,
	)
	if err == sql.ErrNoRows {
		return nil
	}

	return err
}

func (t *tripClient) GetCourierAssignedTrip(courierID uuid.UUID) error {
	return t.getCourierAssignedTrip(courierID)
}

func (t *tripClient) MatchCourier(tripID uuid.UUID, pickup model.TripInput) {
	pkp, parseErr := routing.Routing.ParsePickupDropoff(pickup)
	if parseErr != nil {
		t.logger.Fatalln(parseErr)
	}

	// use a 5/10/15 minute timeout - trick impatiency cancellation from client(user)
	timeoutCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Minute*10,
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
						t.logger.Errorf(assignErr.Error())
						return
					} else if assignErr != nil {
						t.logger.Errorf(err.Error())
						return
					}
				}
			}
		}
	}()
}

func (t *tripClient) CreateTripRecipient(
	tripID uuid.UUID,
	input model.TripRecipientInput,
) error {
	rArgs := sqlStore.CreateRecipientParams{
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
		uziErr := fmt.Errorf("%s:%v", "new recipient", err)
		t.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return nil
}

func (t *tripClient) GetTripRecipient(tripID uuid.UUID) (*model.Recipient, error) {
	r, err := t.store.GetTripRecipient(
		context.Background(),
		uuid.NullUUID{
			UUID:  tripID,
			Valid: true,
		},
	)
	if err != nil {
		uziErr := fmt.Errorf("%s:%v", "get recipient", err)
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
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

func (t *tripClient) GetTrip(tripID uuid.UUID) (*model.Trip, error) {
	trip, err := t.store.GetTrip(context.Background(), tripID)
	if err != nil {
		uziErr := fmt.Errorf("%s:%v", "get trip", err)
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
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

		switch trp.Status {
		case model.TripStatusCourierArriving,
			model.TripStatusCourierAssigned:
			courierGps, err := t.store.GetCourierLocation(context.Background(), *trp.CourierID)
			if err != nil {
				return nil, err
			}

			pickup.Location = &model.GpsInput{
				Lat: location.ParsePostgisLocation(courierGps).Lat,
				Lng: location.ParsePostgisLocation(courierGps).Lng,
			}
			dropoff.Location = &model.GpsInput{
				Lat: location.ParsePostgisLocation(trip.ConfirmedPickup).Lat,
				Lng: location.ParsePostgisLocation(trip.ConfirmedPickup).Lng,
			}
			tripRoute, err := routing.Routing.ComputeTripRoute(model.TripRouteInput{Pickup: &pickup, Dropoff: &dropoff})
			if err != nil {
				return nil, err
			}
			trp.Route = tripRoute
			cost, costErr := pricer.Pricer.GetTripCost(*trp, trp.Route.Distance)
			if costErr != nil {
				return nil, costErr
			}
			trp.Cost = cost
		case model.TripStatusCourierEnRoute:
			pickup.Location = &model.GpsInput{
				Lat: location.ParsePostgisLocation(trip.ConfirmedPickup).Lat,
				Lng: location.ParsePostgisLocation(trip.ConfirmedPickup).Lng,
			}
			dropoff.Location = &model.GpsInput{
				Lat: trp.EndLocation.Lat,
				Lng: trp.EndLocation.Lng,
			}
			tripRoute, err := routing.Routing.ComputeTripRoute(model.TripRouteInput{Pickup: &pickup, Dropoff: &dropoff})
			if err != nil {
				return nil, err
			}
			trp.Route = tripRoute
			cost, costErr := pricer.Pricer.GetTripCost(*trp, trp.Route.Distance)
			if costErr != nil {
				return nil, costErr
			}
			trp.Cost = cost

			// Create cost while en-route. Is it okay to do this here
			go func() {
				_, err := t.store.CreateTripCost(context.Background(), sqlStore.CreateTripCostParams{
					ID:   tripID,
					Cost: sql.NullInt32{Int32: int32(trp.Cost), Valid: true},
				})
				uziErr := fmt.Errorf("%s:%v", "create trip cost", err)
				t.logger.Errorf("%s:%v", uziErr.Error())
				if err != nil {
					return
				}
			}()
		}
	}

	return trp, nil
}

func (t *tripClient) publishTripUpdate(
	tripID uuid.UUID,
	status model.TripStatus,
	channel string,
) error {
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
			uziErr := fmt.Errorf("%s:%v", "trip update", marshalErr)
			logger.Logger.Errorf(uziErr.Error())
			return
		}

		pubTripErr := t.redis.Publish(context.Background(), channel, u).Err()
		if pubTripErr != nil {
			uziErr := fmt.Errorf("%s:%v", "publish update", pubTripErr)
			logger.Logger.Errorf(uziErr.Error())
			return
		}
	}()
	<-done
	return nil
}

func (t *tripClient) GetTripCourier(courierID uuid.UUID) (*model.Courier, error) {
	courier, err := t.store.GetCourierByID(context.Background(), courierID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		uziErr := fmt.Errorf("%s:%v", "trip courier", err)
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return &model.Courier{
		ID:       courier.ID,
		TripID:   &courier.TripID.UUID,
		UserID:   courier.UserID.UUID,
		Location: location.ParsePostgisLocation(courier.Location),
	}, nil
}

func (t *tripClient) GetCourierTrip(tripID uuid.UUID) (*model.Trip, error) {
	tid := uuid.NullUUID{UUID: tripID, Valid: true}
	trip, err := t.store.GetCourierTrip(context.Background(), tid)
	if err == sql.ErrNoRows {
		t.logger.Errorf(ErrCourierTripNotFound.Error())
		return nil, ErrCourierTripNotFound
	} else if err != nil {
		uziErr := fmt.Errorf("%s:%v", "courier trip", err)
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return &model.Trip{
		ID:     trip.ID,
		Status: model.TripStatus(trip.Status),
	}, nil
}

func (t *tripClient) ReportTripStatus(tripID uuid.UUID, status model.TripStatus) error {
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

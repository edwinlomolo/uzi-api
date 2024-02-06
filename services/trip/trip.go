package trip

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/3dw1nM0535/uzi-api/gql/model"
	"github.com/3dw1nM0535/uzi-api/internal/cache"
	"github.com/3dw1nM0535/uzi-api/internal/logger"
	"github.com/3dw1nM0535/uzi-api/internal/pricer"
	"github.com/3dw1nM0535/uzi-api/store"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	TRIP_UPDATES = "trip_updates"
)

var (
	CourierAlreadyAssigned = errors.New("courier has active trip")
)

type TripService interface {
	FindAvailableCourier(pickup model.GpsInput) (*model.Courier, error)
	GetCourierNearPickupPoint(pickup model.GpsInput) ([]*model.Courier, error)
	AssignCourierToTrip(tripID, courierID uuid.UUID) error
	UnassignTrip(courierID uuid.UUID) error
	CreateTrip(sqlStore.CreateTripParams) (*model.Trip, error)
	CreateTripCost(tripID uuid.UUID, cost int) error
	SetTripStatus(tripID uuid.UUID, status model.TripStatus) error
	GetNearbyAvailableProducts(params sqlStore.GetNearbyAvailableCourierProductsParams, tripDistance int) ([]*model.Product, error)
	MatchCourier(tripID uuid.UUID, pickup model.GpsInput)
}

type point struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type tripClient struct {
	redis  *redis.Client
	logger *logrus.Logger
	store  *sqlStore.Queries
	mu     sync.Mutex
}

var Trip TripService

func NewTripService() {
	Trip = &tripClient{cache.Redis, logger.Logger, store.DB, sync.Mutex{}}
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
		uziErr := fmt.Errorf("%s: %v", "getcouriernearpickuppooint", err.Error())
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return &model.Courier{
		ID:        c.ID,
		ProductID: c.ProductID.UUID,
		Location:  parseCourierLocation(c.Location),
	}, nil
}

func parseCourierLocation(p interface{}) *model.Gps {
	var location *point

	if p != nil {
		json.Unmarshal([]byte((p).(string)), &location)

		lat := &location.Coordinates[1]
		lng := &location.Coordinates[0]
		return &model.Gps{
			Lat: *lat,
			Lng: *lng,
		}
	} else {
		return nil
	}
}

func (t *tripClient) AssignCourierToTrip(tripID, courierID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	trip, err := t.getCourierAssignedTrip(courierID)
	if trip != nil {
		return CourierAlreadyAssigned
	} else if err != nil {
		return err
	}

	args := sqlStore.AssignCourierToTripParams{
		ID:     courierID,
		TripID: uuid.NullUUID{UUID: tripID, Valid: true},
	}
	_, assignCourierErr := t.store.AssignCourierToTrip(context.Background(), args)
	if assignCourierErr != nil {
		uziErr := fmt.Errorf("%s:%v", "assign trip to courier", assignCourierErr)
		t.logger.Errorf(uziErr.Error())
		return uziErr
	}

	courierArgs := sqlStore.AssignTripToCourierParams{
		ID:        tripID,
		CourierID: uuid.NullUUID{UUID: courierID, Valid: true},
	}
	_, assignTripErr := t.store.AssignTripToCourier(context.Background(), courierArgs)
	if assignTripErr != nil {
		uziErr := fmt.Errorf("%s:%v", "assign courier to trip", assignTripErr)
		t.logger.Errorf(uziErr.Error())
		return uziErr
	}

	go func() {
		tripUpdate := model.TripUpdate{ID: tripID}
		tripUpdate.Status = model.TripStatusCourierFound
		u, marshalErr := json.Marshal(tripUpdate)
		if marshalErr != nil {
			logger.Logger.Errorf(marshalErr.Error())
			return
		}

		arrivingErr := t.redis.Publish(context.Background(), TRIP_UPDATES, u).Err()
		if arrivingErr != nil {
			logger.Logger.Errorf(arrivingErr.Error())
			return
		}
	}()

	return nil
}

func (t *tripClient) UnassignTrip(courierID uuid.UUID) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, err := t.store.UnassignCourierTrip(context.Background(), courierID); err != nil {
		uziErr := fmt.Errorf("%s:%v", "unassign trip from courier", err)
		t.logger.Errorf(uziErr.Error())
		return nil
	}

	return nil
}

func (t *tripClient) CreateTrip(args sqlStore.CreateTripParams) (*model.Trip, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	createTrip, err := t.store.CreateTrip(context.Background(), args)
	if err != nil {
		uziErr := fmt.Errorf("%s:%v", "create trip", err)
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return &model.Trip{ID: createTrip.ID}, nil
}

func (t *tripClient) CreateTripCost(tripID uuid.UUID, cost int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	tripCost := strconv.Itoa(cost)

	args := sqlStore.CreateTripCostParams{
		ID:   tripID,
		Cost: sql.NullString{String: tripCost, Valid: true},
	}
	if _, err := t.store.CreateTripCost(context.Background(), args); err != nil {
		uziErr := fmt.Errorf("%s:%v", "set trip cost", err)
		t.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return nil
}

func (t *tripClient) SetTripStatus(tripID uuid.UUID, status model.TripStatus) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	tripArgs := sqlStore.SetTripStatusParams{
		ID:     tripID,
		Status: status.String(),
	}
	if _, err := t.store.SetTripStatus(context.Background(), tripArgs); err != nil {
		uziErr := fmt.Errorf("%s:%v", "set trip status", err)
		t.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return nil
}

func (t *tripClient) GetCourierNearPickupPoint(pickup model.GpsInput) ([]*model.Courier, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var couriers []*model.Courier

	args := sqlStore.GetCourierNearPickupPointParams{
		Point:  fmt.Sprintf("SRID=4326;POINT(%.8f %.8f)", pickup.Lng, pickup.Lat),
		Radius: 2000,
	}
	foundCouriers, err := t.store.GetCourierNearPickupPoint(context.Background(), args)
	if err == sql.ErrNoRows {
		return make([]*model.Courier, 0), nil
	} else if err != nil {
		uziErr := fmt.Errorf("%s: %v", "getcouriernearpickuppooint", err.Error())
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	for _, item := range foundCouriers {
		courier := &model.Courier{
			ID:        item.ID,
			ProductID: item.ProductID.UUID,
			Location:  parseCourierLocation(item.Location),
		}

		couriers = append(couriers, courier)
	}

	return couriers, nil
}

func (t *tripClient) GetNearbyAvailableProducts(params sqlStore.GetNearbyAvailableCourierProductsParams, tripDistance int) ([]*model.Product, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var nearbyProducts []*model.Product

	nearbys, nearbyErr := t.store.GetNearbyAvailableCourierProducts(context.Background(), params)
	if nearbyErr == sql.ErrNoRows {
		return make([]*model.Product, 0), nil
	} else if nearbyErr != nil {
		uziErr := fmt.Errorf("%s:%v", "get nearby available courier products", nearbyErr.Error())
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	for _, item := range nearbys {
		earnWithFuel := item.Name != "UziX"
		product := &model.Product{
			ID:          item.ID_2,
			Price:       pricer.Pricer.CalculateTripCost(int(item.WeightClass), tripDistance, earnWithFuel),
			Name:        item.Name,
			Description: item.Description,
			IconURL:     item.Icon,
		}

		nearbyProducts = append(nearbyProducts, product)
	}

	return nearbyProducts, nil
}

func (t *tripClient) getCourierAssignedTrip(courierID uuid.UUID) (*model.Trip, error) {
	cid := uuid.NullUUID{UUID: courierID, Valid: true}
	trip, err := t.store.GetCourierAssignedTrip(context.Background(), cid)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		uziErr := fmt.Errorf("%s:%v", "get courier assigned trip", err)
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return &model.Trip{ID: trip.ID, Status: model.TripStatus(trip.Status)}, nil
}
func (t *tripClient) MatchCourier(tripID uuid.UUID, pickup model.GpsInput) {
	tripUpdate := model.TripUpdate{ID: tripID}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	go func() {
		defer cancel()

		for {
			select {
			case <-timeoutCtx.Done():
				tripUpdate.Status = model.TripStatusCourierNotFound
				u, marshalErr := json.Marshal(tripUpdate)
				if marshalErr != nil {
					logger.Logger.Errorf(marshalErr.Error())
					return
				}

				notFoundErr := t.redis.Publish(context.Background(), TRIP_UPDATES, u).Err()
				if notFoundErr != nil {
					logger.Logger.Errorf(notFoundErr.Error())
					return
				}

				return
			default:
				courier, err := t.FindAvailableCourier(pickup)
				if err != nil {
					logger.Logger.Errorf(err.Error())
					return
				}

				if courier != nil {
					assignErr := t.AssignCourierToTrip(tripID, courier.ID)
					if assignErr == nil {
						return
					} else if assignErr != nil && errors.Is(assignErr, CourierAlreadyAssigned) {
						continue
					} else if assignErr != nil {
						logger.Logger.Errorf(err.Error())
						return
					}
				}
			}
		}
	}()
}

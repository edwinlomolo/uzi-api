package trip

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/3dw1nM0535/uzi-api/gql/model"
	"github.com/3dw1nM0535/uzi-api/internal/cache"
	"github.com/3dw1nM0535/uzi-api/internal/logger"
	"github.com/3dw1nM0535/uzi-api/internal/route"
	"github.com/3dw1nM0535/uzi-api/store"
	sqlStore "github.com/3dw1nM0535/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type TripService interface {
	ComputeTripRoute(input model.TripRouteInput) (*model.TripRoute, error)
	FindAvailableCourier(pickup model.GpsInput) ([]*model.Courier, error)
	AssignTrip(tripID, courierID uuid.UUID) error
	UnassignTrip(courierID uuid.UUID) error
	CreateTrip(sqlStore.CreateTripParams) (*model.Trip, error)
	CreateTripCost(tripID uuid.UUID, cost int) error
	SetTripStatus(tripID uuid.UUID, status model.TripStatus) error
}

type point struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type tripClient struct {
	redis  *redis.Client
	logger *logrus.Logger
	store  *sqlStore.Queries
	route  route.Route
}

var Trip TripService

func NewTripService() {
	Trip = &tripClient{cache.Redis, logger.Logger, store.DB, route.Routing}
	logger.Logger.Infoln("Trip service...OK")
}

func (t *tripClient) ComputeTripRoute(input model.TripRouteInput) (*model.TripRoute, error) {
	return t.route.ComputeTripRoute(input)
}

func (t *tripClient) FindAvailableCourier(pickup model.GpsInput) ([]*model.Courier, error) {
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
func (t *tripClient) AssignTrip(tripID, courierID uuid.UUID) error {
	args := sqlStore.AssignTripToCourierParams{
		ID:     courierID,
		TripID: uuid.NullUUID{UUID: tripID, Valid: true},
	}
	if _, err := t.store.AssignTripToCourier(context.Background(), args); err != nil {
		uziErr := fmt.Errorf("%s:%v", "assign trip to courier", err)
		t.logger.Errorf(uziErr.Error())
		return uziErr
	}

	return nil
}

func (t *tripClient) UnassignTrip(courierID uuid.UUID) error {
	if _, err := t.store.UnassignTripToCourier(context.Background(), courierID); err != nil {
		uziErr := fmt.Errorf("%s:%v", "unassign trip from courier", err)
		t.logger.Errorf(uziErr.Error())
		return nil
	}

	return nil
}

func (t *tripClient) CreateTrip(args sqlStore.CreateTripParams) (*model.Trip, error) {
	createTrip, err := t.store.CreateTrip(context.Background(), args)
	if err != nil {
		uziErr := fmt.Errorf("%s:%v", "create trip", err)
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	return &model.Trip{ID: createTrip.ID}, nil
}

func (t *tripClient) CreateTripCost(tripID uuid.UUID, cost int) error {
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

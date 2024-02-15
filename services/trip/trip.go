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

	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal/cache"
	"github.com/edwinlomolo/uzi-api/internal/logger"
	"github.com/edwinlomolo/uzi-api/internal/pricer"
	"github.com/edwinlomolo/uzi-api/internal/util"
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
	CreateTripRecipient(tripID uuid.UUID, input model.TripRecipientInput) error
	GetTripRecipient(tripID uuid.UUID) (*model.Recipient, error)
	GetTrip(tripID uuid.UUID) (*model.Trip, error)
	GetCourierAssignedTrip(courierID uuid.UUID) error
	PublishTripUpdate(tripID uuid.UUID, status model.TripStatus, channel string) error
	GetTripCourier(courierID uuid.UUID) (*model.Courier, error)
	GetCourierTrip(courierID uuid.UUID) (*model.Trip, error)
}

type tripClient struct {
	redis  *redis.Client
	logger *logrus.Logger
	store  *sqlStore.Queries
	mu     sync.Mutex
}

var Trip TripService

func NewTripService() {
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
		Location:  util.ParsePostgisLocation(c.Location),
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
		ID:     courierID,
		TripID: uuid.NullUUID{UUID: tripID, Valid: true},
	}
	_, assignCourierErr := t.store.AssignCourierToTrip(
		context.Background(),
		args,
	)
	if assignCourierErr != nil {
		uziErr := fmt.Errorf("%s:%v", "assign trip", assignCourierErr)
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

	tripCost := strconv.Itoa(cost)

	args := sqlStore.CreateTripCostParams{
		ID:   tripID,
		Cost: sql.NullString{String: tripCost, Valid: true},
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
			Location:  util.ParsePostgisLocation(item.Location),
		}

		couriers = append(couriers, courier)
	}

	return couriers, nil
}

func (t *tripClient) GetNearbyAvailableProducts(
	params sqlStore.GetNearbyAvailableCourierProductsParams,
	tripDistance int,
) ([]*model.Product, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var nearbyProducts []*model.Product

	nearbys, nearbyErr := t.store.GetNearbyAvailableCourierProducts(
		context.Background(),
		params,
	)
	if nearbyErr == sql.ErrNoRows {
		return make([]*model.Product, 0), nil
	} else if nearbyErr != nil {
		uziErr := fmt.Errorf("%s:%v", "nearby products", nearbyErr.Error())
		t.logger.Errorf(uziErr.Error())
		return nil, uziErr
	}

	for _, item := range nearbys {
		earnWithFuel := item.Name != "UziX"
		product := &model.Product{
			ID: item.ID_2,
			Price: pricer.Pricer.CalculateTripCost(
				int(item.WeightClass),
				tripDistance,
				earnWithFuel,
			),
			Name:        item.Name,
			Description: item.Description,
			IconURL:     item.Icon,
		}

		nearbyProducts = append(nearbyProducts, product)
	}

	return nearbyProducts, nil
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

func (t *tripClient) MatchCourier(tripID uuid.UUID, pickup model.GpsInput) {
	timeoutCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Minute,
	)

	go func() {
		defer cancel()
		courierFound := false

		for {
			select {
			case <-timeoutCtx.Done():
				if !courierFound {
					go t.PublishTripUpdate(tripID, model.TripStatusCourierNotFound, TRIP_UPDATES)

					done := make(chan struct{})
					go func() {
						defer close(done)
						if err := t.SetTripStatus(tripID, model.TripStatusCourierNotFound); err != nil {
							return
						}
					}()
					<-done
				}

				return
			default:
				time.Sleep(250 * time.Millisecond)

				courier, err := t.FindAvailableCourier(pickup)
				if err != nil {
					return
				}

				if courier != nil && !courierFound {
					courierFound = true
					go t.PublishTripUpdate(tripID, model.TripStatusCourierFound, TRIP_UPDATES)

					done := make(chan struct{})
					go func() {
						defer close(done)
						if err := t.SetTripStatus(tripID, model.TripStatusCourierFound); err != nil {
							return
						}
					}()
					<-done

					assignErr := t.AssignCourierToTrip(tripID, courier.ID)
					if assignErr == nil {
						go t.PublishTripUpdate(tripID, model.TripStatusAssigned, ASSIGN_TRIP)
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
		Name:  input.Name,
		Phone: input.Phone,
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

	return &model.Trip{
		ID:        trip.ID,
		Status:    model.TripStatus(trip.Status),
		CourierID: &trip.CourierID.UUID,
	}, nil
}

func (t *tripClient) PublishTripUpdate(
	tripID uuid.UUID,
	status model.TripStatus,
	channel string,
) error {

	done := make(chan struct{})
	go func() {
		defer close(done)
		update := model.TripUpdate{ID: tripID, Status: status}

		if status == model.TripStatusArriving || status == model.TripStatusEnRoute || status == model.TripStatusAssigned {
			getTrip, err := t.GetTrip(tripID)
			if err != nil {
				return
			}

			tripCourier, courierErr := t.GetTripCourier(*getTrip.CourierID)
			if courierErr != nil {
				return
			}

			switch status {
			case model.TripStatusArriving, model.TripStatusEnRoute:
				update.Location = &model.Gps{Lat: tripCourier.Location.Lat, Lng: tripCourier.Location.Lng}
			case model.TripStatusAssigned:
				update.CourierID = getTrip.CourierID
			}
		}

		u, marshalErr := json.Marshal(update)
		if marshalErr != nil {
			uziErr := fmt.Errorf("%s:%v", "marshal trip update", marshalErr)
			logger.Logger.Errorf(uziErr.Error())
			return
		}

		pubTripErr := t.redis.Publish(context.Background(), channel, u).Err()
		if pubTripErr != nil {
			uziErr := fmt.Errorf("%s:%v", "publish trip update", pubTripErr)
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
		Location: util.ParsePostgisLocation(courier.Location),
	}, nil
}

func (t *tripClient) GetCourierTrip(courierID uuid.UUID) (*model.Trip, error) {
	cid := uuid.NullUUID{UUID: courierID, Valid: true}
	trip, err := t.store.GetCourierTrip(context.Background(), cid)
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

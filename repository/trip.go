package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	ErrCourierAlreadyAssigned = errors.New("trip repository: courier has active trip")
	ErrCourierTripNotFound    = errors.New("trip repository: courier trip not found")
)

type TripRepository struct {
	redis    *redis.Client
	location internal.LocationController
	cache    internal.Cache
	mu       sync.Mutex
	p        internal.Pricing
	store    *sqlc.Queries
	log      *logrus.Logger
}

func (t *TripRepository) Init(q *sqlc.Queries) {
	pr := &PricerRepository{}
	pr.Init(q)
	t.redis = internal.GetCache().GetRedis()
	t.location = internal.GetLocationController()
	t.cache = internal.GetCache()
	t.mu = sync.Mutex{}
	t.p = internal.GetPricer()
	t.store = q
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

func (t *TripRepository) GetTripProduct(productID uuid.UUID) (*model.Product, error) {
	product, err := t.store.GetProductByID(
		context.Background(),
		productID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		t.log.WithFields(logrus.Fields{
			"product_id": productID,
		}).WithError(err).Errorf("get courier product")
		return nil, err
	}

	return &model.Product{
		ID:          product.ID,
		IconURL:     product.Icon,
		Name:        product.Name,
		WeightClass: int(product.WeightClass),
	}, nil
}

func (t *TripRepository) CreateTripCost(tripID uuid.UUID, cost int) error {
	args := sqlc.CreateTripCostParams{
		ID:   tripID,
		Cost: int32(cost),
	}
	if _, err := t.store.CreateTripCost(context.Background(), args); err != nil {
		t.log.WithFields(logrus.Fields{
			"trip_id": tripID,
			"cost":    cost,
		}).WithError(err).Errorf("trip repository: create trip cost")
		return err
	}

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
		}).WithError(err).Errorf("trip status")
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

func (t *TripRepository) GetTrip(tripID uuid.UUID) (*model.Trip, error) {
	trip, err := t.store.GetTrip(context.Background(), tripID)
	if err != nil {
		t.log.WithFields(logrus.Fields{
			"trip_id": tripID,
		}).WithError(err).Errorf("get trip")
		return nil, err
	}

	return &model.Trip{
		ID:              trip.ID,
		Status:          model.TripStatus(trip.Status),
		CourierID:       &trip.CourierID.UUID,
		ProductID:       trip.ProductID,
		Cost:            int(trip.Cost),
		StartLocation:   model.ParsePostgisLocation(trip.StartLocation),
		EndLocation:     model.ParsePostgisLocation(trip.EndLocation),
		ConfirmedPickup: model.ParsePostgisLocation(trip.ConfirmedPickup),
	}, nil
}

func (t *TripRepository) GetCourierLocation(courierID uuid.UUID) (*model.Gps, error) {
	courierGps, err := t.store.GetCourierLocation(context.Background(), courierID)
	if err != nil {
		t.log.WithFields(logrus.Fields{
			"courier_id": courierID,
		}).WithError(err).Errorf("trip repository: get courier location")
		return nil, err
	}

	return model.ParsePostgisLocation(courierGps), nil
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

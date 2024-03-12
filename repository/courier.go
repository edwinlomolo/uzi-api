package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/edwinlomolo/uzi-api/gql/model"
	"github.com/edwinlomolo/uzi-api/internal"
	sqlStore "github.com/edwinlomolo/uzi-api/store"
	"github.com/edwinlomolo/uzi-api/store/sqlc"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type CourierRepository struct {
	redis *redis.Client
	store *sqlc.Queries
	log   *logrus.Logger
}

func (c *CourierRepository) Init() {
	c.redis = internal.GetCache().GetRedis()
	c.log = internal.GetLogger()
	c.store = sqlStore.GetDb()
}

func (c *CourierRepository) FindOrCreate(userID uuid.UUID) (*model.Courier, error) {
	courier, err := c.getCourierByUserID(userID)
	if err == nil && courier == nil {
		newCourier, newErr := c.store.CreateCourier(
			context.Background(),
			uuid.NullUUID{
				UUID:  userID,
				Valid: true,
			},
		)
		if newErr != nil {
			c.log.WithFields(logrus.Fields{
				"courier_user_id": userID,
				"error":           newErr,
			}).Errorf("find/create courier")
			return nil, newErr
		}

		return &model.Courier{
			ID: newCourier.ID,
		}, nil
	} else if err != nil {
		return nil, err
	}

	return &model.Courier{
		ID: courier.ID,
	}, nil
}

func (c *CourierRepository) IsCourier(userID uuid.UUID) (bool, error) {
	isCourier, err := c.store.IsCourier(
		context.Background(),
		uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		c.log.WithFields(logrus.Fields{
			"courier_user_id": userID,
			"error":           err,
		}).Errorf("is courier check")
		return false, err
	}

	return isCourier.Bool, nil
}

func (c *CourierRepository) GetCourierStatus(userID uuid.UUID) (model.CourierStatus, error) {
	status, err := c.store.GetCourierStatus(
		context.Background(),
		uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	)
	if err == sql.ErrNoRows {
		return model.CourierStatusOnboarding, nil
	} else if err != nil {
		c.log.WithFields(logrus.Fields{
			"courier_user_id": userID,
			"error":           err,
		}).Errorf("get courier status")
		return model.CourierStatusOffline, err
	}

	return model.CourierStatus(status), nil
}

func (c *CourierRepository) getCourierByUserID(userID uuid.UUID) (*model.Courier, error) {
	foundCourier, err := c.store.GetCourierByUserID(
		context.Background(),
		uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		c.log.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err,
		}).Errorf("get courier by id")
		return nil, err
	}

	return &model.Courier{
		ID:        foundCourier.ID,
		UserID:    foundCourier.UserID.UUID,
		Avatar:    c.getAvatar(foundCourier.ID),
		ProductID: foundCourier.ProductID.UUID,
		Location:  model.ParsePostgisLocation(foundCourier.Location),
	}, nil
}

func (c *CourierRepository) getAvatar(courierID uuid.UUID) *model.Uploads {
	ID := uuid.NullUUID{
		UUID:  courierID,
		Valid: true,
	}
	avatar, err := c.store.GetCourierAvatar(context.Background(), ID)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		c.log.WithFields(logrus.Fields{
			"courier_id": courierID,
			"error":      err,
		}).Errorf("get courier avatar")
		return nil
	}

	return &model.Uploads{
		ID:  avatar.ID,
		URI: avatar.Uri,
	}
}

func (c *CourierRepository) GetCourierByUserID(userID uuid.UUID) (*model.Courier, error) {
	return c.getCourierByUserID(userID)
}

func (c *CourierRepository) TrackCourierLocation(userID uuid.UUID, input model.GpsInput) error {
	courier, err := c.getCourierByUserID(userID)
	if err != nil {
		return err
	}

	args := sqlc.TrackCourierLocationParams{
		UserID: uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
		Location: fmt.Sprintf(
			"SRID=4326;POINT(%.8f %.8f)",
			input.Lng, input.Lat,
		),
	}
	if _, updateErr := c.store.TrackCourierLocation(
		context.Background(),
		args); updateErr != nil {
		c.log.WithFields(logrus.Fields{
			"courier_user_id": userID,
			"error":           updateErr,
		}).Errorf("track courier location")
		return updateErr
	}

	// Check if courier is on a trip?
	c.isCourierTripping(courier, input)

	return nil
}

func (c *CourierRepository) isCourierTripping(courier *model.Courier, input model.GpsInput) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		t, err := c.GetCourierTrip(courier.ID)
		if err != nil && errors.Is(err, ErrCourierTripNotFound) {
			return
		} else if err != nil {
			return
		}

		if t != nil && (t.Status == model.TripStatusCourierEnRoute || t.Status == model.TripStatusCourierArriving) {
			tripUpdate := model.TripUpdate{
				ID:     t.ID,
				Status: model.TripStatus(t.Status),
				Location: &model.Gps{
					Lat: input.Lat,
					Lng: input.Lng,
				},
			}
			u, marshalErr := json.Marshal(tripUpdate)
			if marshalErr != nil {
				c.log.WithError(marshalErr).Errorf("marshal courier arriving/enroute trip update")
				return
			}
			tripUpdateErr := c.redis.Publish(context.Background(), TRIP_UPDATES, u).Err()
			if tripUpdateErr != nil {
				c.log.WithFields(logrus.Fields{
					"status":  t.Status,
					"trip_id": t.ID,
					"error":   tripUpdateErr,
				}).Errorf("publish courier arriving/enroute trip update")
				return
			}
		}
	}()
	<-done
}

func (c *CourierRepository) GetCourierTrip(tripID uuid.UUID) (*model.Trip, error) {
	tid := uuid.NullUUID{UUID: tripID, Valid: true}
	trip, err := c.store.GetCourierTrip(context.Background(), tid)
	if err == sql.ErrNoRows {
		return nil, ErrCourierTripNotFound
	} else if err != nil {
		c.log.WithFields(logrus.Fields{
			"trip_id": tripID,
			"error":   err,
		}).Errorf("get courier trip")
		return nil, err
	}

	return &model.Trip{
		ID:     trip.ID,
		Status: model.TripStatus(trip.Status),
	}, nil
}

func (c *CourierRepository) UpdateCourierStatus(userID uuid.UUID, status model.CourierStatus) (bool, error) {
	args := sqlc.SetCourierStatusParams{
		Status: status.String(),
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
	}
	if _, setErr := c.store.SetCourierStatus(
		context.Background(),
		args); setErr != nil {
		c.log.WithFields(logrus.Fields{
			"courier_user_id": userID,
			"error":           setErr,
		}).Errorf("update courier status")
		return false, setErr
	}

	return true, nil
}

func (c *CourierRepository) GetCourierProduct(productID uuid.UUID) (*model.Product, error) {
	product, err := c.store.GetCourierProductByID(
		context.Background(),
		productID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		c.log.WithFields(logrus.Fields{
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

func (c *CourierRepository) GetCourierByID(courierID uuid.UUID) (*model.Courier, error) {
	courier, err := c.store.GetCourierByID(
		context.Background(),
		courierID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		c.log.WithFields(logrus.Fields{
			"courier_id": courierID,
			"error":      err,
		}).Errorf("get courier by id")
		return nil, err
	}

	return &model.Courier{
		ID:        courier.ID,
		TripID:    &courier.TripID.UUID,
		UserID:    courier.UserID.UUID,
		ProductID: courier.ProductID.UUID,
		Location:  model.ParsePostgisLocation(courier.Location),
		Avatar:    c.getAvatar(courierID),
	}, nil
}
